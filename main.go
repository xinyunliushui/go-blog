/*
 * @Date: 2026-03-18 21:50:24
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-01 22:23:46
 * @Description: main
 */
package main

import (
	"go-blog/internal/common"
	"go-blog/internal/config"
	"go-blog/internal/routes"
)

func main() {
	// 初始化配置
	config.InitConfig()

	// 初始化日志

	// 初始化数据库
	common.InitMysql()

	// 初始化casbin管理器
	common.InitCasbinEnforcer()

	// 初始化validator（汉化）
	// common.InitValidator()

	// 初始化mysql数据
	// common.InitMysqlData()
	common.InitAdmin(common.DB, common.CasbinEnforcer)

	// 初始化路由服务
	routes.InitRoutes()
	// 初始化routes
	// r := gin.Default()
	// r.GET("/test", func(c *gin.Context) {
	// 	c.JSON(200, gin.H{
	// 		"message": "Hello, World!",
	// 	})
	// })
	// r.Run(":8080")
}
