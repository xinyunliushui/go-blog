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

	"github.com/streadway/amqp"
)

// MessageHandler 带 trace 上下文的 MQ 消息处理函数。
type MessageHandler func(ctx context.Context, body []byte) error

/***
 * @description: 消费 RabbitMQ 消息
 * @param ctx context.Context 上下文
 * @param handler MessageHandler 消息处理函数
 */
func ConsumeRabbitMQ(ctx context.Context, handler MessageHandler) {
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
			msgCtx := messageContext(ctx, d)
			logger := common.LoggerFromCtx(msgCtx)
			// 处理消息
			if err := handler(msgCtx, d.Body); err != nil {
				if ack := compensateConsumeFailure(compRepo, msgCtx, d.Body, err); ack {
					// 补偿落库，告诉 RabbitMQ「这条 MQ 消息可以结束了」，不再重投
					if ackErr := d.Ack(false); ackErr != nil {
						logger.Errorf("[消息消费] Ack 失败: %v", ackErr)
					}
				} else {
					logger.Errorf("[消息消费] 处理失败且补偿落库失败，转入死信队列: %v", err)
					// Nack(requeue=false)：由 go_blog_queue 的 x-dead-letter-* 路由至 go_blog_dlq
					if nackErr := d.Nack(false, false); nackErr != nil {
						logger.Errorf("[消息消费] Nack 失败: %v", nackErr)
					}
				}
				continue
			}
			// 消息处理成功，告诉 RabbitMQ「这条 MQ 消息可以结束了」，不再重投
			if ackErr := d.Ack(false); ackErr != nil {
				logger.Errorf("[消息消费] Ack 失败: %v", ackErr)
			}
		}
	}
}

func messageContext(parent context.Context, d amqp.Delivery) context.Context {
	traceID := common.TraceIDFromAMQPHeaders(d.Headers)
	if traceID == "" {
		traceID = common.NewTraceID()
	}
	return common.ContextWithTraceID(parent, traceID)
}

/**
 * @description: 补偿-消费失败-落库
 * @param repo repository.IMQCompensationRepository 补偿仓库
 * @param ctx context.Context 上下文
 * @param body []byte 消息体
 * @param err error 错误
 * @return bool 是否补偿成功
 */
func compensateConsumeFailure(repo repository.IMQCompensationRepository, ctx context.Context, body []byte, err error) bool {
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
	logger := common.LoggerFromCtx(ctx)
	logger.Errorf("[消息消费] 消费失败，写入 CONSUME 补偿 mask=%s err=%v",
		FormatSyncMask(syncErr.PendingMask), syncErr.Cause)
	traceID := common.TraceIDFromCtx(ctx)
	if encErr := enqueueBlogConsumeCompensation(repo, traceID, syncErr); encErr != nil {
		logger.Errorf("[消息消费] CONSUME 补偿表写入失败: %v", encErr)
		return false
	}
	return true
}

/**
 * @description: 处理文章消息
 * @param ctx context.Context 上下文
 * @param data []byte 消息体
 * @return error 错误
 */
func HandleArticleMessage(ctx context.Context, data []byte) error {
	logger := common.LoggerFromCtx(ctx)
	var blog model.Blog
	if err := json.Unmarshal(data, &blog); err != nil {
		logger.Errorf("[消息消费] 反序列化文章失败: %v", err)
		return &BlogSyncConsistentError{Body: data, PendingMask: model.SyncPendingAll, Cause: err}
	}

	if err := ReconcileAndSyncBlog(&blog, model.SyncPendingAll); err != nil {
		logger.Errorf("[消息消费] 同步检索库失败 blog_id=%s mask=%s: %v",
			blog.ID, FormatSyncMask(SyncErrorPendingMask(err)), err)
		return err
	}
	logger.Infof("[消息消费] 文章 %s 已成功同步至 ES 和 ClickHouse", blog.ID)
	return nil
}
