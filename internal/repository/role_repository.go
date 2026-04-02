package repository

import (
	"go-blog/internal/common"
	"go-blog/internal/model"
	"go-blog/internal/vo"
)

type IRoleRepository interface {
	GetRoles(req *vo.RoleListRequest) ([]*model.Role, int64, error) // 获取角色列表
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
