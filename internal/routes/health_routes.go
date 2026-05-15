/*
 * @Date: 2026-05-15 10:18:16
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-05-15 11:19:59
 * @Description: 健康检查（不鉴权、不挂业务 API 版本前缀，便于探针与网关探测）
 */

package routes

import (
	"context"
	"go-blog/internal/clickhouse"
	"go-blog/internal/common"
	"go-blog/internal/elasticsearch"
	"go-blog/internal/rabbitmq"
	"go-blog/internal/response"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// 就绪探针超时时间
const readyProbeTimeout = 5 * time.Second

/**
 * @description: 注册存活/就绪探针
 * @param {*gin.Engine} router
 * @return {*}
 */
func InitHealthRoutes(router *gin.Engine) {
	// health（存活探针）：检查应用是否还“活着”，失败就重启
	router.GET("/health", func(c *gin.Context) {
		response.Success(c, gin.H{"status": "ok"}, "ok")
	})
	// ready（就绪探针）：MySQL、RabbitMQ、Elasticsearch、ClickHouse 均可用才返回 200
	router.GET("/ready", readyHandler)
}

/**
 * @description: 就绪探针处理函数
 * @param {*gin.Context} c
 * @return {*}
 */
func readyHandler(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), readyProbeTimeout)
	defer cancel()

	checks := gin.H{}    // 检查结果
	var errMsgs []string // 错误信息

	// MySQL（GORM 有 context，独立超时）
	if common.DB == nil {
		checks["mysql"] = "not initialized"
		errMsgs = append(errMsgs, "mysql: not initialized")
	} else if sqlDB, err := common.DB.DB(); err != nil {
		checks["mysql"] = err.Error()
		errMsgs = append(errMsgs, "mysql: "+err.Error())
	} else if err := sqlDB.PingContext(ctx); err != nil {
		checks["mysql"] = err.Error()
		errMsgs = append(errMsgs, "mysql: "+err.Error())
	} else {
		checks["mysql"] = "ok"
	}

	// RabbitMQ就绪
	if err := rabbitmq.IsReady(); err != nil {
		checks["rabbitmq"] = err.Error()
		errMsgs = append(errMsgs, "rabbitmq: "+err.Error())
	} else {
		checks["rabbitmq"] = "ok"
	}

	// Elasticsearch
	if elasticsearch.ESClient == nil {
		checks["elasticsearch"] = "not initialized"
		errMsgs = append(errMsgs, "elasticsearch: not initialized")
	} else {
		res, err := elasticsearch.ESClient.Info(elasticsearch.ESClient.Info.WithContext(ctx))
		if err != nil {
			checks["elasticsearch"] = err.Error()
			errMsgs = append(errMsgs, "elasticsearch: "+err.Error())
		} else {
			defer res.Body.Close()
			if res.IsError() {
				msg := res.String()
				checks["elasticsearch"] = msg
				errMsgs = append(errMsgs, "elasticsearch: "+msg)
			} else {
				checks["elasticsearch"] = "ok"
			}
		}
	}

	// ClickHouse
	if clickhouse.ClickHouseDB == nil {
		checks["clickhouse"] = "not initialized"
		errMsgs = append(errMsgs, "clickhouse: not initialized")
	} else if sqlDB, err := clickhouse.ClickHouseDB.DB(); err != nil {
		checks["clickhouse"] = err.Error()
		errMsgs = append(errMsgs, "clickhouse: "+err.Error())
	} else if err := sqlDB.PingContext(ctx); err != nil {
		checks["clickhouse"] = err.Error()
		errMsgs = append(errMsgs, "clickhouse: "+err.Error())
	} else {
		checks["clickhouse"] = "ok"
	}

	if len(errMsgs) > 0 {
		response.Response(c, http.StatusServiceUnavailable, 503, gin.H{"checks": checks}, strings.Join(errMsgs, ";"))
		return
	}
	// 返回就绪探针结果
	response.Success(c, gin.H{"status": "ready", "checks": checks}, "ok")
}
