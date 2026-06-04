/*
 * @Date: 2026-04-02 22:13:36
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-20 21:38:18
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

// RoleWriteBody 创建/更新角色共用字段。
type RoleWriteBody struct {
	Name    string `json:"name" form:"name" validate:"required,min=1,max=20"`
	Keyword string `json:"keyword" form:"keyword" validate:"required,min=1,max=20"`
	Desc    string `json:"desc" form:"desc" validate:"min=0,max=100"`
	Status  uint   `json:"status" form:"status" validate:"oneof=1 2"`
	Sort    uint   `json:"sort" form:"sort" validate:"gte=1,lte=999"`
}

// CreateRoleRequest 创建角色：仅需 requestId，不需 version。
type CreateRoleRequest struct {
	IdempotentCreateRequest
	RoleWriteBody
}

// UpdateRoleRequest 更新角色：必须携带 version。
type UpdateRoleRequest struct {
	OptimisticLockRequest
	RoleWriteBody
}

// UpdateRoleMenusRequest 更新角色的权限菜单
type UpdateRoleMenusRequest struct {
	OptimisticLockRequest
	MenuIds []string `json:"menuIds" form:"menuIds"`
}
