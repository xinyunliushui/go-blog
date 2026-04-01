/*
 * @Date: 2026-03-25 21:57:13
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-03-25 21:59:53
 * @Description: value object for user request
 */
package vo

// 用户登录结构体
type RegisterAndLoginRequest struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}
