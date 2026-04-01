/*
 * @Date: 2026-03-25 22:07:36
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-03-25 23:58:15
 * @Description: user routes
 */
package routes

import (
	"go-blog/internal/controller"

	"github.com/gin-gonic/gin"
)

func InitUserRoutes(apiGroup *gin.RouterGroup) gin.IRoutes {
	userController := controller.NewUserController()
	userRouter := apiGroup.Group("/user")
	{
		userRouter.POST("/userInfo", userController.GetUserInfo)
	}
	return apiGroup
}
