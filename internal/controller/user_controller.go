package controller

import (
	"go-blog/internal/repository"

	"github.com/gin-gonic/gin"
)

type IUserController interface {
	GetUserInfo(ctx *gin.Context)
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
	user, err := uc.UserRepository.GetUserInfo(ctx)
	if err != nil {
		ctx.JSON(500, gin.H{
			"message": "获取用户信息失败",
			"data":    err.Error(),
		})
		return
	}
	// 返回
	ctx.JSON(200, gin.H{
		"message": "Hello, World!",
		"data":    user,
	})
}
