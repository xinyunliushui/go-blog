/*
 * @Date: 2026-03-18 21:50:24
 * @Author: zhongwenhao
 * @LastEditors: zhongwh 746227367@qq.com
 * @LastEditTime: 2026-06-02 16:22:21
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

// listenAndServe 阻塞监听 HTTP；供独立 goroutine 调用，避免 main 内闭包逃逸。
func listenAndServe(server *http.Server) {
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		common.Log.Errorf("listen: %s\n", err)
	}
}

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

	// nq后台任务：单一 WaitGroup + 命名方法，减少 main 栈变量与闭包逃逸
	rootCtx := context.Background()
	bgWorkers := service.StartMqBackgroundWorkers(rootCtx)

	// 初始化路由服务
	router := routes.InitRoutes()

	// 启动服务
	port := config.Conf.Application.Port

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router.Handler(),
	}

	go listenAndServe(server)

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
	ctx, cancel := context.WithTimeout(rootCtx, 5*time.Second)
	defer cancel()
	// 先关流量 -> 再关中间件 -> 再释放资源
	// 1. 停止接受新的 HTTP 请求，等待已有请求结束或超时
	if err := server.Shutdown(ctx); err != nil {
		common.Log.Errorf("HTTP 服务关闭出错: %v", err)
	}

	// 2. 停止 MQ 补偿与消费者（须在关闭 RabbitMQ 连接之前）
	bgWorkers.Stop()

	// 3. 关闭 RabbitMQ 连接
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
