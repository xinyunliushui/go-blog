/*
 * @Date: 2026-03-25 22:07:36
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-13 17:37:28
 * @Description: user routes
 */
package routes

import (
	"go-blog/internal/controller"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func InitUserRoutes(apiGroup *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) gin.IRoutes {
	userController := controller.NewUserController()
	userRouter := apiGroup.Group("/user")
	// 开启jwt认证中间件
	userRouter.Use(authMiddleware.MiddlewareFunc())
	{
		userRouter.GET("/list", userController.GetUsers)
		userRouter.GET("/info", userController.GetUserInfo)
		userRouter.GET("/userInfo", userController.GetUserInfo)
		userRouter.POST("/create", userController.CreateUser)
		userRouter.POST("/update/:userId", userController.UpdateUserById)
		userRouter.POST("/changePwd", userController.ChangePwd)
	}
	return apiGroup
}
