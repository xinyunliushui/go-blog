/*
 * @Date: 2026-03-31 17:03:12
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-05-14 13:50:37
 * @Description: 登录相关 routes
 */
package routes

import (
	"go-blog/internal/controller"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

// 初始化认证路由
func InitAuthRoutes(apiGroup *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) gin.IRoutes {
	authRouter := apiGroup.Group("/auth")
	userController := controller.NewUserController()
	{
		authRouter.POST("/login", authMiddleware.LoginHandler)
		authRouter.POST("/logout", authMiddleware.LogoutHandler)
		authRouter.POST("/refreshToken", authMiddleware.RefreshHandler)
		authRouter.POST("/register", userController.CreateUser)
	}
	return apiGroup
}
