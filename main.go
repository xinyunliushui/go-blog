/*
 * @Date: 2026-03-18 21:50:24
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-05-05 11:34:01
 * @Description: main
 */
package main

import (
	"go-blog/internal/clickhouse"
	"go-blog/internal/common"
	"go-blog/internal/config"
	"go-blog/internal/elasticsearch"
	"go-blog/internal/rabbitmq"
	"go-blog/internal/routes"
	"go-blog/internal/service"
	"log"
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
	routes.InitRoutes()
}
