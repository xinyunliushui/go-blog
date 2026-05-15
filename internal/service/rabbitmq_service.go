/*
 * @Date: 2026-05-04 16:46:23
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-05-13 15:46:41
 * @Description: 消息队列服务
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
)

/** 消费 RabbitMQ 消息 ctx 取消后退出循环并返回，便于主流程 WaitGroup 等待。
 * @param ctx context.Context 上下文
 * @param handler func([]byte) error 处理消息的函数
 * @return void
 */
func ConsumeRabbitMQ(ctx context.Context, handler func([]byte) error) {
	msgs, err := rabbitmq.ConsumeMessage(config.Conf.Rabbitmq.QueueName, "")
	if err != nil {
		common.Log.Errorf("[消息消费] 消费者注册失败: %v", err)
		return
	}

	common.Log.Info("[消息消费] 注册成功，RabbitMQ 消费者已启动，等待消息...")
	for {
		select {
		case <-ctx.Done():
			common.Log.Info("[消息消费] RabbitMQ 消费者已退出")
			return
		case d, ok := <-msgs:
			// 每收到 一条 RabbitMQ 投递，就往这个 channel 里放进 一个 Delivery
			if !ok {
				common.Log.Info("[消息消费] RabbitMQ 消费 channel 已关闭，消费者退出")
				return
			}
			if err := handler(d.Body); err != nil {
				// 处理消息失败，可选择重新入队或记录死信
				// 可选：nack 并重新入队，这里简单丢弃
				common.Log.Errorf("[消息消费] 处理消息失败: %v", err)
				d.Nack(false, false)
			} else {
				// 确认成功消费
				d.Ack(false)
			}
		}
	}
}

/**
 * @description: 处理博客消息 是 RabbitMQ 消费者调用的回调函数，负责将消息中的文章写入 ElasticSearch 和 ClickHouse
 * @param {[]byte} data
 * @return {error}
 */
func HandleArticleMessage(data []byte) error {
	var blog model.Blog
	if err := json.Unmarshal(data, &blog); err != nil {
		common.Log.Errorf("[消息消费] 反序列化文章失败: %v", err)
		return err
	}

	// 写入 ElasticSearch
	if err := repository.IndexBlogToES(&blog); err != nil {
		common.Log.Errorf("[消息消费] 写入 ES 失败: %v", err)
		return err
	}

	// 写入 ClickHouse
	if err := repository.InsertBlogToClickHouse(&blog); err != nil {
		common.Log.Errorf("[消息消费] 写入 ClickHouse 失败: %v", err)
		return err
	}
	common.Log.Infof("[消息消费] 文章 %s 已成功写入 ES 和 ClickHouse", blog.ID)
	return nil
}
