package routes

import (
	"go-blog/internal/controller"
	"go-blog/internal/middleware"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func InitRoleRoutes(apiGroup *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) gin.IRoutes {
	roleController := controller.NewRoleController()
	roleRouter := apiGroup.Group("/role")
	// 开启jwt认证中间件
	roleRouter.Use(authMiddleware.MiddlewareFunc())
	// 启用Casbin中间件（进行权限认证）
	roleRouter.Use(middleware.CasbinMiddleware())
	{
		roleRouter.GET("/list", roleController.GetRoles)
	}
	return apiGroup
}
