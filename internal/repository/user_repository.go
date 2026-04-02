/*
 * @Date: 2026-03-25 22:44:43
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-02 21:52:35
 * @Description: repository layer for user
 */
package repository

import (
	"errors"
	"go-blog/internal/common"
	"go-blog/internal/model"
	"go-blog/internal/vo"

	"github.com/gin-gonic/gin"
)

// 数据层方法接口
type IUserRepository interface {
	Login(user *model.User) (*model.User, error) // 登录
	// GetUsers 分页查询；status 为 nil 时不按状态筛选
	GetUsers(req *vo.UserListRequest) ([]*model.User, int64, error)
	GetCurrentUser(c *gin.Context) (model.User, error) // 获取当前登录用户信息
}

type UserRepository struct {
}

func NewUserRepository() IUserRepository {
	return UserRepository{}
}

func (ur UserRepository) GetUsers(req *vo.UserListRequest) ([]*model.User, int64, error) {

	var list []*model.User
	db := common.DB.Model(&model.User{}).Order("created_at DESC")

	// username := strings.TrimSpace(req.Username)
	// if username != "" {
	// 	db = db.Where("username LIKE ?", fmt.Sprintf("%%%s%%", username))
	// }
	// nickname := strings.TrimSpace(req.Nickname)
	// if nickname != "" {
	// 	db = db.Where("nickname LIKE ?", fmt.Sprintf("%%%s%%", nickname))
	// }
	// mobile := strings.TrimSpace(req.Mobile)
	// if mobile != "" {
	// 	db = db.Where("mobile LIKE ?", fmt.Sprintf("%%%s%%", mobile))
	// }
	status := req.Status
	if status != 0 {
		db = db.Where("status = ?", status)
	}
	// 当pageNum > 0 且 pageSize > 0 才分页
	//记录总条数
	var total int64
	err := db.Count(&total).Error
	if err != nil {
		return list, total, err
	}
	page := int(req.Page)
	pageSize := int(req.PageSize)
	if page > 0 && pageSize > 0 {
		err = db.Offset((page - 1) * pageSize).Limit(pageSize).Preload("Roles").Find(&list).Error
	} else {
		err = db.Preload("Roles").Find(&list).Error
	}
	return list, total, err
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

func (ur UserRepository) GetCurrentUser(c *gin.Context) (model.User, error) {
	var newUser model.User
	ctxUser, exist := c.Get("user")
	if !exist {
		return newUser, errors.New("用户未登录")
	}
	u, _ := ctxUser.(model.User)
	return u, nil
}
