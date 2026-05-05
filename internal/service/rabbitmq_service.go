/*
 * @Date: 2026-05-04 16:46:23
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-05-04 17:05:28
 * @Description: 消息队列服务
 */
package service

import (
	"encoding/json"
	"go-blog/internal/common"
	"go-blog/internal/config"
	"go-blog/internal/model"
	"go-blog/internal/rabbitmq"
	"go-blog/internal/repository"
)

/**
 * @description: 消费RabbitMQ消息
 * @param {func([]byte) error} handler
 * @return {error}
 */
func ConsumeRabbitMQ(handler func([]byte) error) {
	msgs, err := rabbitmq.ConsumeMessage(config.Conf.Rabbitmq.QueueName, "")
	if err != nil {
		common.Log.Errorf("消费者注册失败: %v", err)
	}

	forever := make(chan struct{})
	go func() {
		for d := range msgs {
			// 调用传入的处理函数
			if err := handler(d.Body); err != nil {
				// 处理消息失败，可选择重新入队或记录死信
				common.Log.Errorf("处理消息失败: %v", err)
				// 可选：nack 并重新入队，这里简单丢弃
				d.Nack(false, false)
			} else {
				d.Ack(false) // 确认成功消费
			}
		}
	}()
	common.Log.Info("RabbitMQ 消费者已启动，等待消息...")
	<-forever
}

/**
 * @description: 处理博客消息 是 RabbitMQ 消费者调用的回调函数，负责将消息中的文章写入 ElasticSearch 和 ClickHouse
 * @param {[]byte} data
 * @return {error}
 */
func HandleArticleMessage(data []byte) error {
	var blog model.Blog
	if err := json.Unmarshal(data, &blog); err != nil {
		common.Log.Errorf("反序列化文章失败: %v", err)
		return err
	}

	// 写入 ElasticSearch
	if err := repository.IndexBlogToES(&blog); err != nil {
		common.Log.Errorf("写入 ES 失败: %v", err)
		return err
	}

	// 写入 ClickHouse
	if err := repository.InsertBlogToClickHouse(&blog); err != nil {
		common.Log.Errorf("写入 ClickHouse 失败: %v", err)
		return err
	}
	common.Log.Infof("文章 %d 已成功写入 ES 和 ClickHouse", blog.ID)
	return nil
}
