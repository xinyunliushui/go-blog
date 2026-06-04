/*
 * @Date: 2026-03-25 21:57:13
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-05-15 22:32:11
 * @Description: value object for user request
 */
package vo

// 用户登录结构体
type RegisterAndLoginRequest struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

// 用户列表请求结构体
type UserListRequest struct {
	PaginationRequest
	Status uint `json:"status" form:"status"`
}

// UserWriteBody 创建/更新用户共用字段（不含 roleIds；注册可不传角色）。
type UserWriteBody struct {
	Username     string `form:"username" json:"username" validate:"required,min=2,max=20"`
	Password     string `form:"password" json:"password"`
	Mobile       string `form:"mobile" json:"mobile" validate:"required,checkMobile"`
	Avatar       string `form:"avatar" json:"avatar"`
	Nickname     string `form:"nickname" json:"nickname" validate:"min=0,max=20"`
	Introduction string `form:"introduction" json:"introduction" validate:"min=0,max=255"`
	Status       uint   `form:"status" json:"status" validate:"omitempty,oneof=1 2"`
}

// CreateUserRequest 创建用户（POST /user/create、POST /auth/register）：仅需 requestId，不需 version；roleIds 注册时可选。
type CreateUserRequest struct {
	IdempotentCreateRequest
	UserWriteBody
	RoleIds []string `form:"roleIds" json:"roleIds" validate:"omitempty,dive"`
}

// UpdateUserRequest 更新用户（POST /user/update/:userId）：必须携带 version；roleIds 省略表示不修改角色，传入时才会更新并做等级校验。
type UpdateUserRequest struct {
	OptimisticLockRequest
	UserWriteBody
	RoleIds *[]string `form:"roleIds" json:"roleIds" validate:"omitempty,dive"`
}

// 更新密码结构体
type ChangePwdRequest struct {
	OptimisticLockRequest
	OldPassword string `json:"oldPassword" form:"oldPassword" validate:"required"`
	NewPassword string `json:"newPassword" form:"newPassword" validate:"required"`
}
