/*
 * @Date: 2026-03-18 21:50:24
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-05-06 17:24:19
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
	"syscall"
	"time"
)

func main() {
	// 初始化配置
	if err := config.InitConfig(); err != nil {
		log.Fatalf("初始化配置失败: %v", err)
		return
	}

	// 初始化日志
	common.InitLogger()

	// 初始化数据库
	common.InitMysql()

	// 初始化validator（汉化）
	common.InitValidator()

	// 初始化mysql数据
	common.InitAdmin(common.DB)

	// 初始化RabbitMQ
	if err := rabbitmq.InitRabbitMQ(); err != nil {
		log.Fatalf("初始化 RabbitMQ 失败: %v", err)
	}
	defer rabbitmq.CloseRabbitMQ()

	// 初始化Elasticsearch
	if err := elasticsearch.InitESClient(); err != nil {
		log.Fatalf("初始化 Elasticsearch 失败: %v", err)
	}

	// 初始化ClickHouse
	if err := clickhouse.InitClickHouse(); err != nil {
		log.Fatalf("初始化 ClickHouse 失败: %v", err)
	}

	// 启动消费者 , 消费RabbitMQ消息
	go func() {
		service.ConsumeRabbitMQ(service.HandleArticleMessage)
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
	common.Log.Info("关闭服务中...")
	// 优雅关闭HTTP Server（给正在处理的请求最多5秒完成）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		common.Log.Errorf("服务关闭出错: %v", err)
	}
	common.Log.Info("服务已关闭")
}
