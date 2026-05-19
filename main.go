/*
 * @Date: 2026-03-18 21:50:24
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-05-07 15:16:19
 * @Description: main
 */
package main

import (
	"context"
	"fmt"
	"go-blog/internal/clickhouse"
	"go-blog/internal/common"
	"go-blog/internal/config"
	"go-blog/internal/elasticsearch"
	"go-blog/internal/rabbitmq"
	"go-blog/internal/routes"
	"go-blog/internal/service"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	// 初始化配置
	if err := config.InitConfig(); err != nil {
		log.Fatalf("初始化配置失败: %v", err)
	}

	// 初始化日志
	common.InitLogger()

	// 初始化数据库（失败不退出进程，由 /ready 反映；避免 InitAdmin 在 DB 为 nil 时崩溃）
	if err := common.InitMysql(); err != nil {
		common.Log.Errorf("初始化 MySQL 失败（服务仍将启动，就绪探针会失败）: %v", err)
	}

	// 初始化validator（汉化）
	common.InitValidator()

	if common.DB != nil {
		// 初始化mysql数据
		common.InitAdmin(common.DB)
	}

	// 初始化RabbitMQ
	if err := rabbitmq.InitRabbitMQ(); err != nil {
		common.Log.Errorf("初始化 RabbitMQ 失败（服务仍将启动，就绪探针会失败）: %v", err)
	}

	// 初始化Elasticsearch
	if err := elasticsearch.InitESClient(); err != nil {
		common.Log.Errorf("初始化 Elasticsearch 失败（服务仍将启动，就绪探针会失败）: %v", err)
	}

	// 初始化ClickHouse
	if err := clickhouse.InitClickHouse(); err != nil {
		common.Log.Errorf("初始化 ClickHouse 失败（服务仍将启动，就绪探针会失败）: %v", err)
	}

	// 启动消费者：用独立 ctx，停机时先 cancel 再等 wg，最后关闭 MQ 连接，避免泄漏 goroutine 或对已关闭连接 Ack
	consumerCtx, stopConsumer := context.WithCancel(context.Background())
	var consumerWg sync.WaitGroup
	consumerWg.Add(1)
	go func() {
		defer consumerWg.Done()
		service.ConsumeRabbitMQ(consumerCtx, service.HandleArticleMessage)
	}()

	// 启动死信队列消费者
	dlqCtx, stopDLQ := context.WithCancel(context.Background())
	var dlqWg sync.WaitGroup
	dlqWg.Add(1)
	go func() {
		defer dlqWg.Done()
		service.ConsumeRabbitMQDLQ(dlqCtx)
	}()

	// 博客 MQ 补偿重试（PUBLISH 补 MQ / CONSUME 补 ES+CH，先于关闭 MQ 连接停止）
	compensationCtx, stopCompensation := context.WithCancel(context.Background())
	var compensationWg sync.WaitGroup
	compensationWg.Add(1)
	go func() {
		defer compensationWg.Done()
		service.RunBlogMQCompensationRetry(compensationCtx)
	}()

	// 初始化路由服务
	router := routes.InitRoutes()

	// 启动服务
	port := config.Conf.Application.Port

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router.Handler(),
	}

	go func() {
		// 服务启动, 监听端口
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			common.Log.Errorf("listen: %s\n", err)
		}
	}()

	// ==================== 优雅停机处理 ====================
	// 创建带缓冲的信号channel（缓冲1，确保第一个信号能被捕获）
	quit := make(chan os.Signal, 1)
	// 注册要监听的信号：SIGINT(Ctrl+C), SIGTERM(kill/systemd/K8s)
	// kill (no params) by default sends syscall.SIGTERM，默认发送syscall.SIGTERM信号
	// kill -2 is syscall.SIGINT。kill -2 <进程ID> 或者你在终端前台按 Ctrl+C，都会发送 SIGINT 信号。用于终止进程。
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need add it。kill -9 <进程ID> 用于强制终止进程，这个信号无法被任何程序捕获、阻塞或忽略。
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	// 阻塞等待信号
	<-quit
	// 已接收到信号，停止监听quit
	signal.Stop(quit)
	common.Log.Info("开始关闭服务...")
	// 优雅关闭进程（给正在处理的请求最多5秒完成）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// 先关流量 -> 再关中间件 -> 再释放资源
	// 1. 停止接受新的 HTTP 请求，等待已有请求结束或超时
	if err := server.Shutdown(ctx); err != nil {
		common.Log.Errorf("HTTP 服务关闭出错: %v", err)
	}

	// 2. 停止 MQ 补偿重试（不再 Publish / 直写 ES+CH），再停消费者
	stopCompensation()
	compensationWg.Wait()

	// 3. 停止 RabbitMQ
	// 停止主消费与 DLQ 消费
	stopConsumer()
	consumerWg.Wait()
	// 停止死信队列消费者
	stopDLQ()
	dlqWg.Wait()
	// 关闭 RabbitMQ 连接
	rabbitmq.CloseRabbitMQ()
	common.Log.Info("RabbitMQ 连接已关闭")

	// 4. 关闭 MySQL
	if common.DB != nil {
		if sqlDB, err := common.DB.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				common.Log.Errorf("MySQL 连接关闭出错: %v", err)
			} else {
				common.Log.Info("MySQL 连接已关闭")
			}
		}
	}

	// 5. 关闭 ClickHouse
	clickhouse.CloseClickHouse()

	// go-elasticsearch/v8 客户端基于 net/http，进程退出时会回收；无单独 Close API。

	// 6. 刷盘日志（避免进程退出时丢失缓冲区）
	_ = common.Log.Sync()
	common.Log.Info("日志已刷盘")

	common.Log.Info("服务已关闭")
}
