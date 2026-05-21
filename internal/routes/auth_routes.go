/*
 * @Date: 2026-03-31 17:03:12
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-05-20 11:50:24
 * @Description: 登录相关 routes
 */
package routes

import (
	"go-blog/internal/config"
	"go-blog/internal/controller"
	"go-blog/internal/middleware"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

// 初始化认证路由
func InitAuthRoutes(apiGroup *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) gin.IRoutes {
	authRouter := apiGroup.Group("/auth")
	userController := controller.NewUserController()
	// 启用限流中间件
	// fill-interval 在 config 中为毫秒；time.Duration(50) 为纳秒，须乘 time.Millisecond
	fillInterval := time.Duration(config.Conf.RateLimit.FillInterval) * time.Millisecond
	capacity := config.Conf.RateLimit.Capacity
	authRouter.Use(middleware.RateLimitMiddleware(fillInterval, capacity))
	{
		authRouter.POST("/login", authMiddleware.LoginHandler)
		authRouter.POST("/logout", authMiddleware.LogoutHandler)
		authRouter.POST("/refreshToken", authMiddleware.RefreshHandler)
		authRouter.POST("/register", userController.CreateUser)
	}
	return apiGroup
}
