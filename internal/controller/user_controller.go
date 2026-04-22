/*
 * @Date: 2026-03-25 22:08:27
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-21 17:27:21
 * @Description: user controller
 */
package controller

import (
	"go-blog/internal/common"
	"go-blog/internal/config"
	"go-blog/internal/dto"
	"go-blog/internal/model"
	"go-blog/internal/repository"
	"go-blog/internal/response"
	"go-blog/internal/utils"
	"go-blog/internal/vo"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/thoas/go-funk"
)

type IUserController interface {
	GetUserInfo(ctx *gin.Context)
	GetUsers(ctx *gin.Context)
	CreateUser(ctx *gin.Context)
	UpdateUserById(ctx *gin.Context)
	ChangePwd(ctx *gin.Context)
}

type UserController struct {
	UserRepository repository.IUserRepository
}

// 构造函数
func NewUserController() IUserController {
	return UserController{
		UserRepository: repository.NewUserRepository(),
	}
}

func (uc UserController) GetUserInfo(ctx *gin.Context) {
	user, err := uc.UserRepository.GetCurrentUser(ctx)
	if err != nil {
		response.Fail(ctx, nil, "获取用户信息失败")
		return
	}
	// 返回
	response.Success(ctx, dto.ToUserInfoDto(user), "获取用户信息成功")
}

/**
 * @description: 获取用户列表
 * @param {*} ctx
 * @return {*}
 */
func (uc UserController) GetUsers(ctx *gin.Context) {
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
		response.Fail(ctx, nil, "获取用户列表失败："+err.Error())
		return
	}
	response.Success(ctx, gin.H{"content": dto.ToUsersDto(users), "total": total, "page": req.Page, "pageSize": req.PageSize}, "获取用户列表成功")
}

// 创建用户
func (uc UserController) CreateUser(ctx *gin.Context) {
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
			response.Fail(ctx, nil, "根据角色ID获取角色信息失败: "+err.Error())
			return
		}
	}
	userStatus := req.Status
	if userStatus == 0 {
		userStatus = 1
	}
	user := model.User{
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
		decodeData, err := utils.RSADecrypt([]byte(req.Password), config.Config.Application.RSAPrivateBytes)
		if err != nil {
			response.Fail(ctx, nil, err.Error())
			return
		}
		user.Password = utils.GenPasswd(string(decodeData))
	}
	err = uc.UserRepository.CreateUser(&user)
	if err != nil {
		response.Fail(ctx, nil, "创建用户失败: ")
		return
	}
	response.Success(ctx, nil, "创建用户成功")
}

