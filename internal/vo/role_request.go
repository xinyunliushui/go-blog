/*
 * @Date: 2026-04-02 22:13:36
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-02 22:13:51
 * @Description: value object for role request
 */
package vo

// 角色列表请求结构体
type RoleListRequest struct {
	PaginationRequest
	Name    string `json:"name" form:"name"`
	Keyword string `json:"keyword" form:"keyword"`
	Status  uint   `json:"status" form:"status"`
}
