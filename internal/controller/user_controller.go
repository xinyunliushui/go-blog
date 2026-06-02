/*
 * @Date: 2026-03-25 22:08:27
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-05-15 22:05:18
 * @Description: 用户控制器接口实现
 */
package controller

import (
	"errors"
	"go-blog/internal/common"
	"go-blog/internal/config"
	"go-blog/internal/dto"
	"go-blog/internal/model"
	"go-blog/internal/repository"
	"go-blog/internal/response"
	"go-blog/internal/utils"
	"go-blog/internal/vo"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/thoas/go-funk"
)

type IUserController interface {
	GetUserInfo(ctx *gin.Context)    // 获取当前登录用户信息
	GetUsers(ctx *gin.Context)       // 获取用户列表
	CreateUser(ctx *gin.Context)     // 创建用户
	UpdateUserById(ctx *gin.Context) // 更新用户
	ChangePwd(ctx *gin.Context)      // 更新用户登录密码
}

type UserController struct {
	UserRepository repository.IUserRepository
}

func NewUserController() IUserController {
	return &UserController{
		UserRepository: repository.NewUserRepository(),
	}
}

/** 获取当前登录用户信息
 * @param ctx *gin.Context 上下文
 * @return void
 */
func (uc *UserController) GetUserInfo(ctx *gin.Context) {
	user, err := uc.UserRepository.GetCurrentUser(ctx)
	if err != nil {
		response.FailErr(ctx, nil, "获取用户信息失败", err)
		return
	}
	// 返回
	response.Success(ctx, dto.ToUserInfoDto(user), "获取用户信息成功")
}

/** 获取用户列表
 * @param ctx *gin.Context 上下文
 * @return void
 */
func (uc *UserController) GetUsers(ctx *gin.Context) {
	var req vo.UserListRequest
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
	users, total, err := uc.UserRepository.GetUsers(&req)
	if err != nil {
		response.FailErr(ctx, nil, "获取用户列表失败", err)
		return
	}
	response.Success(ctx, gin.H{"content": dto.ToUsersDto(users), "total": total, "page": req.Page, "pageSize": req.PageSize}, "获取用户列表成功")
}

/** 创建用户
 * @param ctx *gin.Context 上下文
 * @return void
 */
func (uc *UserController) CreateUser(ctx *gin.Context) {
	var req vo.CreateOrUpdateUserRequest
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
	// 创建用户
	var roles []*model.Role
	var err error
	if len(req.RoleIds) > 0 {
		rr := repository.NewRoleRepository()
		roles, err = rr.GetRolesByIds(req.RoleIds)
		if err != nil {
			response.FailErr(ctx, nil, "根据角色ID获取角色信息失败", err)
			return
		}
	}
	userStatus := req.Status
	if userStatus == 0 {
		userStatus = 1
	}
	user := &model.User{
		Username:     req.Username,
		Mobile:       req.Mobile,
		Avatar:       req.Avatar,
		Nickname:     req.Nickname,
		Introduction: req.Introduction,
		Status:       userStatus,
		Roles:        roles,
		Creator:      req.Username,
	}
	// 密码赋值
	if req.Password != "" {
		// 密码通过RSA解密
		decodeData, err := utils.RSADecrypt([]byte(req.Password), config.Conf.Application.RSAPrivateBytes)
		if err != nil {
			response.FailErr(ctx, nil, "密码解密失败", err)
			return
		}
		user.Password = utils.GenPasswd(string(decodeData))
	}
	err = uc.UserRepository.CreateUser(user)
	if err != nil {
		response.FailErr(ctx, nil, "创建用户失败", err)
		return
	}
	response.Success(ctx, nil, "创建用户成功")
}

/** 通过ID更新用户
 * @param ctx *gin.Context 上下文
 * @return void
 */
