/*
 * @Date: 2026-03-25 22:44:43
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-03-31 22:31:16
 * @Description: repository layer for user
 */
package repository

import (
	"errors"
	"go-blog/internal/common"
	"go-blog/internal/model"

	"github.com/gin-gonic/gin"
)

// 数据层方法接口
type IUserRepository interface {
	Login(user *model.User) (*model.User, error) // 登录
	GetUserInfo(ctx *gin.Context) ([]model.User, error)
}

type UserRepository struct {
}

func NewUserRepository() IUserRepository {
	return UserRepository{}
}

func (ur UserRepository) GetUserInfo(ctx *gin.Context) ([]model.User, error) {
	var users []model.User
	err := common.DB.Find(&users).Error
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	return users, nil
}

func (ur UserRepository) Login(user *model.User) (*model.User, error) {
	// 根据用户名获取用户(正常状态:用户状态正常)
	var firstUser model.User
	err := common.DB.
		Where("username = ?", user.Username).
		Preload("Roles").
		First(&firstUser).Error
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	// 校验密码
	// err = utils.ComparePasswd(firstUser.Password, user.Password)
	if firstUser.Password != user.Password {
		return &firstUser, errors.New("密码错误")
	}

	// 判断用户的状态
	userStatus := firstUser.Status
	if userStatus != 1 {
		return nil, errors.New("用户被禁用")
	}
	return &firstUser, nil
}
