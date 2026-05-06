/*
 * @Date: 2026-04-27 13:41:48
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-05-04 19:32:19
 * @Description: rabbitmq
 */
package rabbitmq

import (
	"encoding/json"
	"fmt"
	"go-blog/internal/common"
	"go-blog/internal/config"
	"sync"

	"github.com/streadway/amqp"
)

var (
	connection *amqp.Connection
	channel    *amqp.Channel
	once       sync.Once
)

/**
 * @description: 初始化RabbitMQ连接
 * @return {error}
 */
func InitRabbitMQ() error {
	var err error
	once.Do(func() {
		rabbitURL := fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
			config.Conf.Rabbitmq.Username,
			config.Conf.Rabbitmq.Password,
			config.Conf.Rabbitmq.Host,
			config.Conf.Rabbitmq.Port,
			config.Conf.Rabbitmq.VHost,
		)
		connection, err = amqp.Dial(rabbitURL)
		if err != nil {
			common.Log.Errorf("连接RabbitMQ失败: %s", err)
			return
		}

		channel, err = connection.Channel()
		if err != nil {
			common.Log.Errorf("创建通道失败: %s", err)
			return
		}

		// 声明队列（确保队列存在）
		_, err = channel.QueueDeclare(
			config.Conf.Rabbitmq.QueueName, // 队列名称
			true,                           // 是否持久化
			false,                          // 不使用时是否自动删除
			false,                          // 是否具有排他性
			false,                          // 是否阻塞等待服务器处理
			nil,                            // 额外属性
		)
		if err != nil {
			common.Log.Errorf("声明队列失败: %s", err)
			return
		}
		common.Log.Info("初始化RabbitMQ连接成功")
	})
	return err
}

/**
 * @description: 发布消息
 * @param {string} queueName 队列名称
 * @param {[]byte} message 消息
 * @return {error}
 */
func PublishMessage(queueName string, message interface{}) error {
	// 如果通道为空，则初始化RabbitMQ连接
	if channel == nil {
		if err := InitRabbitMQ(); err != nil {
			common.Log.Errorf("初始化RabbitMQ连接失败: %s", err)
			return err
		}
	}
	// 序列化消息
	body, err := json.Marshal(message)
	if err != nil {
		common.Log.Errorf("序列化消息失败: %s", err)
		return err
	}
	// 发布消息
	err = channel.Publish(
		"",        // 交换机名称（空字符串表示使用默认交换机）
		queueName, // 路由键（这里直接填队列名）
		false,     // 如果为true，当交换机找不到队列时会返回错误
		false,     // 如果为true，当队列没有消费者时会返回错误
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent, // 消息持久化
		},
	)
	return err
}

/**
 * @description: 消费消息-默认需要手动回执成功才能收到下一条消息
 * @param {string} queueName 队列名称
 * @param {string} consumerName 消费者名称
 * @return {<-chan amqp.Delivery, error}
 */
func ConsumeMessage(queueName string, consumerName string) (<-chan amqp.Delivery, error) {
	// 如果通道为空，则初始化RabbitMQ连接
	if channel == nil {
		if err := InitRabbitMQ(); err != nil {
			common.Log.Errorf("初始化RabbitMQ连接失败: %s", err)
			return nil, err
		}
	}
	// 消费消息
	msgs, err := channel.Consume(
		queueName,    // 队列名称
		consumerName, // 消费者标签
		false,        // 是否自动确认（ACK）
		false,        // 是否排他
		false,        // 是否本地
		false,        // 是否阻塞等待
		nil,          // 额外参数
	)
	return msgs, err
}

/**
 * @description: 关闭RabbitMQ连接 释放资源
 */
func CloseRabbitMQ() {
	if channel != nil {
		channel.Close()
	}
	if connection != nil {
		connection.Close()
	}
}
