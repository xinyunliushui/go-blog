/*
 * @Date: 2026-04-08 21:24:51
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-13 14:36:23
 * @Description:
 */
package routes

import (
	"go-blog/internal/controller"
	"go-blog/internal/middleware"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func InitMenuRoutes(apiGroup *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) gin.IRoutes {
	menuController := controller.NewMenuController()
	menuRouter := apiGroup.Group("/menu")
	// 开启jwt认证中间件
	menuRouter.Use(authMiddleware.MiddlewareFunc())
	// 启用Casbin中间件（进行权限认证）
	menuRouter.Use(middleware.CasbinMiddleware())
	{
		menuRouter.GET("/tree", menuController.GetMenuTree)
		menuRouter.GET("/list", menuController.GetMenus)
		menuRouter.POST("/create", menuController.CreateMenu)
		menuRouter.POST("/update/:menuId", menuController.UpdateMenuById)
		menuRouter.GET("/access/tree/:userId", menuController.GetUserMenuTreeByUserId)
	}
	return apiGroup
}
