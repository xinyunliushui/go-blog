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
		menuRouter.GET("/access/tree/:userId", menuController.GetUserMenuTreeByUserId)
	}
	return apiGroup
}
