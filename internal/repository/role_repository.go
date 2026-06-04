/*
 * @Date: 2026-04-02 22:10:45
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-05-15
 * @Description: 角色仓库
 */
package repository

import (
	"go-blog/internal/common"
	"go-blog/internal/model"
	"go-blog/internal/vo"

	"gorm.io/gorm"
)

type IRoleRepository interface {
	GetRoles(req *vo.RoleListRequest) ([]*model.Role, int64, error) // 获取角色列表
	GetRolesByIds(roleIds []string) ([]*model.Role, error)          // 根据角色ID获取角色
	GetRoleMenusById(roleId string) ([]*model.Menu, error)          // 获取角色的权限菜单
	CreateRole(role *model.Role) (duplicate bool, err error)                              // 创建角色
	UpdateRoleById(roleId string, version uint, role *model.Role) (duplicate bool, err error) // 更新角色
	UpdateRoleMenus(role *model.Role, version uint) (duplicate bool, err error)           // 更新角色的权限菜单
}

type RoleRepository struct{}

func NewRoleRepository() IRoleRepository {
	return &RoleRepository{}
}

/** 获取角色列表
 * @param req *vo.RoleListRequest 角色列表请求
 * @return []*model.Role, int64, error
 */
func (*RoleRepository) GetRoles(req *vo.RoleListRequest) ([]*model.Role, int64, error) {
	var list []*model.Role
	db := common.DB.Model(&model.Role{}).Preload("Menus").Order("sort ASC")

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

/** 根据角色ID获取角色
 * @param roleIds []string 角色ID列表
 * @return []*model.Role, error
 */
func (*RoleRepository) GetRolesByIds(roleIds []string) ([]*model.Role, error) {
	var list []*model.Role
	err := common.DB.Where("id IN (?)", roleIds).Find(&list).Error
	return list, err
}

/** 获取角色的权限菜单
 * @param roleId string 角色ID
 * @return []*model.Menu, error
 */
func (*RoleRepository) GetRoleMenusById(roleId string) ([]*model.Menu, error) {
	var role model.Role
	err := common.DB.Where("id = ?", roleId).Preload("Menus").First(&role).Error
	return role.Menus, err
}

/** 创建角色
 * @param role *model.Role 角色
 * @return duplicate, error
 */
func (*RoleRepository) CreateRole(role *model.Role) (bool, error) {
	intended := *role
	err := common.DB.Create(role).Error
	requestID := ""
	if role.RequestId != nil {
		requestID = *role.RequestId
	}
	return handleCreateIdempotency(common.DB, requestID, err, func() (bool, error) {
		var existing model.Role
		if loadErr := common.DB.Where("request_id = ?", requestID).First(&existing).Error; loadErr != nil {
			return false, err
		}
		if !roleFieldsMatch(&existing, &intended) {
			return false, nil
		}
		*role = existing
		return true, nil
	})
}

func roleFieldsMatch(current, target *model.Role) bool {
	return current.Name == target.Name &&
		current.Keyword == target.Keyword &&
		common.StringPtrEqual(current.Desc, target.Desc) &&
		current.Status == target.Status &&
		current.Sort == target.Sort
}

/** 通过ID更新角色
 * @param roleId string 角色ID
 * @param version 乐观锁版本号
 * @param role *model.Role 角色
 * @return duplicate, error
 */
func (*RoleRepository) UpdateRoleById(roleId string, version uint, role *model.Role) (bool, error) {
	updates := map[string]interface{}{
		"name":    role.Name,
		"keyword": role.Keyword,
		"desc":    role.Desc,
		"status":  role.Status,
		"sort":    role.Sort,
		"creator": role.Creator,
		"version": version + 1,
	}
	result := common.DB.Model(&model.Role{}).Where("id = ? AND version = ?", roleId, version).Updates(updates)
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected > 0 {
		return false, nil
	}
	var current model.Role
	if err := common.DB.Where("id = ?", roleId).First(&current).Error; err != nil {
		return false, err
	}
	return applyOptimisticLockResult(roleFieldsMatch(&current, role), result.RowsAffected, nil)
}

/** 更新角色的权限菜单
 * @param role *model.Role 角色
 * @param version 乐观锁版本号
 * @return duplicate, error
 */
func (*RoleRepository) UpdateRoleMenus(role *model.Role, version uint) (bool, error) {
	targetMenuIDs := make([]string, 0, len(role.Menus))
	for _, menu := range role.Menus {
		targetMenuIDs = append(targetMenuIDs, menu.ID)
	}

	var duplicate bool
	err := common.Transaction(func(gdb *gorm.DB) error {
		result := gdb.Model(&model.Role{}).Where("id = ? AND version = ?", role.ID, version).
			Update("version", version+1)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			var current model.Role
			if err := gdb.Where("id = ?", role.ID).Preload("Menus").First(&current).Error; err != nil {
				return err
			}
			currentMenuIDs := make([]string, 0, len(current.Menus))
			for _, menu := range current.Menus {
				currentMenuIDs = append(currentMenuIDs, menu.ID)
			}
			if common.StringSliceEqualUnordered(currentMenuIDs, targetMenuIDs) {
				duplicate = true
				*role = current
				return nil
			}
			return common.ErrOptimisticLockConflict
		}
		if err := gdb.Model(role).Association("Menus").Replace(role.Menus); err != nil {
			return err
		}
		role.Version = version + 1
		return nil
	})
	return duplicate, err
}
