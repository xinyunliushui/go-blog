/*
 * @Date: 2026-04-01 22:02:21
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-05-15 14:08:20
 * @Description: 初始化管理员用户和角色策略
 */
package common

import (
	"errors"
	"go-blog/internal/model"
	"go-blog/internal/utils"

	"gorm.io/gorm"
)

// InitAdmin 种子数据的稳定幂等键（与 migrations/202606020001 中回填 SQL 保持一致）。
const (
	seedRequestIDRoleAdmin     = "init:seed:role:role_admin"
	seedRequestIDUserAdmin     = "init:seed:user:admin"
	seedRequestIDMenuSystem    = "init:seed:menu:system"
	seedRequestIDMenuUsers     = "init:seed:menu:users"
	seedRequestIDMenuRoles     = "init:seed:menu:roles"
	seedRequestIDMenuResources = "init:seed:menu:resources"
	seedRequestIDMenuContent   = "init:seed:menu:content"
	seedRequestIDMenuBlogs     = "init:seed:menu:blogs"
)

/**
 * @description: 初始化管理员用户和角色策略
 * @param {*gorm.DB} db
 * @return {*}
 */
func InitAdmin(db *gorm.DB) {
	// 1. 确保管理员角色存在（动态 ID）
	var adminRole model.Role
	err := db.Where("keyword = ?", "role_admin").First(&adminRole).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		desc := "系统默认管理员角色"
		adminRole = model.Role{
			UUIDModel: model.UUIDModel{
				Version:   1,
				RequestId: RequestIDPtr(seedRequestIDRoleAdmin),
			},
			Name:    "管理员",
			Keyword: "role_admin",
			Desc:    &desc,
			Sort:    1,
			Status:  1,
			Creator: "系统",
		}
		if err := db.Create(&adminRole).Error; err != nil {
			Log.Errorf("创建默认管理员角色失败: %v", err)
			return
		}
	} else if err != nil {
		Log.Errorf("查询管理员角色失败: %v", err)
		return
	}
	ensureSeedRequestID(db, &model.Role{}, adminRole.ID, seedRequestIDRoleAdmin)

	// 管理员角色关联
	rolesForAssoc := []*model.Role{&adminRole}

	// 2. 创建默认管理员用户（如果不存在）
	var admin model.User
	result := db.Where("username = ?", "admin").First(&admin)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		Log.Errorf("查询管理员用户失败: %v", result.Error)
		return
	}
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		admin = model.User{
			UUIDModel: model.UUIDModel{
				Version:   1,
				RequestId: RequestIDPtr(seedRequestIDUserAdmin),
			},
			Username:     "admin",
			Password:     utils.GenPasswd("admin123"),
			Mobile:       "18888888888",
			Avatar:       "https://gips1.baidu.com/it/u=1397155327,2622199615&fm=3074&app=3074&f=PNG?w=2048&h=2048",
			Nickname:     "动物园园长",
			Introduction: "系统默认管理员",
			Status:       1,
			Creator:      "系统",
			Roles:        rolesForAssoc, // 管理员角色关联
		}
		if err := db.Create(&admin).Error; err != nil {
			Log.Errorf("创建管理员用户失败: %v", err)
			return
		}
	} else {
		ensureSeedRequestID(db, &model.User{}, admin.ID, seedRequestIDUserAdmin)
	}

	sysIconStr := "SettingOutlined"
	systemUserStr := "/system/users"
	usersIconStr := "UserOutlined"
	rolesIconStr := "TeamOutlined"
	menusIconStr := "AppstoreOutlined"
	contentIconStr := "ReadOutlined"
	contentBlogsStr := "/content/blogs"

	// 3. 菜单：父子关系使用上一节创建的菜单 ID
	systemMenu := model.Menu{
		UUIDModel: model.UUIDModel{
			Version:   1,
			RequestId: RequestIDPtr(seedRequestIDMenuSystem),
		},
		Name:     "system",
		Title:    "系统管理",
		Icon:     &sysIconStr,
		Path:     "system",
		Redirect: &systemUserStr,
		Sort:     10,
		ParentId: nil,
		Roles:    rolesForAssoc, // 管理员角色关联
		Creator:  "系统",
	}
	if err := findOrCreateMenu(db, &systemMenu, seedRequestIDMenuSystem); err != nil {
		Log.Errorf("初始化菜单 system 失败: %v", err)
		return
	}

	usersMenu := model.Menu{
		UUIDModel: model.UUIDModel{
			Version:   1,
			RequestId: RequestIDPtr(seedRequestIDMenuUsers),
		},
		Name:     "users",
		Title:    "用户管理",
		Icon:     &usersIconStr,
		Path:     "users",
		Sort:     11,
		ParentId: &systemMenu.ID,
		Roles:    rolesForAssoc, // 管理员角色关联
		Creator:  "系统",
	}
	if err := findOrCreateMenu(db, &usersMenu, seedRequestIDMenuUsers); err != nil {
		Log.Errorf("初始化菜单 users 失败: %v", err)
		return
	}

	rolesMenu := model.Menu{
		UUIDModel: model.UUIDModel{
			Version:   1,
			RequestId: RequestIDPtr(seedRequestIDMenuRoles),
		},
		Name:     "roles",
		Title:    "角色管理",
		Icon:     &rolesIconStr,
		Path:     "roles",
		Sort:     12,
		ParentId: &systemMenu.ID,
		Roles:    rolesForAssoc, // 管理员角色关联
		Creator:  "系统",
	}
	if err := findOrCreateMenu(db, &rolesMenu, seedRequestIDMenuRoles); err != nil {
		Log.Errorf("初始化菜单 roles 失败: %v", err)
		return
	}

	resourcesMenu := model.Menu{
		UUIDModel: model.UUIDModel{
			Version:   1,
			RequestId: RequestIDPtr(seedRequestIDMenuResources),
		},
		Name:     "resources",
		Title:    "资源管理",
		Icon:     &menusIconStr,
		Path:     "resources",
		Sort:     13,
		ParentId: &systemMenu.ID,
		Roles:    rolesForAssoc, // 管理员角色关联
		Creator:  "系统",
	}
	if err := findOrCreateMenu(db, &resourcesMenu, seedRequestIDMenuResources); err != nil {
		Log.Errorf("初始化菜单 resources 失败: %v", err)
		return
	}

	contentMenu := model.Menu{
		UUIDModel: model.UUIDModel{
			Version:   1,
			RequestId: RequestIDPtr(seedRequestIDMenuContent),
		},
		Name:     "content",
		Title:    "内容管理",
		Icon:     &contentIconStr,
		Path:     "content",
		Redirect: &contentBlogsStr,
		Type:     1,
		Sort:     20,
		ParentId: nil,
		Roles:    rolesForAssoc, // 管理员角色关联
		Creator:  "系统",
	}
	if err := findOrCreateMenu(db, &contentMenu, seedRequestIDMenuContent); err != nil {
		Log.Errorf("初始化菜单 content 失败: %v", err)
		return
	}

	blogsMenu := model.Menu{
		UUIDModel: model.UUIDModel{
			Version:   1,
			RequestId: RequestIDPtr(seedRequestIDMenuBlogs),
		},
		Name:     "blogs",
		Title:    "博客列表",
		Icon:     &contentIconStr,
		Path:     "blogs",
		Type:     2,
		Sort:     21,
		ParentId: &contentMenu.ID,
		Roles:    rolesForAssoc, // 管理员角色关联
		Creator:  "系统",
	}
	if err := findOrCreateMenu(db, &blogsMenu, seedRequestIDMenuBlogs); err != nil {
		Log.Errorf("初始化菜单 blogs 失败: %v", err)
		return
	}

	Log.Info("初始化管理员用户和角色策略完成！")
}

