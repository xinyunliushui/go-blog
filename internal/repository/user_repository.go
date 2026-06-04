/*
 * @Date: 2026-03-25 22:44:43
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-05-15 22:37:51
 * @Description: repository layer for user
 */
package repository

import (
	"errors"
	"fmt"
	"go-blog/internal/common"
	"go-blog/internal/model"
	"go-blog/internal/utils"
	"go-blog/internal/vo"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ErrInvalidCredentials 表示登录凭证无效（用户不存在、密码错误、账号非可用状态等），
// 对外统一使用该错误，避免区分失败原因造成用户名枚举。
var ErrInvalidCredentials = errors.New("用户名或密码不正确")
var ErrUserNotAssignedRoles = errors.New("用户未分配角色")

// 数据层方法接口
type IUserRepository interface {
	Login(user *model.User) (*model.User, error) // 登录
	// GetUsers 分页查询；status 为 nil 时不按状态筛选
	GetUsers(req *vo.UserListRequest) ([]model.User, int64, error)
	GetCurrentUser(c *gin.Context) (model.User, error)                  // 获取当前登录用户信息
	GetUserById(id string) (model.User, error)                          // 获取单个用户信息
	CreateUser(user *model.User) (duplicate bool, err error)                                  // 创建用户
	UpdateUserById(user *model.User, version uint) (duplicate bool, err error)                // 更新用户
	GetUserMinRoleSortsByIds(ids []string) ([]int, error)               // 根据用户ID获取用户角色排序最小值
	GetCurrentUserMinRoleSort(c *gin.Context) (uint, model.User, error) // 获取当前用户角色排序最小值（最高等级角色）以及当前用户信息
	ChangePwd(username string, version uint, hashNewPasswd string) (duplicate bool, err error) // 更新密码
}

type UserRepository struct {
}

func NewUserRepository() IUserRepository {
	return &UserRepository{}
}

/** 获取用户列表
 * @param req *vo.UserListRequest 用户列表请求
 * @return []model.User, int64, error
 */
func (ur *UserRepository) GetUsers(req *vo.UserListRequest) ([]model.User, int64, error) {

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

/** 登录
 * @param user *model.User 用户
 * @return *model.User, error
 */
func (ur *UserRepository) Login(user *model.User) (*model.User, error) {
	var firstUser model.User
	err := common.DB.
		Where("username = ?", user.Username).
		Preload("Roles").
		First(&firstUser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	if err := utils.ComparePasswd(firstUser.Password, user.Password); err != nil {
		return nil, ErrInvalidCredentials
	}
	if firstUser.Status != 1 {
		return nil, ErrInvalidCredentials
	}
	return &firstUser, nil
}

/** 获取当前登录用户信息
 * @param c *gin.Context 上下文
 * @return model.User, error
 */
func (ur *UserRepository) GetCurrentUser(c *gin.Context) (model.User, error) {
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

/** 获取单个用户信息
 * @param id string 用户ID
 * @return model.User, error
 */
func (ur *UserRepository) GetUserById(id string) (model.User, error) {
	var user model.User
	err := common.DB.Where("id = ?", id).Preload("Roles").First(&user).Error
	return user, err
}

/** 创建用户
 * @param user *model.User 用户
 * @return duplicate, error
 */
func (ur *UserRepository) CreateUser(user *model.User) (bool, error) {
	intended := *user
	targetRoleIDs := make([]string, 0, len(user.Roles))
	for _, role := range user.Roles {
		targetRoleIDs = append(targetRoleIDs, role.ID)
	}
	err := common.DB.Create(user).Error
	requestID := ""
	if user.RequestId != nil {
		requestID = *user.RequestId
	}
	return handleCreateIdempotency(common.DB, requestID, err, func() (bool, error) {
		var existing model.User
		if loadErr := common.DB.Where("request_id = ?", requestID).Preload("Roles").First(&existing).Error; loadErr != nil {
			return false, err
		}
		if !userFieldsMatch(&existing, &intended, targetRoleIDs) {
			return false, nil
		}
		*user = existing
		return true, nil
	})
}

func userFieldsMatch(current, target *model.User, targetRoleIDs []string) bool {
	if current.Username != target.Username ||
		current.Mobile != target.Mobile ||
		current.Avatar != target.Avatar ||
		current.Nickname != target.Nickname ||
		current.Introduction != target.Introduction ||
		current.Status != target.Status {
		return false
	}
	currentRoleIDs := make([]string, 0, len(current.Roles))
	for _, role := range current.Roles {
		currentRoleIDs = append(currentRoleIDs, role.ID)
	}
	return common.StringSliceEqualUnordered(currentRoleIDs, targetRoleIDs)
}

/** 更新用户信息及其角色关联
 * @param user *model.User 用户
 * @param version 乐观锁版本号
 * @return duplicate, error
 */
func (ur *UserRepository) UpdateUserById(user *model.User, version uint) (bool, error) {
	targetRoleIDs := make([]string, 0, len(user.Roles))
	for _, role := range user.Roles {
		targetRoleIDs = append(targetRoleIDs, role.ID)
	}

	var duplicate bool
	err := common.Transaction(func(gdb *gorm.DB) error {
		updates := map[string]interface{}{
			"username":     user.Username,
			"mobile":       user.Mobile,
			"avatar":       user.Avatar,
			"nickname":     user.Nickname,
			"introduction": user.Introduction,
			"status":       user.Status,
			"creator":      user.Creator,
			"version":      version + 1,
		}
		result := gdb.Model(&model.User{}).Where("id = ? AND version = ?", user.ID, version).Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			var current model.User
			if err := gdb.Where("id = ?", user.ID).Preload("Roles").First(&current).Error; err != nil {
				return err
			}
			if userFieldsMatch(&current, user, targetRoleIDs) {
				duplicate = true
				*user = current
				return nil
			}
			return common.ErrOptimisticLockConflict
		}
		if err := gdb.Model(user).Association("Roles").Replace(user.Roles); err != nil {
			return err
		}
		user.Version = version + 1
		return nil
	})
	return duplicate, err
}

/** 根据用户ID获取用户角色排序最小值
 * @param ids []string 用户ID列表
 * @return []int, error
 */
func (ur *UserRepository) GetUserMinRoleSortsByIds(ids []string) ([]int, error) {
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
		if len(user.Roles) == 0 {
			return nil, ErrUserNotAssignedRoles
		}
		roleSortList := make([]int, 0, len(user.Roles))
		for _, role := range user.Roles {
			roleSortList = append(roleSortList, int(role.Sort))
		}
		roleMinSortList = append(roleMinSortList, minRoleSort(roleSortList))
	}
	return roleMinSortList, nil
}

/** 获取当前用户角色排序最小值（最高等级角色）以及当前用户信息
 * @param c *gin.Context 上下文
 * @return uint, model.User, error
 */
func (ur *UserRepository) GetCurrentUserMinRoleSort(c *gin.Context) (uint, model.User, error) {
	// 获取当前用户
	ctxUser, err := ur.GetCurrentUser(c)
	if err != nil {
		return 999, ctxUser, err
	}
	currentRoles := ctxUser.Roles
	if len(currentRoles) == 0 {
		return 0, ctxUser, errors.New("当前用户未分配角色")
	}
	currentRoleSorts := make([]int, 0, len(currentRoles))
	for _, role := range currentRoles {
		currentRoleSorts = append(currentRoleSorts, int(role.Sort))
	}
	currentRoleSortMin := uint(minRoleSort(currentRoleSorts))

	return currentRoleSortMin, ctxUser, nil
}

// minRoleSort 返回角色 Sort 的最小值（数值越小权限越高）。vals 必须非空。
func minRoleSort(vals []int) int {
	m := vals[0]
	for i := 1; i < len(vals); i++ {
		if vals[i] < m {
			m = vals[i]
		}
	}
	return m
}

/** 更新密码
 * @param username string 用户名
 * @param version 乐观锁版本号
 * @param hashNewPasswd string 新密码
 * @return duplicate, error
 */
func (ur *UserRepository) ChangePwd(username string, version uint, hashNewPasswd string) (bool, error) {
	result := common.DB.Model(&model.User{}).
		Where("username = ? AND version = ?", username, version).
		Updates(map[string]interface{}{"password": hashNewPasswd, "version": version + 1})
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected > 0 {
		return false, nil
	}
	var current model.User
	if err := common.DB.Where("username = ?", username).First(&current).Error; err != nil {
		return false, err
	}
	return applyOptimisticLockResult(current.Password == hashNewPasswd, result.RowsAffected, nil)
}
