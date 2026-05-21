/*
 * @Date: 2026-05-20 10:01:48
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-05-20 10:02:06
 * @Description: trace_middleware.go
 */
package middleware

import (
	"time"

	"go-blog/internal/common"

	"github.com/gin-gonic/gin"
)

// TraceMiddleware 解析或生成 traceId，写入 context/gin，并在响应头回传。
func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 解析或生成 traceId
		traceID := common.ResolveInboundTraceID(c)
		if traceID == "" {
			traceID = common.NewTraceID()
		}

		// 将 traceId 写入 context/gin，并在响应头回传
		ctx := common.ContextWithTraceID(c.Request.Context(), traceID)
		c.Request = c.Request.WithContext(ctx)
		c.Set(common.GinContextTraceKey, traceID)
		c.Writer.Header().Set(common.HeaderTraceID, traceID)

		// 记录请求开始时间
		start := time.Now()
		// 处理请求
		c.Next()

		// 记录请求结束时间
		if logger := common.LoggerFromGin(c); logger != nil {
			path := c.FullPath()
			if path == "" {
				path = c.Request.URL.Path
			}
			logger.Infof("[%s] %s status=%d latency=%s",
				c.Request.Method, path, c.Writer.Status(), time.Since(start))
		}
	}
}
