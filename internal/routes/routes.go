/*
 * @Date: 2026-03-25 22:11:30
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-01 10:41:58
 * @Description: routes
 */
package routes

import (
	"fmt"
	"go-blog/internal/config"
	"go-blog/internal/middleware"

	"github.com/gin-gonic/gin"
)

func InitRoutes() {

	router := gin.Default()
	// 启用全局跨域中间件（需在创建路由分组前挂载）
	router.Use(middleware.CORSMiddleware())

	// api分组
	apiGroup := router.Group("/" + config.Config.Application.UrlPathPrefix)
	// 初始化JWT认证中间件
	authMiddleware, err := middleware.InitAuth()
	if err != nil {
		// common.Log.Panicf("初始化JWT中间件失败：%v", err)
		panic(fmt.Sprintf("初始化JWT中间件失败：%v", err))
	}

	// 注册认证路由
	InitAuthRoutes(apiGroup, authMiddleware)

	// 注册用户路由
	InitUserRoutes(apiGroup, authMiddleware)

	// 启动
	router.Run(":8080")
}
