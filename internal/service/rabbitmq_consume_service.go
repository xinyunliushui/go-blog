/*
 * @Date: 2026-05-08 16:47:20
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-05-18 22:16:21
 * @Description: 消息队列消费与补偿处理
 */

package service

import (
	"context"
	"encoding/json"
	"errors"
	"go-blog/internal/common"
	"go-blog/internal/config"
	"go-blog/internal/model"
	"go-blog/internal/rabbitmq"
	"go-blog/internal/repository"
)

/***
 * @description: 消费 RabbitMQ 消息
 * @param ctx context.Context 上下文
 * @param handler func([]byte) error 消息处理函数
 */
func ConsumeRabbitMQ(ctx context.Context, handler func([]byte) error) {
	msgs, err := rabbitmq.ConsumeMessage(config.Conf.Rabbitmq.QueueName, "")
	if err != nil {
		common.Log.Errorf("[消息消费] 消费者注册失败: %v", err)
		return
	}

	compRepo := repository.NewMQCompensationRepository()
	common.Log.Info("[消息消费] 注册成功，RabbitMQ 消费者已启动，等待消息...")
	for {
		select {
		case <-ctx.Done():
			common.Log.Info("[消息消费] RabbitMQ 消费者已退出")
			return
		case d, ok := <-msgs:
			if !ok {
				common.Log.Info("[消息消费] RabbitMQ 消费 channel 已关闭，消费者退出")
				return
			}
			// 处理消息
			if err := handler(d.Body); err != nil {
				if ack := compensateConsumeFailure(compRepo, d.Body, err); ack {
					// 补偿落库，告诉 RabbitMQ「这条 MQ 消息可以结束了」，不再重投
					// func (d Delivery) Ack(multiple bool) error
					if ackErr := d.Ack(false); ackErr != nil {
						common.Log.Errorf("[消息消费] Ack 失败: %v", ackErr)
					}
				} else {
					common.Log.Errorf("[消息消费] 处理失败且补偿落库失败，Nack 丢弃: %v", err)
					// func (d Delivery) Nack(multiple bool, requeue bool) error
					// 不重投, 丢弃 (MQ可配置死信队列来处理)
					_ = d.Nack(false, false)
				}
				continue
			}
			// 消息处理成功，告诉 RabbitMQ「这条 MQ 消息可以结束了」，不再重投
			if ackErr := d.Ack(false); ackErr != nil {
				common.Log.Errorf("[消息消费] Ack 失败: %v", ackErr)
			}
		}
	}
}

/**
 * @description: 补偿-消费失败-落库
 * @param repo repository.IMQCompensationRepository 补偿仓库
 * @param body []byte 消息体
 * @param err error 错误
 * @return bool 是否补偿成功
 */
func compensateConsumeFailure(repo repository.IMQCompensationRepository, body []byte, err error) bool {
	var syncErr *BlogSyncConsistentError
	if !errors.As(err, &syncErr) {
		syncErr = &BlogSyncConsistentError{Body: body, PendingMask: model.SyncPendingAll, Cause: err}
	}
	if len(syncErr.Body) == 0 {
		syncErr.Body = body
	}
	if syncErr.PendingMask == 0 {
		syncErr.PendingMask = model.SyncPendingAll
	}
	common.Log.Errorf("[消息消费] 消费失败，写入 CONSUME 补偿 mask=%s err=%v",
		FormatSyncMask(syncErr.PendingMask), syncErr.Cause)
	if encErr := enqueueBlogConsumeCompensation(repo, syncErr); encErr != nil {
		common.Log.Errorf("[消息消费] CONSUME 补偿表写入失败: %v", encErr)
		return false
	}
	return true
}

/**
 * @description: 处理文章消息
 * @param data []byte 消息体
 * @return error 错误
 */
func HandleArticleMessage(data []byte) error {
	var blog model.Blog
	if err := json.Unmarshal(data, &blog); err != nil {
		common.Log.Errorf("[消息消费] 反序列化文章失败: %v", err)
		return &BlogSyncConsistentError{Body: data, PendingMask: model.SyncPendingAll, Cause: err}
	}

	if err := ReconcileAndSyncBlog(&blog, model.SyncPendingAll); err != nil {
		common.Log.Errorf("[消息消费] 同步检索库失败 blog_id=%s mask=%s: %v",
			blog.ID, FormatSyncMask(SyncErrorPendingMask(err)), err)
		return err
	}
	common.Log.Infof("[消息消费] 文章 %s 已成功同步至 ES 和 ClickHouse", blog.ID)
	return nil
}
