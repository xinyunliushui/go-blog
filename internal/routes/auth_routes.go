/*
 * @Date: 2026-03-31 17:03:12
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-03-31 23:37:39
 * @Description: 登录相关 routes
 */
package routes

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func InitAuthRoutes(apiGroup *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) gin.IRoutes {
	authRouter := apiGroup.Group("/auth")
	{
		authRouter.POST("/login", authMiddleware.LoginHandler)
		authRouter.POST("/logout", authMiddleware.LogoutHandler)
		authRouter.POST("/refreshToken", authMiddleware.RefreshHandler)
	}
	return apiGroup
}