/**
 * @description: 创建或加载菜单，写回 m（含 ID）。
 * @param {*gorm.DB} db
 * @param {*model.Menu} m
 * @return {error}
 */
// ensureSeedRequestID 为迁移前已存在的种子记录补写 request_id / version，避免与 goose 回填不一致。
func ensureSeedRequestID(db *gorm.DB, tableModel interface{}, id string, requestID string) {
	if id == "" || requestID == "" {
		return
	}
	_ = db.Model(tableModel).Where("id = ? AND (request_id IS NULL OR request_id = '')", id).
		Updates(map[string]interface{}{
			"request_id": requestID,
			"version":    1,
		}).Error
}

func findOrCreateMenu(db *gorm.DB, m *model.Menu, seedRequestID string) error {
	var existing model.Menu
	err := db.Where("request_id = ?", seedRequestID).First(&existing).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = db.Where("name = ? AND path = ?", m.Name, m.Path).First(&existing).Error
	}
	if err == nil {
		*m = existing
		ensureSeedRequestID(db, &model.Menu{}, existing.ID, seedRequestID)
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if m.RequestId == nil {
		m.RequestId = RequestIDPtr(seedRequestID)
	}
	if m.Version == 0 {
		m.Version = 1
	}
	return db.Create(m).Error
}
