/*
 * @Date: 2026-05-07 17:04:50
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-05-07 17:24:45
 * @Description: 博客MQ投递失败时落库并定时重试
 */
package service

import (
	"context"
	"encoding/json"
	"go-blog/internal/common"
	"go-blog/internal/config"
	"go-blog/internal/model"
	"go-blog/internal/rabbitmq"
	"go-blog/internal/repository"
	"time"
)

const (
	blogOutboxRetryInterval = 30 * time.Second // 30秒重试一次
	blogOutboxMaxRetries    = 15               // 15次重试
	blogOutboxBatchSize     = 50               // 50条记录一批
)

/**
 * @description: 博客MQ投递失败时落库并定时重试
 * @param {context.Context} ctx
 * @return {void}
 */
func RunBlogMQOutboxRetry(ctx context.Context) {
	ticker := time.NewTicker(blogOutboxRetryInterval)
	defer ticker.Stop()
	repo := repository.NewMQOutboxRepository()

	for {
		select {
		case <-ctx.Done():
			common.Log.Info("Blog MQ Outbox 重试任务已退出")
			return
		case <-ticker.C:
			processBlogOutboxBatch(repo)
		}
	}
}

/**
 * @description: 处理博客MQ投递失败时落库并定时重试
 * @param {repository.IMQOutboxRepository} repo
 * @return {void}
 */
func processBlogOutboxBatch(repo repository.IMQOutboxRepository) {
	rows, err := repo.ListPendingForRetry(blogOutboxBatchSize)
	if err != nil {
		common.Log.Errorf("[MQ_OUTBOX_ALERT] 查询待重试 Outbox 失败: %v", err)
		return
	}
	if len(rows) == 0 {
		return
	}

	queue := config.Conf.Rabbitmq.QueueName
	for _, row := range rows {
		// 反序列化 payload 为 blog
		var blog model.Blog
		if err := json.Unmarshal(row.Payload, &blog); err != nil {
			common.Log.Errorf("[MQ_OUTBOX_ALERT] outbox_id=%d blog_id=%d payload 解析失败: %v", row.ID, row.BlogID, err)
			next := row.RetryCount + 1
			dead := next >= blogOutboxMaxRetries
			if err2 := repo.MarkRetry(row.ID, next, err.Error(), dead); err2 != nil {
				common.Log.Errorf("[MQ_OUTBOX_ALERT] 更新 Outbox 失败 outbox_id=%d: %v", row.ID, err2)
			}
			if dead {
				common.Log.Errorf("[MQ_OUTBOX_DEAD_LETTER] outbox_id=%d blog_id=%d payload 损坏且已达最大重试", row.ID, row.BlogID)
			}
			continue
		}

		if err := rabbitmq.PublishMessage(queue, &blog); err != nil {
			next := row.RetryCount + 1
			dead := next >= blogOutboxMaxRetries
			common.Log.Errorf("[MQ_OUTBOX_ALERT] outbox_id=%d blog_id=%d 第%d次重试仍失败: %v dead=%v",
				row.ID, row.BlogID, next, err, dead)
			if err2 := repo.MarkRetry(row.ID, next, err.Error(), dead); err2 != nil {
				common.Log.Errorf("[MQ_OUTBOX_ALERT] 更新 Outbox 失败 outbox_id=%d: %v", row.ID, err2)
				continue
			}
			if dead {
				common.Log.Errorf("[MQ_OUTBOX_DEAD_LETTER] outbox_id=%d blog_id=%d 已达最大重试次数，请人工补投或修 MQ", row.ID, row.BlogID)
			}
			continue
		}

		if err := repo.MarkSent(row.ID); err != nil {
			common.Log.Errorf("[MQ_OUTBOX_ALERT] 标记 Outbox 成功失败 outbox_id=%d blog_id=%d: %v", row.ID, row.BlogID, err)
			continue
		}
		common.Log.Infof("Blog MQ Outbox 重试成功 outbox_id=%d blog_id=%d", row.ID, row.BlogID)
	}
}
