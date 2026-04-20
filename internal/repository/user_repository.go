/*
 * @Date: 2026-03-25 22:44:43
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-20 20:35:08
 * @Description: repository layer for user
 */
package repository

import (
	"errors"
	"go-blog/internal/common"
	"go-blog/internal/model"
	"go-blog/internal/vo"

	"github.com/gin-gonic/gin"
	"github.com/thoas/go-funk"
	"gorm.io/gorm"
)

// 数据层方法接口
type IUserRepository interface {
	Login(user *model.User) (*model.User, error) // 登录
	// GetUsers 分页查询；status 为 nil 时不按状态筛选
	GetUsers(req *vo.UserListRequest) ([]model.User, int64, error)
	GetCurrentUser(c *gin.Context) (model.User, error)  // 获取当前登录用户信息
	GetUserById(id uint) (model.User, error)            // 获取单个用户信息
	CreateUser(user *model.User) error                  // 创建用户
	UpdateUserById(id uint, user *model.User) error     // 更新用户
	GetUserMinRoleSortsByIds(ids []uint) ([]int, error) // 根据用户ID获取用户角色排序最小值
}

type UserRepository struct {
}

func NewUserRepository() IUserRepository {
	return UserRepository{}
}

func (ur UserRepository) GetUsers(req *vo.UserListRequest) ([]model.User, int64, error) {

	var list []model.User
	db := common.DB.Model(&model.User{}).Order("created_at DESC")
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
	//  参数判断
	u, _ := ctxUser.(model.User)
	user, err := ur.GetUserById(u.ID)
	if err != nil {
		return newUser, err
	}
	return user, err
}

// 获取单个用户信息
func (ur UserRepository) GetUserById(id uint) (model.User, error) {
	var user model.User
	err := common.DB.Where("id = ?", id).Preload("Roles").First(&user).Error
	return user, err
}

// 创建用户
func (ur UserRepository) CreateUser(user *model.User) error {
	err := common.DB.Create(user).Error
	return err
}

// 更新用户信息及其角色关联
func (ur UserRepository) UpdateUserById(id uint, user *model.User) error {
	err := common.Transaction(func(tx *gorm.DB) error {
		// 1. 更新 users 表
		if err := tx.Model(user).Updates(user).Error; err != nil {
			return err
		}
		// 2. 更新 user_role 关联表（Replace 会先删旧关联再建新关联）
		if err := tx.Model(user).Association("Roles").Replace(user.Roles); err != nil {
			return err
		}
		return nil
	})
	return err
}

// 根据用户ID获取用户角色排序最小值
func (ur UserRepository) GetUserMinRoleSortsByIds(ids []uint) ([]int, error) {
	// 根据用户ID获取用户信息
	var userList []model.User
	err := common.DB.Where("id IN (?)", ids).Preload("Roles").Find(&userList).Error
	if err != nil {
		return []int{}, err
	}
	if len(userList) == 0 {
		return []int{}, errors.New("未获取到任何用户信息")
	}
	var roleMinSortList []int
	for _, user := range userList {
		roles := user.Roles
		var roleSortList []int
		for _, role := range roles {
			roleSortList = append(roleSortList, int(role.Sort))
		}
		roleMinSort := funk.MinInt(roleSortList)
		roleMinSortList = append(roleMinSortList, roleMinSort)
	}
	return roleMinSortList, nil
}
