/*
 * @Description: 请求链路 traceId 上下文与日志辅助
 */
package common

import (
	"context"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

const (
	// GinContextTraceKey 在 gin.Context 中存储 traceId 的键名。
	GinContextTraceKey = "traceId"
	// HeaderTraceID 入站/出站 HTTP 头。
	HeaderTraceID = "X-Trace-Id"
	// HeaderTraceIDFallback 兼容网关常用的请求 ID 头。
	HeaderTraceIDFallback = "X-Request-Id"
	// MQHeaderTraceID RabbitMQ 消息头中的 traceId 键名。
	MQHeaderTraceID = "traceId"
)

type traceIDContextKey struct{}

var traceIDKey traceIDContextKey

// NewTraceID 生成新的 traceId。
func NewTraceID() string {
	return uuid.New().String()
}

// ContextWithTraceID 将 traceId 写入标准 context。
func ContextWithTraceID(ctx context.Context, traceID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	traceID = strings.TrimSpace(traceID)
	if traceID == "" {
		return ctx
	}
	return context.WithValue(ctx, traceIDKey, traceID)
}

// TraceIDFromCtx 从标准 context 读取 traceId。
func TraceIDFromCtx(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v, ok := ctx.Value(traceIDKey).(string); ok {
		return strings.TrimSpace(v)
	}
	return ""
}

// TraceIDFromGin 从 gin.Context 读取 traceId（优先 context，其次 gin key）。
func TraceIDFromGin(c *gin.Context) string {
	if c == nil {
		return ""
	}
	if id := TraceIDFromCtx(c.Request.Context()); id != "" {
		return id
	}
	if v, ok := c.Get(GinContextTraceKey); ok {
		if id, ok := v.(string); ok {
			return strings.TrimSpace(id)
		}
	}
	return ""
}

// ResolveInboundTraceID 解析入站 HTTP 头中的 traceId。
func ResolveInboundTraceID(c *gin.Context) string {
	if c == nil {
		return ""
	}
	if id := strings.TrimSpace(c.GetHeader(HeaderTraceID)); id != "" {
		return id
	}
	return strings.TrimSpace(c.GetHeader(HeaderTraceIDFallback))
}

// TraceIDFromAMQPHeaders 从 RabbitMQ 消息头读取 traceId。
func TraceIDFromAMQPHeaders(headers amqp.Table) string {
	if headers == nil {
		return ""
	}
	v, ok := headers[MQHeaderTraceID]
	if !ok {
		return ""
	}
	switch s := v.(type) {
	case string:
		return strings.TrimSpace(s)
	case []byte:
		return strings.TrimSpace(string(s))
	default:
		return strings.TrimSpace(fmt.Sprint(v))
	}
}

// LoggerFromCtx 返回带 traceId 字段的 logger；无 traceId 时返回全局 logger。
func LoggerFromCtx(ctx context.Context) *zap.SugaredLogger {
	if Log == nil {
		return nil
	}
	if id := TraceIDFromCtx(ctx); id != "" {
		return Log.With("traceId", id)
	}
	return Log
}

// LoggerFromGin 返回带 traceId 字段的 logger。
func LoggerFromGin(c *gin.Context) *zap.SugaredLogger {
	if Log == nil {
		return nil
	}
	if id := TraceIDFromGin(c); id != "" {
		return Log.With("traceId", id)
	}
	return Log
}

// TraceIDForPublish 发布 MQ 时使用的 traceId：优先沿用请求上下文，否则新建。
func TraceIDForPublish(ctx context.Context) string {
	if id := TraceIDFromCtx(ctx); id != "" {
		return id
	}
	return NewTraceID()
}

// ResolveTraceIDForCompensation 写入补偿表时使用的 traceId。
func ResolveTraceIDForCompensation(traceID string) string {
	if id := strings.TrimSpace(traceID); id != "" {
		return id
	}
	return NewTraceID()
}

// CompensationTraceID 补偿重试时用于日志/MQ 的 traceId：优先读库，历史数据兜底。
func CompensationTraceID(storedTraceID, recordID string) string {
	if id := strings.TrimSpace(storedTraceID); id != "" {
		return id
	}
	if id := strings.TrimSpace(recordID); id != "" {
		return fmt.Sprintf("compensation-%s", id)
	}
	return NewTraceID()
}
