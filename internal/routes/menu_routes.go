/*
 * @Date: 2026-04-08 21:24:51
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-21 14:11:32
 * @Description:
 */
package routes

import (
	"go-blog/internal/controller"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

// 初始化菜单路由
func InitMenuRoutes(apiGroup *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) gin.IRoutes {
	menuController := controller.NewMenuController()
	menuRouter := apiGroup.Group("/menu")
	// 开启jwt认证中间件
	menuRouter.Use(authMiddleware.MiddlewareFunc())
	{
		menuRouter.GET("/tree", menuController.GetMenuTree)
		menuRouter.GET("/list", menuController.GetMenus)
		menuRouter.POST("/create", menuController.CreateMenu)
		menuRouter.POST("/update/:menuId", menuController.UpdateMenuById)
		menuRouter.GET("/access/tree/:userId", menuController.GetUserMenuTreeByUserId)
	}
	return apiGroup
}
