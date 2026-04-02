/*
 * @Date: 2026-03-25 22:08:27
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-02 22:05:10
 * @Description: user controller
 */
package controller

import (
	"go-blog/internal/common"
	"go-blog/internal/dto"
	"go-blog/internal/repository"
	"go-blog/internal/response"
	"go-blog/internal/vo"

	"github.com/gin-gonic/gin"
)

type IUserController interface {
	GetUserInfo(ctx *gin.Context)
	GetUsers(ctx *gin.Context)
	GetRoles(ctx *gin.Context)
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
	response.Success(ctx, gin.H{"user": user}, "获取用户信息成功")
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

func (uc UserController) GetRoles(ctx *gin.Context) {
	var req vo.RoleListRequest
	// 参数绑定
	if err := ctx.ShouldBind(&req); err != nil {
		response.Fail(ctx, nil, common.ValidationErrString(err))
		return
	}
}