func (uc *UserController) UpdateUserById(ctx *gin.Context) {
	userId := strings.TrimSpace(ctx.Param("userId"))
	if userId == "" {
		response.Fail(ctx, nil, "用户ID不正确")
		return
	}
	var req vo.CreateOrUpdateUserRequest
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

	// 根据path中的userId获取用户信息
	oldUser, err := uc.UserRepository.GetUserById(userId)
	if err != nil {
		response.FailErr(ctx, nil, "获取需要更新的用户信息失败", err)
		return
	}

	// 获取当前用户
	ctxUser, err := uc.UserRepository.GetCurrentUser(ctx)
	if err != nil {
		response.FailErr(ctx, nil, "获取当前用户信息失败", err)
		return
	}
	// 获取当前用户的所有角色
	currentRoles := ctxUser.Roles
	// 获取当前用户角色的排序，和前端传来的角色排序做比较
	var currentRoleSorts []int
	// 当前用户角色ID集合
	var currentRoleIds []string
	var currentRoleSortMin int = 999 // 默认等级最低
	// 如果当前用户有角色，则获取当前用户角色排序最小值
	if len(currentRoles) != 0 {
		for _, role := range currentRoles {
			currentRoleSorts = append(currentRoleSorts, int(role.Sort))
			currentRoleIds = append(currentRoleIds, role.ID)
		}
		currentRoleSortMin = funk.MinInt(currentRoleSorts)
	}

	// 更新的角色信息-预期目标
	targetRoleIds := req.RoleIds
	var targetRoles []*model.Role
	var targetRoleSortMin int = 999 // 默认等级最低
	if len(targetRoleIds) != 0 {
		// 根据角色id获取角色
		rr := repository.NewRoleRepository()
		targetRoles, err = rr.GetRolesByIds(targetRoleIds)
		if err != nil {
			response.FailErr(ctx, nil, "获取用户角色信息失败", err)
			return
		}
		if len(targetRoles) == 0 {
			response.Fail(ctx, nil, "未获取到角色信息")
			return
		}
		var targetRoleSorts []int
		for _, role := range targetRoles {
			targetRoleSorts = append(targetRoleSorts, int(role.Sort))
		}
		// 前端传来用户角色排序最小值（最高等级角色）
		targetRoleSortMin = funk.MinInt(targetRoleSorts)
	}

	user := &model.User{
		UUIDModel:    oldUser.UUIDModel,
		Username:     req.Username,
		Password:     oldUser.Password,
		Mobile:       req.Mobile,
		Avatar:       req.Avatar,
		Nickname:     req.Nickname,
		Introduction: req.Introduction,
		Status:       req.Status,
		Creator:      ctxUser.Username,
	}

	// 判断是更新自己还是更新别人
	if userId == ctxUser.ID {
		// 如果是更新自己
		// 不能禁用自己
		if req.Status == 2 {
			response.Fail(ctx, nil, "不能禁用自己")
			return
		}
		// 不能变更自己的角色
		reqDiff, currentDiff := funk.DifferenceString(req.RoleIds, currentRoleIds)
		if len(reqDiff) > 0 || len(currentDiff) > 0 {
			// 对比有差异，则不能更改
			response.Fail(ctx, nil, "不能更改自己的角色")
			return
		}
		// 不能更新自己的密码，只能在个人中心更新
		if req.Password != "" {
			response.Fail(ctx, nil, "请到个人中心更新自身密码")
			return
		}
		// 角色信息沿用之前的，不做更新
		user.Roles = currentRoles
	} else {
		// 如果是更新别人
		// 根据path中的userId获取用户角色排序最小值
		minRoleSorts, err := uc.UserRepository.GetUserMinRoleSortsByIds([]string{userId})
		// 获取角色处理失败 非不是用户没角色
		if !errors.Is(err, repository.ErrUserNotAssignedRoles) {
			if err != nil || len(minRoleSorts) == 0 {
				response.FailErr(ctx, nil, "根据用户ID获取用户角色排序最小值失败", err)
				return
			}
			if currentRoleSortMin >= minRoleSorts[0] {
				response.Fail(ctx, nil, "用户不能更新比自己角色等级高的或者相同等级的用户")
				return
			}
		}
		if currentRoleSortMin >= targetRoleSortMin {
			response.Fail(ctx, nil, "用户不能把别的用户角色等级更新得比自己高或相等")
			return
		}
		user.Roles = targetRoles
	}

	// 更新用户
	err = uc.UserRepository.UpdateUserById(user)
	if err != nil {
		response.FailErr(ctx, nil, "更新用户失败", err)
		return
	}
	response.Success(ctx, nil, "更新用户成功")

}

/** 更新用户登录密码
 * @param ctx *gin.Context 上下文
 * @return void
 */
func (uc *UserController) ChangePwd(ctx *gin.Context) {
	var req vo.ChangePwdRequest

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

	// 前端传来的密码是rsa加密的,先解密
	// 密码通过RSA解密
	decodeOldPassword, err := utils.RSADecrypt([]byte(req.OldPassword), config.Conf.Application.RSAPrivateBytes)
	if err != nil {
		response.FailErr(ctx, nil, "密码解密失败", err)
		return
	}
	decodeNewPassword, err := utils.RSADecrypt([]byte(req.NewPassword), config.Conf.Application.RSAPrivateBytes)
	if err != nil {
		response.FailErr(ctx, nil, "密码解密失败", err)
		return
	}
	req.OldPassword = string(decodeOldPassword)
	req.NewPassword = string(decodeNewPassword)

	// 获取当前用户
	user, err := uc.UserRepository.GetCurrentUser(ctx)
	if err != nil {
		response.FailErr(ctx, nil, "获取当前用户信息失败", err)
		return
	}
	// 获取用户的真实正确密码
	correctPasswd := user.Password
	// 判断前端请求的密码是否等于真实密码
	err = utils.ComparePasswd(correctPasswd, req.OldPassword)
	if err != nil {
		response.FailErr(ctx, nil, "原密码有误", err)
		return
	}
	// 更新密码
	err = uc.UserRepository.ChangePwd(user.Username, utils.GenPasswd(req.NewPassword))
	if err != nil {
		response.FailErr(ctx, nil, "更新密码失败", err)
		return
	}
	response.Success(ctx, nil, "更新密码成功")
}
