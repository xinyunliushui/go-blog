/*
 * @Date: 2026-05-19 13:56:46
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-05-19 13:58:27
 * @Description: 死信队列消费：主消费失败且补偿落库失败时，由 Broker 转入 DLQ，此处再次尝试写入补偿表
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

const dlqConsumerTag = "go-blog-dlq-consumer"

var errDLQPendingCompensation = errors.New("主消费失败且补偿落库失败，消息已由死信队列承接")

/**
 * @description: 消费死信队列消息
 * @param ctx context.Context 上下文
 */
func ConsumeRabbitMQDLQ(ctx context.Context) {
	msgs, err := rabbitmq.ConsumeMessage(config.Conf.Rabbitmq.DLQName, dlqConsumerTag)
	if err != nil {
		common.Log.Errorf("[DLQ消费] 消费者注册失败: %v", err)
		return
	}

	compRepo := repository.NewMQCompensationRepository()
	common.Log.Info("[DLQ消费] 注册成功，死信队列消费者已启动，等待消息...")
	for {
		select {
		case <-ctx.Done():
			common.Log.Info("[DLQ消费] 死信队列消费者已退出")
			return
		case d, ok := <-msgs:
			if !ok {
				common.Log.Info("[DLQ消费] 消费 channel 已关闭，消费者退出")
				return
			}
			msgCtx := messageContext(ctx, d)
			logger := common.LoggerFromCtx(msgCtx)
			if handleDLQMessage(compRepo, msgCtx, d.Body) {
				if ackErr := d.Ack(false); ackErr != nil {
					logger.Errorf("[DLQ消费] Ack 失败: %v", ackErr)
				}
				continue
			}
			logger.Errorf("[DLQ消费] [MQ_DLQ_DEAD] 再次落补偿表仍失败，请人工核对 ES/CH body_len=%d",
				len(d.Body))
			// 必须 Ack，避免在 DLQ 内 Nack 造成死循环
			if ackErr := d.Ack(false); ackErr != nil {
				logger.Errorf("[DLQ消费] Ack 失败: %v", ackErr)
			}
		}
	}
}

/**
 * @description: 处理死信消息，仅尝试写入 CONSUME 补偿表，由定时任务负责 ES/CH 同步
 * @param repo repository.IMQCompensationRepository 补偿仓库
 * @param ctx context.Context 上下文
 * @param body []byte 消息体
 * @return bool 是否处理成功
 */
func handleDLQMessage(repo repository.IMQCompensationRepository, ctx context.Context, body []byte) bool {
	syncErr := &BlogSyncConsistentError{
		Body:        body,
		PendingMask: model.SyncPendingAll,
		Cause:       errDLQPendingCompensation,
	}
	var blog model.Blog
	if err := json.Unmarshal(body, &blog); err == nil && blog.ID != "" {
		syncErr.Blog = &blog
	}
	common.LoggerFromCtx(ctx).Warnf("[DLQ消费] 尝试再次写入 CONSUME 补偿 blog_id=%s", blog.ID)
	return compensateConsumeFailure(repo, ctx, body, syncErr)
}
