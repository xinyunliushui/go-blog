/*
 * @Date: 2026-04-22 15:58:00
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-05-05 16:34:06
 * @Description: 博客路由
 */
package routes

import (
	"go-blog/internal/config"
	"go-blog/internal/controller"
	"go-blog/internal/middleware"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

// 初始化博客路由
func InitBlogRoutes(apiGroup *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) gin.IRoutes {
	blogController := controller.NewBlogController()
	// 公开路由：不需要登录录认证
	blogPublicRouter := apiGroup.Group("/blog")
	// 启用限流中间
	// 默认每50毫秒填充一个令牌，最多填充200个
	fillInterval := time.Duration(config.Conf.RateLimit.FillInterval)
	capacity := config.Conf.RateLimit.Capacity
	blogPublicRouter.Use(middleware.RateLimitMiddleware(fillInterval, capacity))
	{
		blogPublicRouter.GET("/list", blogController.GetBlogs)
		blogPublicRouter.GET("/detail/:blogId", blogController.GetBlogById)
		blogPublicRouter.GET("/search", blogController.SearchBlogs)
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
