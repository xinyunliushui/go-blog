/*
 * @Date: 2026-03-25 22:11:30
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-05-06 16:45:50
 * @Description: routes
 */
package routes

import (
	"fmt"
	"go-blog/internal/config"
	"go-blog/internal/middleware"

	"github.com/gin-gonic/gin"
)

func InitRoutes() *gin.Engine {

	router := gin.Default()
	// 启用全局跨域中间件（需在创建路由分组前挂载）
	router.Use(middleware.CORSMiddleware())

	// api分组
	apiGroup := router.Group("/" + config.Conf.Application.UrlPathPrefix)
	// V1版本
	v1ApiGroup := apiGroup.Group("/v1")
	// 初始化JWT认证中间件
	authMiddleware, err := middleware.InitAuth()
	if err != nil {
		panic(fmt.Sprintf("初始化JWT中间件失败：%v", err))
	}

	// 注册认证路由
	InitAuthRoutes(v1ApiGroup, authMiddleware)

	// 注册用户路由
	InitUserRoutes(v1ApiGroup, authMiddleware)

	// 注册角色路由
	InitRoleRoutes(v1ApiGroup, authMiddleware)

	// 注册菜单路由
	InitMenuRoutes(v1ApiGroup, authMiddleware)

	// 注册博客路由
	InitBlogRoutes(v1ApiGroup, authMiddleware)

	return router
}
