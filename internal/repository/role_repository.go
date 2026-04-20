/*
 * @Date: 2026-04-02 22:10:45
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-20 15:25:30
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
}

type RoleRepository struct{}

func NewRoleRepository() IRoleRepository {
	return RoleRepository{}
}

func (r RoleRepository) GetRoles(req *vo.RoleListRequest) ([]*model.Role, int64, error) {
	var list []*model.Role
	db := common.DB.Model(&model.Role{}).Order("created_at DESC")

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
