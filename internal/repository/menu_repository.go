/*
 * @Date: 2026-04-08 21:28:24
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-13 15:18:04
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
	GetMenus() ([]*model.Menu, error)                           // 获取菜单列表
	GetMenuTree() ([]dto.MenuDto, error)                        // 获取菜单树
	CreateMenu(menu *model.Menu) error                          // 创建菜单
	UpdateMenuById(menuId uint, menu *model.Menu) error         // 更新菜单
	GetUserMenuTreeByUserId(userId uint) ([]dto.MenuDto, error) // 根据用户ID获取用户的权限(可访问)菜单树
}

type MenuRepository struct {
}

func NewMenuRepository() IMenuRepository {
	return MenuRepository{}
}

// 根据用户ID获取用户的权限(可访问)菜单列表
func (m MenuRepository) GetUserMenusByUserId(userId uint) ([]*model.Menu, error) {
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
		err := common.DB.Where("id = ?", role.ID).Preload("Menus").First(&userRole).Error
		if err != nil {
			return nil, err
		}
		// 获取角色的菜单
		menus := userRole.Menus
		allRoleMenus = append(allRoleMenus, menus...)
	}

	// 所有角色的菜单集合去重
	allRoleMenusId := make([]int, 0)
	for _, menu := range allRoleMenus {
		allRoleMenusId = append(allRoleMenusId, int(menu.ID))
	}
	allRoleMenusIdUniq := funk.UniqInt(allRoleMenusId)
	allRoleMenusUniq := make([]*model.Menu, 0)
	for _, id := range allRoleMenusIdUniq {
		for _, menu := range allRoleMenus {
			if id == int(menu.ID) {
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

	return accessMenus, err
}

// 根据用户ID获取用户的权限(可访问)菜单树
func (m MenuRepository) GetUserMenuTreeByUserId(userId uint) ([]dto.MenuDto, error) {
	menus, err := m.GetUserMenusByUserId(userId)
	if err != nil {
		return nil, err
	}
	tree := GenMenuTreeDto(0, menus)
	return tree, err
}

// 创建菜单树
func GenMenuTreeDto(parentId uint, menus []*model.Menu) []dto.MenuDto {
	tree := make([]dto.MenuDto, 0)
	for _, m := range menus {
		if *m.ParentId == parentId {
			newMenu := dto.ToMenuDto(m)
			children := GenMenuTreeDto(m.ID, menus)
			newMenu.Children = children
			tree = append(tree, newMenu)
		}
	}
	return tree
}

func (m MenuRepository) GetMenus() ([]*model.Menu, error) {
	var menus []*model.Menu
	err := common.DB.Order("sort").Find(&menus).Error
	return menus, err
}

func (m MenuRepository) GetMenuTree() ([]dto.MenuDto, error) {
	menus, err := m.GetMenus()
	tree := GenMenuTreeDto(0, menus)
	return tree, err
}

func (m MenuRepository) CreateMenu(menu *model.Menu) error {
	err := common.DB.Create(menu).Error
	return err
}

func (m MenuRepository) UpdateMenuById(menuId uint, menu *model.Menu) error {
	err := common.DB.Model(&model.Menu{}).Where("id = ?", menuId).Updates(menu).Error
	return err
}
