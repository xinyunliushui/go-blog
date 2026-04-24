/*
 * @Date: 2026-04-22 15:58:00
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-24 10:18:23
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
	// 公开路由：不需要登录录认证
	blogPublicRouter := apiGroup.Group("/blog")
	{
		blogPublicRouter.GET("/list", blogController.GetBlogs)
		blogPublicRouter.GET("/detail/:blogId", blogController.GetBlogById)
	}
	// 管理路由：需要登录认证
	blogRouter := apiGroup.Group("/blog")
	blogRouter.Use(authMiddleware.MiddlewareFunc())
	{
		blogRouter.POST("/create", blogController.CreateBlog)
		blogRouter.POST("/publish/:blogId", blogController.UpdateBlogPublishStatusById)
		blogRouter.POST("/update/:blogId", blogController.UpdateBlogById)
	}
	return apiGroup
}
