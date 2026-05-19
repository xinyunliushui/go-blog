/*
 * @Date: 2026-04-27 13:41:48
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-05-19 15:04:19
 * @Description: rabbitmq
 */
package rabbitmq

import (
	"encoding/json"
	"errors"
	"fmt"
	"go-blog/internal/common"
	"go-blog/internal/config"
	"strings"
	"sync"

	"github.com/streadway/amqp"
)

var (
	connection *amqp.Connection
	publishCh  *amqp.Channel // 专用于发布，不与消费、探针共用
	mu         sync.Mutex
)

/**
 * @description: 初始化RabbitMQ连接
 * @return {error}
 */
func InitRabbitMQ() error {
	mu.Lock()
	defer mu.Unlock()

	if isReadyLocked() {
		return nil
	}
	resetLocked()

	rabbitURL := fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
		config.Conf.Rabbitmq.Username,
		config.Conf.Rabbitmq.Password,
		config.Conf.Rabbitmq.Host,
		config.Conf.Rabbitmq.Port,
		config.Conf.Rabbitmq.VHost,
	)
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		common.Log.Errorf("连接RabbitMQ失败: %s", err)
		return err
	}

	// 声明队列使用独立 channel，避免与消费、探针争用同一 channel（AMQP channel 非线程安全）
	topoCh, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		common.Log.Errorf("创建队列通道失败: %s", err)
		return err
	}
	if err = declareTopology(topoCh); err != nil {
		_ = topoCh.Close()
		_ = conn.Close()
		common.Log.Errorf("声明 RabbitMQ 队列失败: %s", err)
		return err
	}
	_ = topoCh.Close()

	pubCh, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		common.Log.Errorf("创建发布通道失败: %s", err)
		return err
	}

	connection = conn
	publishCh = pubCh
	common.Log.Info("初始化RabbitMQ连接成功")
	return nil
}

// 检查连接是否就绪
func isReadyLocked() bool {
	return connection != nil && !connection.IsClosed() && publishCh != nil
}

// 重置连接
func resetLocked() {
	if publishCh != nil {
		_ = publishCh.Close()
		publishCh = nil
	}
	if connection != nil {
		_ = connection.Close()
		connection = nil
	}
}

/**
 * @description: 确保连接已就绪
 * @return {error}
 */
func ensureConnected() error {
	mu.Lock()
	ready := isReadyLocked()
	mu.Unlock()
	if ready {
		return nil
	}
	return InitRabbitMQ()
}

/**
 * @description: 声明交换机、业务队列、死信交换机与死信队列及绑定关系
 */
func declareTopology(ch *amqp.Channel) error {
	cfg := config.Conf.Rabbitmq
	exchangeType := cfg.ExchangeType
	if exchangeType == "" {
		exchangeType = "direct"
	}

	// 死信交换机
	if err := ch.ExchangeDeclare(
		cfg.DLXExchange, // 交换机名称
		"direct",        // 死信交换机固定为 direct
		cfg.Durable,     // 是否持久化
		cfg.AutoDelete,  // 是否自动删除
		false,           // 是否强制
		false,           // 是否等待服务器确认
		nil,             // 绑定参数
	); err != nil {
		return fmt.Errorf("declare dlx exchange: %w", err)
	}

	// 业务交换机
	if err := ch.ExchangeDeclare(
		cfg.ExchangeName, // 交换机名称
		exchangeType,     // 交换机类型
		cfg.Durable,      // 是否持久化
		cfg.AutoDelete,   // 是否自动删除
		false,            // 是否强制
		false,            // 是否等待服务器确认
		nil,              // 绑定参数
	); err != nil {
		return fmt.Errorf("declare exchange: %w", err)
	}

	// 死信队列（不设 DLX，避免 DLQ 内 Nack 形成循环）
	dlqArgs := amqp.Table{}
	if cfg.DLQMessageTTLMs > 0 {
		dlqArgs["x-message-ttl"] = int32(cfg.DLQMessageTTLMs)
	}
	if err := declareQueue(ch, cfg.DLQName, cfg.Durable, cfg.AutoDelete, cfg.Exclusive, dlqArgs); err != nil {
		return fmt.Errorf("declare dlq: %w", err)
	}
	if err := ch.QueueBind(
		cfg.DLQName,              // 队列名称
		cfg.DeadLetterRoutingKey, // 路由键
		cfg.DLXExchange,          // 交换机名称
		false,                    // 是否强制
		nil,                      // 绑定参数
	); err != nil {
		return fmt.Errorf("bind dlq: %w", err)
	}

	// 业务队列：消费失败且 Nack(requeue=false) 时由 Broker 转入死信交换机
	queueArgs := amqp.Table{
		"x-dead-letter-exchange":    cfg.DLXExchange,
		"x-dead-letter-routing-key": cfg.DeadLetterRoutingKey,
	}
	if err := declareQueue(ch, cfg.QueueName, cfg.Durable, cfg.AutoDelete, cfg.Exclusive, queueArgs); err != nil {
		return fmt.Errorf("declare queue: %w", err)
	}
	if err := ch.QueueBind(
		cfg.QueueName,    // 队列名称
		cfg.RoutingKey,   // 路由键
		cfg.ExchangeName, // 交换机名称
		false,            // 是否强制
		nil,              // 绑定参数
	); err != nil {
		return fmt.Errorf("bind queue: %w", err)
	}

	common.Log.Infof("[RabbitMQ] 队列已声明 exchange=%s queue=%s dlx=%s dlq=%s",
		cfg.ExchangeName, cfg.QueueName, cfg.DLXExchange, cfg.DLQName)
	return nil
}

