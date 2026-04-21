/*
 * @Date: 2026-04-01 22:02:21
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-21 15:15:46
 * @Description: 初始化管理员用户和角色策略
 */
package common

import (
	"errors"
	"go-blog/internal/model"
	"go-blog/internal/utils"

	"gorm.io/gorm"
)

func InitAdmin(db *gorm.DB) {
	// 1. 创建默认管理员用户（如果不存在）
	var admin model.User
	roles := []*model.Role{
		{
			Model:   gorm.Model{ID: 1},
			Name:    "管理员",
			Keyword: "role_admin",
			Desc:    new(string),
			Sort:    1, // 排序越大权限越低
			Status:  1, // 正常
			Creator: "系统",
		},
	}
	result := db.Where("username = ?", "admin").First(&admin)
	if result.Error != nil {

		// 用户不存在，创建新管理员
		admin = model.User{
			Username:     "admin",
			Password:     utils.GenPasswd("admin123"),
			Mobile:       "18888888888",
			Avatar:       "https://gips1.baidu.com/it/u=1397155327,2622199615&fm=3074&app=3074&f=PNG?w=2048&h=2048",
			Nickname:     "动物园园长",
			Introduction: "系统默认管理员",
			Status:       1,
			Creator:      "系统",
			Roles:        roles[:1], // 可选角色字段，用于业务逻辑
		}
		if err := db.Create(&admin).Error; err != nil {
			Log.Errorf("创建管理员用户失败: %v", err)
		}
	}

	// 2写入菜单
	newMenus := make([]model.Menu, 0)
	var uint0 uint = 0
	var uint1 uint = 1
	sysIconStr := "SettingOutlined"
	systemUserStr := "/system/users"
	usersIconStr := "UserOutlined"
	rolesIconStr := "TeamOutlined"
	menusIconStr := "AppstoreOutlined"
	menus := []model.Menu{
		{
			Model:    gorm.Model{ID: 1},
			Name:     "system",
			Title:    "系统管理",
			Icon:     &sysIconStr,
			Path:     "system",
			Redirect: &systemUserStr,
			Sort:     10,
			ParentId: &uint0,
			Roles:    roles[:1],
			Creator:  "系统",
		},
		{
			Model:    gorm.Model{ID: 2},
			Name:     "Users",
			Title:    "用户管理",
			Icon:     &usersIconStr,
			Path:     "users",
			Sort:     11,
			ParentId: &uint1,
			Roles:    roles[:1],
			Creator:  "系统",
		},
		{
			Model:    gorm.Model{ID: 3},
			Name:     "Roles",
			Title:    "角色管理",
			Icon:     &rolesIconStr,
			Path:     "roles",
			Sort:     12,
			ParentId: &uint1,
			Roles:    roles[:1],
			Creator:  "系统",
		},
		{
			Model:    gorm.Model{ID: 4},
			Name:     "Resources",
			Title:    "资源管理",
			Icon:     &menusIconStr,
			Path:     "resources",
			Sort:     13,
			ParentId: &uint1,
			Roles:    roles[:1],
			Creator:  "系统",
		},
	}
	for _, menu := range menus {
		err := DB.First(&menu, menu.ID).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newMenus = append(newMenus, menu)
		}
	}
	if len(newMenus) > 0 {
		err := DB.Create(&newMenus).Error
		if err != nil {
			Log.Errorf("创建管理员菜单数据失败: %v", err)
		}
	}
	Log.Info("初始化管理员用户和角色策略完成！")
}
