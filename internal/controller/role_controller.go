/*
 * @Date: 2026-04-02 22:09:11
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-05-13 15:12:41
 * @Description: 角色控制器接口实现
 */
package controller

import (
	"fmt"
	"go-blog/internal/common"
	"go-blog/internal/dto"
	"go-blog/internal/model"
	"go-blog/internal/repository"
	"go-blog/internal/response"
	"go-blog/internal/vo"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/thoas/go-funk"
)

type IRoleController interface {
	GetRoles(ctx *gin.Context)
	CreateRole(c *gin.Context)          // 创建角色
	UpdateRoleById(c *gin.Context)      // 更新角色
	GetRoleMenusById(c *gin.Context)    // 获取角色的权限菜单
	UpdateRoleMenusById(c *gin.Context) // 更新角色的权限菜单
}

type RoleController struct {
	RoleRepository repository.IRoleRepository
}

func NewRoleController() IRoleController {
	return &RoleController{
		RoleRepository: repository.NewRoleRepository(),
	}
}

/** 获取角色列表
 * @param ctx *gin.Context 上下文
 * @return void
 */
func (rc *RoleController) GetRoles(ctx *gin.Context) {
	var req vo.RoleListRequest
	// 参数绑定
	if err := ctx.ShouldBind(&req); err != nil {
		response.Fail(ctx, nil, common.ValidationErrString(err))
		return
	}
	// 参数校验
	if err := common.Validate.Struct(&req); err != nil {
		response.Fail(ctx, nil, common.ValidationErrString(err))
		return
	}

	// 获取角色列表
	roles, total, err := rc.RoleRepository.GetRoles(&req)
	if err != nil {
		response.Fail(ctx, nil, "获取角色列表失败：")
		return
	}
	response.Success(ctx, gin.H{"content": dto.ToRolesDto(roles), "total": total, "page": req.Page, "pageSize": req.PageSize}, "获取角色列表成功")
}

/** 创建角色
 * @param ctx *gin.Context 上下文
 * @return void
 */
func (rc *RoleController) CreateRole(ctx *gin.Context) {
	var req vo.CreateRoleRequest
	// 参数绑定
	if err := ctx.ShouldBind(&req); err != nil {
		response.Fail(ctx, nil, common.ValidationErrString(err))
		return
	}
	// 参数校验
	if err := common.Validate.Struct(&req); err != nil {
		response.Fail(ctx, nil, common.ValidationErrString(err))
		return
	}

	// 获取当前用户最高角色等级
	uc := repository.NewUserRepository()
	sort, ctxUser, err := uc.GetCurrentUserMinRoleSort(ctx)
	if err != nil {
		response.Fail(ctx, nil, "获取当前用户最高角色等级失败: ")
		return
	}

	// 用户不能创建比自己等级高或相同等级的角色
	if sort >= req.Sort {
		response.Fail(ctx, nil, "不能创建比自己等级高或相同等级的角色")
		return
	}

	role := model.Role{
		Name:    req.Name,
		Keyword: req.Keyword,
		Desc:    &req.Desc,
		Status:  req.Status,
		Sort:    req.Sort,
		Creator: ctxUser.Username,
	}

	// 创建角色
	err = rc.RoleRepository.CreateRole(&role)
	if err != nil {
		response.Fail(ctx, nil, "创建角色失败: ")
		return
	}
	response.Success(ctx, nil, "创建角色成功")

}

/** 通过ID更新角色
 * @param ctx *gin.Context 上下文
 * @return void
 */
func (rc *RoleController) UpdateRoleById(ctx *gin.Context) {
	var req vo.CreateRoleRequest
	// 参数绑定
	if err := ctx.ShouldBind(&req); err != nil {
		response.Fail(ctx, nil, common.ValidationErrString(err))
		return
	}
	// 参数校验
	if err := common.Validate.Struct(&req); err != nil {
		response.Fail(ctx, nil, common.ValidationErrString(err))
		return
	}
	// 获取path中的roleId
	roleId, _ := strconv.Atoi(ctx.Param("roleId"))
	if roleId <= 0 {
		response.Fail(ctx, nil, "角色ID不正确")
		return
	}

	// 当前用户角色排序最小值（最高等级角色）以及当前用户
	ur := repository.NewUserRepository()
	minSort, ctxUser, err := ur.GetCurrentUserMinRoleSort(ctx)
	if err != nil {
		response.Fail(ctx, nil, err.Error())
		return
	}

	// 不能更新比自己角色等级高或相等的角色
	// 根据path中的角色ID获取该角色信息
	roles, err := rc.RoleRepository.GetRolesByIds([]uint{uint(roleId)})
	if err != nil {
		response.Fail(ctx, nil, err.Error())
		return
	}
	if len(roles) == 0 {
		response.Fail(ctx, nil, "未获取到角色信息")
		return
	}
	if minSort >= roles[0].Sort {
		response.Fail(ctx, nil, "不能更新比自己角色等级高或相等的角色")
		return
	}

	// 不能把角色等级更新得比当前用户的等级高
	if minSort >= req.Sort {
		response.Fail(ctx, nil, "不能把角色等级更新得比当前用户的等级高或相同")
		return
	}

	role := model.Role{
		Name:    req.Name,
		Keyword: req.Keyword,
		Desc:    &req.Desc,
		Status:  req.Status,
		Sort:    req.Sort,
		Creator: ctxUser.Username,
	}

	// 更新角色
	err = rc.RoleRepository.UpdateRoleById(uint(roleId), &role)
	if err != nil {
		response.Fail(ctx, nil, "更新角色失败: ")
		return
	}

	response.Success(ctx, nil, "更新角色成功")
}

