package routes

import (
	"go-blog/internal/controller"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func InitRoleRoutes(apiGroup *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) gin.IRoutes {
	roleController := controller.NewRoleController()
	roleRouter := apiGroup.Group("/role")
	// 开启jwt认证中间件
	roleRouter.Use(authMiddleware.MiddlewareFunc())
	{
		roleRouter.GET("/list", roleController.GetRoles)
		roleRouter.POST("/create", roleController.CreateRole)
		roleRouter.POST("/update/:roleId", roleController.UpdateRoleById)
		roleRouter.GET("/menus/get/:roleId", roleController.GetRoleMenusById)
		roleRouter.POST("/menus/update/:roleId", roleController.UpdateRoleMenusById)
	}
	return apiGroup
}



























