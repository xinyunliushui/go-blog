/*
 * @Date: 2026-04-08 21:27:02
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-05-15
 * @Description: 菜单控制器接口实现
 */
package controller

import (
	"go-blog/internal/common"
	"go-blog/internal/dto"
	"go-blog/internal/model"
	"go-blog/internal/repository"
	"go-blog/internal/response"
	"go-blog/internal/vo"
	"strings"

	"github.com/gin-gonic/gin"
)

type IMenuController interface {
	GetMenus(c *gin.Context)                // 获取菜单列表
	GetMenuTree(c *gin.Context)             // 获取菜单树
	CreateMenu(c *gin.Context)              // 创建菜单
	UpdateMenuById(c *gin.Context)          // 更新菜单
	GetUserMenuTreeByUserId(c *gin.Context) // 获取用户的可访问菜单树
}

type MenuController struct {
	MenuRepository repository.IMenuRepository
}

func NewMenuController() IMenuController {
	return &MenuController{
		MenuRepository: repository.NewMenuRepository(),
	}
}

/** 根据用户ID获取用户的可访问菜单树
 * @param c *gin.Context 上下文
 * @return void
 */
func (mc *MenuController) GetUserMenuTreeByUserId(c *gin.Context) {
	userId := strings.TrimSpace(c.Param("userId"))
	if userId == "" {
		response.Fail(c, nil, "用户ID不正确")
		return
	}

	menuTree, err := mc.MenuRepository.GetUserMenuTreeByUserId(userId)
	if err != nil {
		response.Fail(c, nil, "获取用户的可访问菜单树失败: "+err.Error())
		return
	}
	response.Success(c, menuTree, "获取用户的可访问菜单树成功")
}

/** 获取菜单列表
 * @param c *gin.Context 上下文
 * @return void
 */
func (mc *MenuController) GetMenus(c *gin.Context) {
	menus, err := mc.MenuRepository.GetMenus()
	if err != nil {
		response.Fail(c, nil, "获取菜单列表失败: "+err.Error())
		return
	}
	response.Success(c, dto.ToMenuListDto(menus), "获取菜单列表成功")
}

/** 创建菜单
 * @param c *gin.Context 上下文
 * @return void
 */
func (mc *MenuController) CreateMenu(c *gin.Context) {
	var req vo.CreateMenuRequest
	// 参数绑定
	if err := c.ShouldBind(&req); err != nil {
		response.Fail(c, nil, common.ValidationErrString(err))
		return
	}
	// 参数校验
	if err := common.Validate.Struct(&req); err != nil {
		response.Fail(c, nil, common.ValidationErrString(err))
		return
	}

	// 获取当前用户
	ur := repository.NewUserRepository()
	ctxUser, err := ur.GetCurrentUser(c)
	if err != nil {
		response.Fail(c, nil, "获取当前用户信息失败")
		return
	}

	var parentPtr *string
	if pid := strings.TrimSpace(req.ParentId); pid != "" {
		parentPtr = &pid
	}

	menu := model.Menu{
		Name:     req.Name,
		Title:    req.Title,
		Icon:     &req.Icon,
		Path:     req.Path,
		Redirect: &req.Redirect,
		Sort:     req.Sort,
		Status:   req.Status,
		Type:     req.Type,
		ParentId: parentPtr,
		Creator:  ctxUser.Username,
	}

	err = mc.MenuRepository.CreateMenu(&menu)
	if err != nil {
		response.Fail(c, nil, "创建菜单失败: "+err.Error())
		return
	}
	response.Success(c, nil, "创建菜单成功")
}

/** 获取应用下的菜单列表
 * @param c *gin.Context 上下文
 * @return void
 */
func (mc *MenuController) GetMenuTree(c *gin.Context) {
	menuTree, err := mc.MenuRepository.GetMenuTree()
	if err != nil {
		response.Fail(c, nil, "获取菜单树失败: "+err.Error())
		return
	}
	response.Success(c, menuTree, "获取菜单树成功")
}

/** 通过ID更新菜单
 * @param c *gin.Context 上下文
 * @return void
 */
func (mc *MenuController) UpdateMenuById(c *gin.Context) {
	menuId := strings.TrimSpace(c.Param("menuId"))
	if menuId == "" {
		response.Fail(c, nil, "菜单ID不正确")
		return
	}
	var req vo.CreateMenuRequest
	// 参数绑定
	if err := c.ShouldBind(&req); err != nil {
		response.Fail(c, nil, common.ValidationErrString(err))
		return
	}
	// 参数校验
	if err := common.Validate.Struct(&req); err != nil {
		response.Fail(c, nil, common.ValidationErrString(err))
		return
	}
	// 获取当前用户
	ur := repository.NewUserRepository()
	ctxUser, err := ur.GetCurrentUser(c)
	if err != nil {
		response.Fail(c, nil, "获取当前用户信息失败")
		return
	}

	var parentPtr *string
	if pid := strings.TrimSpace(req.ParentId); pid != "" {
		parentPtr = &pid
	}

	menu := model.Menu{
		Name:     req.Name,
		Title:    req.Title,
		Icon:     &req.Icon,
		Path:     req.Path,
		Redirect: &req.Redirect,
		Sort:     req.Sort,
		Status:   req.Status,
		Type:     req.Type,
		ParentId: parentPtr,
		Creator:  ctxUser.Username,
	}

	err = mc.MenuRepository.UpdateMenuById(menuId, &menu)
	if err != nil {
		response.Fail(c, nil, "更新菜单失败: "+err.Error())
		return
	}
	response.Success(c, nil, "更新菜单成功")
}
