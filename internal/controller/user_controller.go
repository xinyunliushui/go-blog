/*
 * @Date: 2026-03-25 22:08:27
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-13 21:48:58
 * @Description: user controller
 */
package controller

import (
	"go-blog/internal/common"
	"go-blog/internal/dto"
	"go-blog/internal/model"
	"go-blog/internal/repository"
	"go-blog/internal/response"
	"go-blog/internal/vo"
	"strconv"

	"github.com/gin-gonic/gin"
)

type IUserController interface {
	GetUserInfo(ctx *gin.Context)
	GetUsers(ctx *gin.Context)
	CreateUser(ctx *gin.Context)
	UpdateUserById(ctx *gin.Context)
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
	// 密码通过RSA解密
	// 密码不为空就解密
	// if req.Password != "" {
	// 	decodeData, err := util.RSADecrypt([]byte(req.Password), config.Conf.System.RSAPrivateBytes)
	// 	if err != nil {
	// 		response.Fail(ctx, nil, err.Error())
	// 		return
	// 	}
	// 	req.Password = string(decodeData)
	// 	if len(req.Password) < 6 {
	// 		response.Fail(ctx, nil, "密码长度至少为6位")
	// 		return
	// 	}
	// }
	// 创建用户
	user := model.User{
		Username:     req.Username,
		Password:     req.Password,
		Mobile:       req.Mobile,
		Avatar:       req.Avatar,
		Nickname:     req.Nickname,
		Introduction: req.Introduction,
		Status:       1,
		// Creator:      req.Username,
	}
	err := uc.UserRepository.CreateUser(&user)
	if err != nil {
		response.Fail(ctx, nil, "创建用户失败: "+err.Error())
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
	// BindAndValidate(ctx, &req)
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

	user := model.User{
		Username:     req.Username,
		Mobile:       req.Mobile,
		Avatar:       req.Avatar,
		Nickname:     req.Nickname,
		Introduction: req.Introduction,
		Status:       req.Status,
		Roles:        req.Roles,
	}

	err := uc.UserRepository.UpdateUserById(uint(userId), &user)
	if err != nil {
		response.Fail(ctx, nil, "更新用户失败: "+err.Error())
		return
	}
	response.Success(ctx, nil, "更新用户成功")

}