/** 获取角色的权限菜单
 * @param ctx *gin.Context 上下文
 * @return void
 */
func (rc *RoleController) GetRoleMenusById(ctx *gin.Context) {
	roleId, _ := strconv.Atoi(ctx.Param("roleId"))
	if roleId <= 0 {
		response.Fail(ctx, nil, "角色ID不正确")
		return
	}
	menus, err := rc.RoleRepository.GetRoleMenusById(uint(roleId))
	if err != nil {
		response.Fail(ctx, nil, "获取角色的权限菜单失败: "+err.Error())
		return
	}
	response.Success(ctx, gin.H{"menus": menus}, "获取角色的权限菜单成功")
}

/** 更新角色的权限菜单
 * @param ctx *gin.Context 上下文
 * @return void
 */
func (rc *RoleController) UpdateRoleMenusById(ctx *gin.Context) {
	var req vo.UpdateRoleMenusRequest
	// 参数绑定
	if err := ctx.ShouldBind(&req); err != nil {
		response.Fail(ctx, nil, common.ValidationErrString(err))
		return
	}
	// 参数校验
	if err := common.Validate.Struct(&req); err != nil {
		response.Fail(ctx, nil, common.ValidationErrString(err))
		return
	}
	// 获取path中的roleId
	roleId, _ := strconv.Atoi(ctx.Param("roleId"))
	if roleId <= 0 {
		response.Fail(ctx, nil, "角色ID不正确")
		return
	}
	// 根据path中的角色ID获取该角色信息
	roles, err := rc.RoleRepository.GetRolesByIds([]uint{uint(roleId)})
	if err != nil {
		response.Fail(ctx, nil, err.Error())
		return
	}
	if len(roles) == 0 {
		response.Fail(ctx, nil, "未获取到角色信息")
		return
	}

	// 当前用户角色排序最小值（最高等级角色）以及当前用户
	ur := repository.NewUserRepository()
	minSort, ctxUser, err := ur.GetCurrentUserMinRoleSort(ctx)
	if err != nil {
		response.Fail(ctx, nil, err.Error())
		return
	}

	// (非管理员)不能更新比自己角色等级高或相等角色的权限菜单
	if minSort != 1 {
		if minSort >= roles[0].Sort {
			response.Fail(ctx, nil, "不能更新比自己角色等级高或相等角色的权限菜单")
			return
		}
	}

	// 获取当前用户所拥有的权限菜单
	mr := repository.NewMenuRepository()
	ctxUserMenus, err := mr.GetUserMenusByUserId(ctxUser.ID)
	if err != nil {
		response.Fail(ctx, nil, "获取当前用户的可访问菜单列表失败: "+err.Error())
		return
	}

	// 获取当前用户所拥有的权限菜单ID
	ctxUserMenusIds := make([]uint, 0)
	for _, menu := range ctxUserMenus {
		ctxUserMenusIds = append(ctxUserMenusIds, menu.ID)
	}

	// 前端传来最新的MenuIds集合
	menuIds := req.MenuIds

	// 用户需要修改的菜单集合
	reqMenus := make([]*model.Menu, 0)

	// (非管理员)不能把角色的权限菜单设置的比当前用户所拥有的权限菜单多
	if minSort != 1 {
		for _, id := range menuIds {
			if !funk.Contains(ctxUserMenusIds, id) {
				response.Fail(ctx, nil, fmt.Sprintf("无权设置ID为%d的菜单", id))
				return
			}
		}

		for _, id := range menuIds {
			for _, menu := range ctxUserMenus {
				if id == menu.ID {
					reqMenus = append(reqMenus, menu)
					break
				}
			}
		}
	} else {
		// 管理员随意设置
		// 根据menuIds查询查询菜单
		menus, err := mr.GetMenus()
		if err != nil {
			response.Fail(ctx, nil, "获取菜单列表失败: "+err.Error())
			return
		}
		for _, menuId := range menuIds {
			for _, menu := range menus {
				if menuId == menu.ID {
					reqMenus = append(reqMenus, menu)
				}
			}
		}
	}

	roles[0].Menus = reqMenus

	err = rc.RoleRepository.UpdateRoleMenus(roles[0])
	if err != nil {
		response.Fail(ctx, nil, "更新角色的权限菜单失败: "+err.Error())
		return
	}

	response.Success(ctx, nil, "更新角色的权限菜单成功")

}
