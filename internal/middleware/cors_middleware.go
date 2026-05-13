/*
 * @Date: 2026-04-01 00:01:18
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-02 09:28:30
 * @Description:
 */
package middleware

import (
	"net/http"
	"slices"
	"strings"

	"go-blog/internal/config"

	"github.com/gin-gonic/gin"
)

/** CORS跨域中间件 说明：带 credentials 的请求不能使用 Allow-Origin: *，须为具体 Origin。
 * @return gin.HandlerFunc CORS跨域中间件
 */
func CORSMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		method := ctx.Request.Method
		origin := strings.TrimSpace(ctx.Request.Header.Get("Origin"))
		// 从配置名单中加载允许的跨域请求源
		allowOrigin := ""
		allowed := config.Conf.Application.CorsAllowOrigins
		if len(allowed) > 0 {
			if origin != "" && slices.Contains(allowed, origin) {
				allowOrigin = origin
			}
		} else if origin != "" {
			allowOrigin = origin
		}

		if allowOrigin != "" {
			ctx.Writer.Header().Set("Access-Control-Allow-Origin", allowOrigin)
			//服务器支持的所有跨域请求的方法
			ctx.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			//允许跨域设置可以返回其他子段，可以自定义字段
			ctx.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, Content-Length, X-CSRF-Token, Token, Session")
			// 允许浏览器（客户端）可以解析的头部 （重要）
			ctx.Header("Access-Control-Expose-Headers", "Content-Length, Content-Type, Access-Control-Allow-Origin, Access-Control-Allow-Headers")
			//设置缓存时间
			ctx.Header("Access-Control-Max-Age", "172800")
			//允许客户端传递校验信息比如 cookie (重要)
			ctx.Header("Access-Control-Allow-Credentials", "true")
			//避免缓存错误复用跨域响应
			ctx.Header("Vary", "Origin")
		}

		// 允许类型校验
		if method == http.MethodOptions {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}

		ctx.Next()
	}
}
