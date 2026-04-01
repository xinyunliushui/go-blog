/*
 * @Date: 2026-03-24 23:02:37
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-03-25 23:28:18
 * @Description: casbin enforcer
 */
package common

import (
	"fmt"
	"go-blog/internal/config"

	"github.com/casbin/casbin/v3"
	gormadapter "github.com/casbin/gorm-adapter/v3"
)

// 全局CasbinEnforcer
var CasbinEnforcer *casbin.Enforcer

func InitCasbinEnforcer() {
	enforcer, err := initSqliteCasbin()
	if err != nil {
		panic(fmt.Sprintf("初始化Casbin失败：%v", err))
	}
	CasbinEnforcer = enforcer
	fmt.Println("初始化Casbin完成!")
}

func initSqliteCasbin() (*casbin.Enforcer, error) {
	// 创建gorm适配器
	adapter, err := gormadapter.NewAdapterByDB(DB)
	if err != nil {
		return nil, err
	}

	// 创建casbin管理器
	enforcer, err := casbin.NewEnforcer(config.Config.Casbin.ModelPath, adapter)
	if err != nil {
		return nil, err
	}
	// 加载策略
	err = enforcer.LoadPolicy()
	if err != nil {
		return nil, err
	}
	return enforcer, nil
}
