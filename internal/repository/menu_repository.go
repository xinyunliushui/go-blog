/*
 * @Date: 2026-04-08 21:28:24
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-05-15 15:04:37
 * @Description: 菜单仓库
 */
package repository

import (
	"go-blog/internal/common"
	"go-blog/internal/dto"
	"go-blog/internal/model"

	"github.com/thoas/go-funk"
)

type IMenuRepository interface {
	GetMenus() ([]*model.Menu, error)                             // 获取菜单列表
	GetMenuTree() ([]dto.MenuDto, error)                          // 获取菜单树
	CreateMenu(menu *model.Menu) (duplicate bool, err error)                            // 创建菜单
	UpdateMenuById(menuId string, version uint, menu *model.Menu) (duplicate bool, err error) // 更新菜单
	GetUserMenuTreeByUserId(userId string) ([]dto.MenuDto, error) // 根据用户ID获取用户的权限(可访问)菜单树
	GetUserMenusByUserId(userId string) ([]*model.Menu, error)    // 根据用户ID获取用户的权限(可访问)菜单列表
}

type MenuRepository struct {
}

func NewMenuRepository() IMenuRepository {
	return &MenuRepository{}
}

/** 根据用户ID获取用户的权限(可访问)菜单列表
 * @param userId string 用户ID
 * @return []*model.Menu, error
 */
func (mr *MenuRepository) GetUserMenusByUserId(userId string) ([]*model.Menu, error) {
	// 获取用户
	var user model.User
	err := common.DB.Where("id = ?", userId).Preload("Roles").First(&user).Error
	if err != nil {
		return nil, err
	}
	// 获取角色
	roles := user.Roles
	// 所有角色的菜单集合
	allRoleMenus := make([]*model.Menu, 0)
	for _, role := range roles {
		var userRole model.Role
		if err := common.DB.Where("id = ?", role.ID).Preload("Menus").First(&userRole).Error; err != nil {
			return nil, err
		}
		// 获取角色的菜单
		menus := userRole.Menus
		allRoleMenus = append(allRoleMenus, menus...)
	}

	// 所有角色的菜单集合去重
	allRoleMenusId := make([]string, 0)
	for _, menu := range allRoleMenus {
		allRoleMenusId = append(allRoleMenusId, menu.ID)
	}
	allRoleMenusIdUniq := funk.UniqString(allRoleMenusId)
	allRoleMenusUniq := make([]*model.Menu, 0)
	for _, id := range allRoleMenusIdUniq {
		for _, menu := range allRoleMenus {
			if id == menu.ID {
				allRoleMenusUniq = append(allRoleMenusUniq, menu)
				break
			}
		}
	}

	// 获取状态status为1的菜单
	accessMenus := make([]*model.Menu, 0)
	for _, menu := range allRoleMenusUniq {
		if menu.Status == 1 {
			accessMenus = append(accessMenus, menu)
		}
	}

	return accessMenus, nil
}

/** 根据用户ID获取用户的权限(可访问)菜单树
 * @param userId string 用户ID
 * @return []dto.MenuDto, error
 */
func (mr *MenuRepository) GetUserMenuTreeByUserId(userId string) ([]dto.MenuDto, error) {
	menus, err := mr.GetUserMenusByUserId(userId)
	if err != nil {
		return nil, err
	}
	tree := GenMenuTreeDto("", menus)
	return tree, err
}

/** 创建菜单树
 * @param parentKey 父菜单 ID，空字符串表示根
 * @param menus []*model.Menu 菜单列表
 * @return []dto.MenuDto
 */
func GenMenuTreeDto(parentKey string, menus []*model.Menu) []dto.MenuDto {
	tree := make([]dto.MenuDto, 0)
	for _, m := range menus {
		if menuBelongsUnderParent(m, parentKey) {
			newMenu := dto.ToMenuDto(m)
			newMenu.Children = GenMenuTreeDto(m.ID, menus)
			tree = append(tree, newMenu)
		}
	}
	return tree
}

/** 获取菜单列表
 * @return []*model.Menu, error
 */
func (mr *MenuRepository) GetMenus() ([]*model.Menu, error) {
	var menus []*model.Menu
	err := common.DB.Order("sort").Find(&menus).Error
	return menus, err
}

/** 获取菜单树
 * @return []dto.MenuDto, error
 */
func (mr *MenuRepository) GetMenuTree() ([]dto.MenuDto, error) {
	menus, err := mr.GetMenus()
	if err != nil {
		return nil, err
	}
	tree := GenMenuTreeDto("", menus)
	return tree, nil
}

/** 创建菜单
 * @param menu *model.Menu 菜单
 * @return duplicate, error
 */
func (mr *MenuRepository) CreateMenu(menu *model.Menu) (bool, error) {
	intended := *menu
	err := common.DB.Create(menu).Error
	requestID := ""
	if menu.RequestId != nil {
		requestID = *menu.RequestId
	}
	return handleCreateIdempotency(common.DB, requestID, err, func() (bool, error) {
		var existing model.Menu
		if loadErr := common.DB.Where("request_id = ?", requestID).First(&existing).Error; loadErr != nil {
			return false, err
		}
		if !menuFieldsMatch(&existing, &intended) {
			return false, nil
		}
		*menu = existing
		return true, nil
	})
}

func menuFieldsMatch(current, target *model.Menu) bool {
	return current.Name == target.Name &&
		current.Title == target.Title &&
		common.StringPtrEqual(current.Icon, target.Icon) &&
		current.Path == target.Path &&
		common.StringPtrEqual(current.Redirect, target.Redirect) &&
		current.Sort == target.Sort &&
		current.Status == target.Status &&
		current.Type == target.Type &&
		common.StringPtrEqual(current.ParentId, target.ParentId)
}

/** 通过ID更新菜单
 * @param menuId string 菜单ID
 * @param version 乐观锁版本号
 * @param menu *model.Menu 菜单
 * @return duplicate, error
 */
func (mr *MenuRepository) UpdateMenuById(menuId string, version uint, menu *model.Menu) (bool, error) {
	updates := map[string]interface{}{
		"name":     menu.Name,
		"title":    menu.Title,
		"icon":     menu.Icon,
		"path":     menu.Path,
		"redirect": menu.Redirect,
		"sort":     menu.Sort,
		"status":   menu.Status,
		"type":     menu.Type,
		"parent_id": menu.ParentId,
		"creator":  menu.Creator,
		"version":  version + 1,
	}
	result := common.DB.Model(&model.Menu{}).Where("id = ? AND version = ?", menuId, version).Updates(updates)
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected > 0 {
		return false, nil
	}
	var current model.Menu
	if err := common.DB.Where("id = ?", menuId).First(&current).Error; err != nil {
		return false, err
	}
	return applyOptimisticLockResult(menuFieldsMatch(&current, menu), result.RowsAffected, nil)
}

/** 判断菜单是否属于父菜单
 * @param m *model.Menu 菜单
 * @param parentKey string 父菜单ID
 * @return bool
 */
func menuBelongsUnderParent(m *model.Menu, parentKey string) bool {
	if parentKey == "" {
		return m.ParentId == nil || *m.ParentId == ""
	}
	if m.ParentId == nil {
		return false
	}
	return *m.ParentId == parentKey
}