/**
 * @description: 声明队列；若与 Broker 已有队列参数冲突（如旧版无 DLX），则删除后重建
 */
func declareQueue(ch *amqp.Channel, name string, durable, autoDelete, exclusive bool, args amqp.Table) error {
	_, err := ch.QueueDeclare(
		name,
		durable,
		autoDelete,
		exclusive,
		false, // noWait：false 表示等待 Broker 确认声明成功
		args,
	)
	if err == nil {
		return nil
	}
	if !isPreconditionFailed(err) {
		return err
	}

	common.Log.Warnf("[RabbitMQ] 队列 %s 参数与现网不一致（常见于启用死信前已创建的队列），尝试删除后重建", name)
	if _, delErr := ch.QueueDelete(name, false, false, false); delErr != nil {
		return fmt.Errorf("%w（删除旧队列失败: %v，请在 RabbitMQ 管理端手动删除队列 %s 后重启）", err, delErr, name)
	}
	_, err = ch.QueueDeclare(name, durable, autoDelete, exclusive, false, args)
	return err
}

func isPreconditionFailed(err error) bool {
	var amqpErr *amqp.Error
	if errors.As(err, &amqpErr) {
		return amqpErr.Code == 406
	}
	return strings.Contains(err.Error(), "PRECONDITION_FAILED")
}

/**
 * @description: 发布消息
 * @param {string} queueName 队列名称（兼容旧调用，实际走 exchange + routing key）
 * @param {interface{}} message 消息
 * @return {error}
 */
func PublishMessage(queueName string, message interface{}) error {
	_ = queueName
	if err := ensureConnected(); err != nil {
		common.Log.Errorf("初始化RabbitMQ连接失败: %s", err)
		return err
	}
	// 序列化消息
	body, err := json.Marshal(message)
	if err != nil {
		common.Log.Errorf("序列化消息失败: %s", err)
		return err
	}
	// 设置消息持久化
	deliveryMode := amqp.Transient
	if config.Conf.Rabbitmq.Durable {
		deliveryMode = amqp.Persistent
	}
	cfg := config.Conf.Rabbitmq

	mu.Lock()
	defer mu.Unlock()
	if publishCh == nil {
		return fmt.Errorf("publish channel not ready")
	}
	// 发布消息
	err = publishCh.Publish(
		cfg.ExchangeName, // 交换机名称
		cfg.RoutingKey,   // 路由键
		false,            // mandatory：交换机找不到队列时是否返回错误
		false,            // immediate：队列无消费者时是否返回错误（RabbitMQ 3.0+ 已废弃）
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: deliveryMode, // 与队列 durable 策略一致
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
	// 如果连接未就绪，则初始化 RabbitMQ 连接
	if err := ensureConnected(); err != nil {
		common.Log.Errorf("初始化RabbitMQ连接失败: %s", err)
		return nil, err
	}

	mu.Lock()
	conn := connection
	mu.Unlock()
	if conn == nil || conn.IsClosed() {
		return nil, fmt.Errorf("connection not ready")
	}

	// 每个消费者独立 channel，避免主消费 / DLQ 消费 / 探针争用同一 channel
	consumeCh, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	// 设置QoS（质量保证服务），确保每个消费者只接收指定数量的消息
	if err := consumeCh.Qos(
		config.Conf.Rabbitmq.PrefetchCount,
		0,     // prefetch size：未使用
		false, // global：仅针对本 channel
	); err != nil {
		_ = consumeCh.Close()
		return nil, err
	}
	// 消费消息
	msgs, err := consumeCh.Consume(
		queueName,                    // 队列名称
		consumerName,                 // 消费者名称
		config.Conf.Rabbitmq.AutoAck, // 是否自动确认（ACK）
		false,                        // exclusive
		false,                        // noLocal（RabbitMQ 未实现，保持 false）
		false,                        // noWait
		nil,
	)
	if err != nil {
		_ = consumeCh.Close()
		return nil, err
	}
	return msgs, nil
}

/**
 * @description: 关闭RabbitMQ连接 释放资源
 */
func CloseRabbitMQ() {
	mu.Lock()
	defer mu.Unlock()
	resetLocked()
}

/**
 * @description: 用于就绪探针：连接未关闭且可对业务队列做 passive 探测。
 * @return {error}
 */
func IsReady() error {
	if err := ensureConnected(); err != nil {
		return err
	}

	mu.Lock()
	conn := connection
	mu.Unlock()

	// 探针使用独立 channel，不与消费 Ack/Nack 争用
	probeCh, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("probe channel: %w", err)
	}
	defer probeCh.Close()

	cfg := config.Conf.Rabbitmq
	// 对配置中的业务队列执行 QueueInspect（passive），确认 Broker 与队列可用
	if _, err := probeCh.QueueInspect(cfg.QueueName); err != nil {
		return fmt.Errorf("queue inspect: %w", err)
	}
	if _, err := probeCh.QueueInspect(cfg.DLQName); err != nil {
		return fmt.Errorf("dlq inspect: %w", err)
	}
	return nil
}