func (uc UserController) UpdateUserById(ctx *gin.Context) {
	userId, _ := strconv.Atoi(ctx.Param("userId"))
	if userId <= 0 {
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
	oldUser, err := uc.UserRepository.GetUserById(uint(userId))
	if err != nil {
		response.Fail(ctx, nil, "获取需要更新的用户信息失败: "+err.Error())
		return
	}

	// 获取当前用户
	ctxUser, err := uc.UserRepository.GetCurrentUser(ctx)
	if err != nil {
		response.Fail(ctx, nil, err.Error())
		return
	}
	// 获取当前用户的所有角色
	currentRoles := ctxUser.Roles
	// 获取当前用户角色的排序，和前端传来的角色排序做比较
	var currentRoleSorts []int
	// 当前用户角色ID集合
	var currentRoleIds []uint
	for _, role := range currentRoles {
		currentRoleSorts = append(currentRoleSorts, int(role.Sort))
		currentRoleIds = append(currentRoleIds, role.ID)
	}
	// 当前用户角色排序最小值（最高等级角色）
	currentRoleSortMin := funk.MinInt(currentRoleSorts)

	// 获取前端传来的用户角色id
	reqRoleIds := req.RoleIds
	if len(reqRoleIds) == 0 {
		response.Fail(ctx, nil, "用户角色不能为空")
		return
	}
	// 根据角色id获取角色
	rr := repository.NewRoleRepository()
	roles, err := rr.GetRolesByIds(reqRoleIds)
	if err != nil {
		response.Fail(ctx, nil, "获取用户角色信息失败: ")
		return
	}
	if len(roles) == 0 {
		response.Fail(ctx, nil, "未获取到角色信息")
		return
	}
	var reqRoleSorts []int
	for _, role := range roles {
		reqRoleSorts = append(reqRoleSorts, int(role.Sort))
	}
	// 前端传来用户角色排序最小值（最高等级角色）
	reqRoleSortMin := funk.MinInt(reqRoleSorts)

	user := model.User{
		Model:        oldUser.Model,
		Username:     req.Username,
		Password:     oldUser.Password,
		Mobile:       req.Mobile,
		Avatar:       req.Avatar,
		Nickname:     req.Nickname,
		Introduction: req.Introduction,
		Status:       req.Status,
		Creator:      ctxUser.Username,
		Roles:        roles,
	}

	// 判断是更新自己还是更新别人
	if userId == int(ctxUser.ID) {
		// 如果是更新自己
		// 不能禁用自己
		if req.Status == 2 {
			response.Fail(ctx, nil, "不能禁用自己")
			return
		}
		// 不能更改自己的角色
		reqDiff, currentDiff := funk.Difference(req.RoleIds, currentRoleIds)
		if len(reqDiff.([]uint)) > 0 || len(currentDiff.([]uint)) > 0 {
			response.Fail(ctx, nil, "不能更改自己的角色")
			return
		}

		// 不能更新自己的密码，只能在个人中心更新
		if req.Password != "" {
			response.Fail(ctx, nil, "请到个人中心更新自身密码")
			return
		}

		// 密码赋值
		user.Password = ctxUser.Password

	} else {
		// 如果是更新别人
		// 用户不能更新比自己角色等级高的或者相同等级的用户
		// 根据path中的userIdID获取用户角色排序最小值
		minRoleSorts, err := uc.UserRepository.GetUserMinRoleSortsByIds([]uint{uint(userId)})
		if err != nil || len(minRoleSorts) == 0 {
			response.Fail(ctx, nil, "根据用户ID获取用户角色排序最小值失败")
			return
		}
		if currentRoleSortMin >= minRoleSorts[0] {
			response.Fail(ctx, nil, "用户不能更新比自己角色等级高的或者相同等级的用户")
			return
		}

		// 用户不能把别的用户角色等级更新得比自己高或相等
		if currentRoleSortMin >= reqRoleSortMin {
			response.Fail(ctx, nil, "用户不能把别的用户角色等级更新得比自己高或相等")
			return
		}

		// 密码赋值
		if req.Password != "" {
			// 密码通过RSA解密
			decodeData, err := utils.RSADecrypt([]byte(req.Password), config.Config.Application.RSAPrivateBytes)
			if err != nil {
				response.Fail(ctx, nil, err.Error())
				return
			}
			user.Password = utils.GenPasswd(string(decodeData))
		}
	}

	err = uc.UserRepository.UpdateUserById(uint(userId), &user)
	if err != nil {
		response.Fail(ctx, nil, "更新用户失败: "+err.Error())
		return
	}
	response.Success(ctx, nil, "更新用户成功")

}

// 更新用户登录密码
func (uc UserController) ChangePwd(ctx *gin.Context) {
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
	decodeOldPassword, err := utils.RSADecrypt([]byte(req.OldPassword), config.Config.Application.RSAPrivateBytes)
	if err != nil {
		response.Fail(ctx, nil, err.Error())
		return
	}
	decodeNewPassword, err := utils.RSADecrypt([]byte(req.NewPassword), config.Config.Application.RSAPrivateBytes)
	if err != nil {
		response.Fail(ctx, nil, err.Error())
		return
	}
	req.OldPassword = string(decodeOldPassword)
	req.NewPassword = string(decodeNewPassword)

	// 获取当前用户
	user, err := uc.UserRepository.GetCurrentUser(ctx)
	if err != nil {
		response.Fail(ctx, nil, err.Error())
		return
	}
	// 获取用户的真实正确密码
	correctPasswd := user.Password
	// 判断前端请求的密码是否等于真实密码
	err = utils.ComparePasswd(correctPasswd, req.OldPassword)
	if err != nil {
		response.Fail(ctx, nil, "原密码有误")
		return
	}
	// 更新密码
	err = uc.UserRepository.ChangePwd(user.Username, utils.GenPasswd(req.NewPassword))
	if err != nil {
		response.Fail(ctx, nil, "更新密码失败: "+err.Error())
		return
	}
	response.Success(ctx, nil, "更新密码成功")
}
