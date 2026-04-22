/*
 * @Date: 2026-03-25 21:57:13
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-13 16:29:19
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

// 创建或更新用户结构体
type CreateOrUpdateUserRequest struct {
	Username     string `form:"username" json:"username" validate:"required,min=2,max=20"`
	Password     string `form:"password" json:"password"`
	Mobile       string `form:"mobile" json:"mobile" validate:"required,checkMobile"`
	Avatar       string `form:"avatar" json:"avatar"`
	Nickname     string `form:"nickname" json:"nickname" validate:"min=0,max=20"`
	Introduction string `form:"introduction" json:"introduction" validate:"min=0,max=255"`
	Status       uint   `form:"status" json:"status" validate:"oneof=1 2"`
	RoleIds      []uint `form:"roleIds" json:"roleIds" validate:"required"`
}

// 更新密码结构体
type ChangePwdRequest struct {
	OldPassword string `json:"oldPassword" form:"oldPassword" validate:"required"`
	NewPassword string `json:"newPassword" form:"newPassword" validate:"required"`
}
