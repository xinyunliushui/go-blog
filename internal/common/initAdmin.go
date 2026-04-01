package common

import (
	"go-blog/internal/model"
	"log"

	"github.com/casbin/casbin/v3"
	"gorm.io/gorm"
)

func InitAdmin(db *gorm.DB, enforcer *casbin.Enforcer) {
	// 1. 创建默认管理员用户（如果不存在）
	var admin model.User
	result := db.Where("username = ?", "admin").First(&admin)
	if result.Error != nil {
		role := []*model.Role{
			{
				Model:   gorm.Model{ID: 1},
				Name:    "管理员",
				Keyword: "role_admin",
				Desc:    new(string),
				Sort:    1,
				Status:  1,
				Creator: "系统",
			},
		}
		// 用户不存在，创建新管理员
		admin = model.User{
			Username: "admin",
			Password: "admin123",
			Roles:    role[:1], // 可选角色字段，用于业务逻辑
		}
		if err := db.Create(&admin).Error; err != nil {
			log.Fatalf("创建管理员用户失败: %v", err)
		}
	}

	// 1. 定义角色策略: p, role_admin, *, *
	// 2. 将用户加入角色: g, admin, role_admin
	// 检查并添加角色策略

	hasRolePolicy, _ := enforcer.HasPolicy("role_admin", "*", "*")
	if !hasRolePolicy {
		_, err := enforcer.AddPolicy("role_admin", "*", "*")
		if err != nil {
			log.Fatalf("添加角色策略失败: %v", err)
		}
	}
	// 检查并添加用户角色绑定
	hasGrouping, _ := enforcer.HasGroupingPolicy("admin", "role_admin")
	if !hasGrouping {
		_, err := enforcer.AddGroupingPolicy("admin", "role_admin")
		if err != nil {
			log.Fatalf("添加用户角色绑定失败: %v", err)
		}
	}
}
