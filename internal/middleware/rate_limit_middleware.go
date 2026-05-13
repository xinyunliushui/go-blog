package middleware

import (
	"go-blog/internal/response"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/juju/ratelimit"
)

/** 限流中间件
 * @param fillInterval time.Duration 填充一个令牌需要的时间间隔
 * @param capacity int64 桶容量
 * @return gin.HandlerFunc 限流中间件
 */
func RateLimitMiddleware(fillInterval time.Duration, capacity int64) gin.HandlerFunc {
	bucket := ratelimit.NewBucket(fillInterval, capacity)
	return func(c *gin.Context) {
		if bucket.TakeAvailable(1) < 1 {
			response.Fail(c, nil, "访问限流")
			c.Abort()
			return
		}
		c.Next()
	}
}
