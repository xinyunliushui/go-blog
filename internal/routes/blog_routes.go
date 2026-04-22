/*
 * @Date: 2026-04-22 15:58:00
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-22 15:59:56
 * @Description: 博客路由
 */
package routes

import (
	"go-blog/internal/controller"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func InitBlogRoutes(apiGroup *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) gin.IRoutes {
	blogController := controller.NewBlogController()
	// TODO:文章发布不需要登录
	blogRouter := apiGroup.Group("/blog")
	// 开启jwt认证中间件
	blogRouter.Use(authMiddleware.MiddlewareFunc())
	{
		blogRouter.GET("/list", blogController.GetBlogs)
		blogRouter.POST("/create", blogController.CreateBlog)
	}
	return apiGroup
}
