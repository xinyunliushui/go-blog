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
	CreateMenu(menu *model.Menu) error                            // 创建菜单
	UpdateMenuById(menuId string, menu *model.Menu) error         // 更新菜单
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
 * @return error
 */
func (mr *MenuRepository) CreateMenu(menu *model.Menu) error {
	err := common.DB.Create(menu).Error
	return err
}

/** 通过ID更新菜单
 * @param menuId string 菜单ID
 * @param menu *model.Menu 菜单
 * @return error
 */
func (mr *MenuRepository) UpdateMenuById(menuId string, menu *model.Menu) error {
	err := common.DB.Model(&model.Menu{}).Where("id = ?", menuId).Omit("ID").Updates(menu).Error
	return err
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
