package controller

import (
	"go-blog/internal/repository"
	"go-blog/internal/response"

	"github.com/gin-gonic/gin"
)

type IUserController interface {
	GetUserInfo(ctx *gin.Context)
	GetUsers(ctx *gin.Context)
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
	response.Success(ctx, user, "获取用户信息成功")
}

func (uc UserController) GetUsers(ctx *gin.Context) {
	users, err := uc.UserRepository.GetUsers(ctx)
	if err != nil {
		response.Fail(ctx, nil, "获取用户列表失败")
		return
	}
	// 返回
	response.Success(ctx, users, "获取用户列表成功")
}
