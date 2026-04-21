/*
 * @Date: 2026-04-02 22:10:45
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-21 09:30:01
 * @Description:
 */
package repository

import (
	"go-blog/internal/common"
	"go-blog/internal/model"
	"go-blog/internal/vo"
)

type IRoleRepository interface {
	GetRoles(req *vo.RoleListRequest) ([]*model.Role, int64, error) // 获取角色列表
	GetRolesByIds(roleIds []uint) ([]*model.Role, error)            // 根据角色ID获取角色
	GetRoleMenusById(roleId uint) ([]*model.Menu, error)            // 获取角色的权限菜单
	CreateRole(role *model.Role) error                              // 创建角色
	UpdateRoleById(roleId uint, role *model.Role) error             // 更新角色
	UpdateRoleMenus(role *model.Role) error                         // 更新角色的权限菜单
}

type RoleRepository struct{}

func NewRoleRepository() IRoleRepository {
	return RoleRepository{}
}

func (r RoleRepository) GetRoles(req *vo.RoleListRequest) ([]*model.Role, int64, error) {
	var list []*model.Role
	db := common.DB.Model(&model.Role{}).Order("sort ASC")

	total := int64(0)
	err := db.Count(&total).Error
	if err != nil {
		return list, total, err
	}

	page := int(req.Page)
	pageSize := int(req.PageSize)
	if page > 0 && pageSize > 0 {
		err = db.Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	} else {
		err = db.Find(&list).Error
	}
	return list, total, err
}

// 根据角色ID获取角色
func (r RoleRepository) GetRolesByIds(roleIds []uint) ([]*model.Role, error) {
	var list []*model.Role
	err := common.DB.Where("id IN (?)", roleIds).Find(&list).Error
	return list, err
}

// 获取角色的权限菜单
func (r RoleRepository) GetRoleMenusById(roleId uint) ([]*model.Menu, error) {
	var role model.Role
	err := common.DB.Where("id = ?", roleId).Preload("Menus").First(&role).Error
	return role.Menus, err
}

// 创建角色
func (r RoleRepository) CreateRole(role *model.Role) error {
	err := common.DB.Create(role).Error
	return err
}

// 更新角色
func (r RoleRepository) UpdateRoleById(roleId uint, role *model.Role) error {
	err := common.DB.Model(&model.Role{}).Where("id = ?", roleId).Updates(role).Error
	return err
}

// 更新角色的权限菜单
func (r RoleRepository) UpdateRoleMenus(role *model.Role) error {
	err := common.DB.Model(role).Association("Menus").Replace(role.Menus)
	return err
}
