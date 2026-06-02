/*
 * @Author: zhongwh 746227367@qq.com
 * @Date: 2026-06-02 14:48:32
 * @LastEditors: zhongwh 746227367@qq.com
 * @LastEditTime: 2026-06-02 16:22:14
 * @FilePath: \golang\go-blog\internal\service\mq_background_workers.go
 * @Description: 这是默认设置,请设置`customMade`, 打开koroFileHeader查看配置 进行设置: https://github.com/OBKoro1/koro1FileHeader/wiki/%E9%85%8D%E7%BD%AE
 */
/*
 * @Description: 后台 MQ 消费者与补偿任务的生命周期管理（减少 main 中 WaitGroup/闭包逃逸）
 */
package service

import (
	"context"
	"sync"
)

// BackgroundWorkers 统一管理 RabbitMQ 主消费、DLQ 消费与 MQ 补偿重试。
type BackgroundWorkers struct {
	wg sync.WaitGroup

	stopCompensation context.CancelFunc
	stopConsumer     context.CancelFunc
	stopDLQ          context.CancelFunc
}

/**
 * @description: 启动后台任务在 root 下启动全部后台任务；返回的实例需在关闭 MQ 前调用 Stop。
 * @param root context.Context 上下文
 * @return *BackgroundWorkers 后台任务实例
 */
func StartMqBackgroundWorkers(root context.Context) *BackgroundWorkers {
	w := &BackgroundWorkers{}

	consumerCtx, stopConsumer := context.WithCancel(root)
	dlqCtx, stopDLQ := context.WithCancel(root)
	compensationCtx, stopCompensation := context.WithCancel(root)
	w.stopConsumer = stopConsumer
	w.stopDLQ = stopDLQ
	w.stopCompensation = stopCompensation

	w.wg.Add(3)
	go w.runConsumer(consumerCtx)
	go w.runDLQ(dlqCtx)
	go w.runCompensation(compensationCtx)
	return w
}

/**
 * @description: 消费 MQ 消息
 * @param ctx context.Context 上下文
 */
func (w *BackgroundWorkers) runConsumer(ctx context.Context) {
	defer w.wg.Done()
	ConsumeRabbitMQ(ctx, HandleArticleMessage)
}

/**
 * @description: 消费 DLQ 消息
 * @param ctx context.Context 上下文
 */
func (w *BackgroundWorkers) runDLQ(ctx context.Context) {
	defer w.wg.Done()
	ConsumeRabbitMQDLQ(ctx)
}

/**
 * @description: 消费补偿消息
 * @param ctx context.Context 上下文
 */
func (w *BackgroundWorkers) runCompensation(ctx context.Context) {
	defer w.wg.Done()
	RunBlogMQCompensationRetry(ctx)
}

// Stop 按依赖顺序取消任务并等待全部退出（须在 rabbitmq.CloseRabbitMQ 之前调用）。
func (w *BackgroundWorkers) Stop() {
	w.stopCompensation()
	w.stopConsumer()
	w.stopDLQ()
	w.wg.Wait()
}
