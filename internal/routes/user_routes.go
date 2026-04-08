/*
 * @Date: 2026-03-25 22:07:36
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-02 22:17:24
 * @Description: user routes
 */
package routes

import (
	"go-blog/internal/controller"
	"go-blog/internal/middleware"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func InitUserRoutes(apiGroup *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) gin.IRoutes {
	userController := controller.NewUserController()
	userRouter := apiGroup.Group("/user")
	// 开启jwt认证中间件
	userRouter.Use(authMiddleware.MiddlewareFunc())
	// 启用Casbin中间件（进行权限认证）
	userRouter.Use(middleware.CasbinMiddleware())
	{
		userRouter.GET("/list", userController.GetUsers)
		userRouter.GET("/info", userController.GetUserInfo)
		userRouter.GET("/userInfo", userController.GetUserInfo)
	}
	return apiGroup
}
