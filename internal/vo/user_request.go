/*
 * @Date: 2026-03-25 21:57:13
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-02 22:13:43
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
