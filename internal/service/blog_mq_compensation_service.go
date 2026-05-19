/*
 * @Date: 2026-05-18 15:42:40
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-05-18 15:58:22
 * @Description: 博客 MQ 补偿统一重试（PUBLISH + CONSUME）
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
	blogCompensationRetryInterval = 30 * time.Second // 补偿重试间隔
	blogCompensationMaxRetries    = 5                // 补偿重试最大次数
	blogCompensationBatchSize     = 10               // 补偿重试批次大小
)

/**
 * @description: 运行博客 MQ 补偿统一重试
 * @param ctx context.Context 上下文
 */
func RunBlogMQCompensationRetry(ctx context.Context) {
	// 创建补偿批处理定时器
	ticker := time.NewTicker(blogCompensationRetryInterval)
	defer ticker.Stop()
	repo := repository.NewMQCompensationRepository()
	// 循环处理补偿任务
	for {
		// 监听上下文取消信号，用于优雅关闭
		select {
		case <-ctx.Done():
			common.Log.Info("[MQ补偿] 统一重试任务已退出")
			return
		case <-ticker.C:
			// 处理补偿任务批次
			processBlogCompensationBatch(repo)
		}
	}
}

/**
 * @description: 处理补偿任务批次
 * @param repo repository.IMQCompensationRepository 补偿仓库
 */
func processBlogCompensationBatch(repo repository.IMQCompensationRepository) {
	rows, err := repo.ListPendingForRetry(blogCompensationBatchSize)
	if err != nil {
		common.Log.Errorf("[MQ补偿] 查询待重试记录失败: %v", err)
		return
	}
	for i := range rows {
		row := &rows[i]
		taskType := row.TaskType
		// 如果任务类型为空，则默认为发布任务类型
		if taskType == "" {
			taskType = model.TaskTypePublish
		}
		switch taskType {
		case model.TaskTypeConsume:
			// 处理消费补偿
			processConsumeCompensation(repo, row)
		default:
			// 处理发布补偿
			processPublishCompensation(repo, row)
		}
	}
}

/***
 * @description: 处理发布补偿
 * @param repo repository.IMQCompensationRepository 补偿仓库
 * @param row *model.BlogMQCompensation 补偿记录
 */
func processPublishCompensation(repo repository.IMQCompensationRepository, row *model.BlogMQCompensation) {
	var blog model.Blog
	if err := json.Unmarshal(row.Payload, &blog); err != nil {
		handleCompensationRetry(repo, row, 0, err, "[MQ补偿-推送]")
		return
	}

	queue := config.Conf.Rabbitmq.QueueName
	// 重新投递 MQ
	if err := rabbitmq.PublishMessage(queue, &blog); err != nil {
		common.Log.Errorf("[MQ补偿-推送] id=%s blog_id=%s 第%d次重试仍失败: %v",
			row.ID, row.BlogID, row.RetryCount+1, err)
		handleCompensationRetry(repo, row, 0, err, "[MQ补偿-推送]")
		return
	}

	if err := repo.MarkSent(row.ID); err != nil {
		common.Log.Errorf("[MQ补偿-推送] 标记成功失败 id=%s: %v", row.ID, err)
		return
	}
	common.Log.Infof("[MQ补偿-推送] 重试成功 id=%s blog_id=%s", row.ID, row.BlogID)
}

/**
 * @description: 处理消费补偿
 * @param repo repository.IMQCompensationRepository 补偿仓库
 * @param row *model.BlogMQCompensation 补偿记录
 */
func processConsumeCompensation(repo repository.IMQCompensationRepository, row *model.BlogMQCompensation) {
	var blog model.Blog
	if err := json.Unmarshal(row.Payload, &blog); err != nil {
		handleCompensationRetry(repo, row, row.EffectivePendingMask(), err, "[MQ补偿-消费]")
		return
	}

	mask := row.EffectivePendingMask()
	if mask == 0 {
		mask = model.SyncPendingAll
	}
	before := mask

	if err := ReconcileAndSyncBlog(&blog, mask); err != nil {
		nextMask := SyncErrorPendingMask(err)
		if nextMask == 0 {
			nextMask = mask
		}
		LogSyncReconcile(blog.ID, before, nextMask)
		common.Log.Errorf("[MQ补偿-消费] id=%s blog_id=%s mask=%s 第%d次失败: %v next=%s",
			row.ID, row.BlogID, FormatSyncMask(mask), row.RetryCount+1, err, FormatSyncMask(nextMask))
		handleCompensationRetry(repo, row, nextMask, err, "[MQ补偿-消费]")
		return
	}

	if err := repo.MarkSent(row.ID); err != nil {
		common.Log.Errorf("[MQ补偿-消费] 标记成功失败 id=%s: %v", row.ID, err)
		return
	}
	common.Log.Infof("[MQ补偿-消费] 重试成功 id=%s blog_id=%s", row.ID, row.BlogID)
}

/**
 * @description: 补偿相关状态位变更
 * @param repo repository.IMQCompensationRepository 补偿仓库
 * @param row *model.BlogMQCompensation 补偿记录
 * @param pendingMask uint8 待同步位图
 * @param cause error 错误
 * @param logPrefix string 日志前缀
 */
func handleCompensationRetry(repo repository.IMQCompensationRepository, row *model.BlogMQCompensation, pendingMask uint8, cause error, logPrefix string) {
	next := row.RetryCount + 1
	dead := next >= blogCompensationMaxRetries
	errMsg := cause.Error()

	if err := repo.MarkRetry(row.ID, next, errMsg, dead, pendingMask); err != nil {
		common.Log.Errorf("%s 更新补偿记录失败 id=%s: %v", logPrefix, row.ID, err)
		return
	}
	if dead {
		if row.TaskType == model.TaskTypeConsume {
			common.Log.Errorf("%s [MQ_CONSUME_DEAD] id=%s blog_id=%s mask=%s 请人工核对 ES/CH",
				logPrefix, row.ID, row.BlogID, FormatSyncMask(pendingMask))
		} else {
			common.Log.Errorf("%s [MQ_PUBLISH_DEAD] id=%s blog_id=%s 请人工补投 MQ", logPrefix, row.ID, row.BlogID)
		}
	}
}
