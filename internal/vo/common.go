/*
 * @Date: 2026-04-02 11:02:09
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-02 17:46:51
 * @Description:
 */
package vo

// 分页请求参数结构体
type PaginationRequest struct {
	Page     int    `json:"page" form:"page" binding:"required,min=1"`
	PageSize int    `json:"pageSize" form:"pageSize" binding:"required,min=1,max=100"`
	Sort     string `json:"sort" form:"sort"`
}

// IdempotentCreateRequest 创建类接口（POST …/create）：幂等 requestId，不携带 version。
type IdempotentCreateRequest struct {
	RequestId string `json:"requestId" form:"requestId" validate:"max=64"`
}

// OptimisticLockRequest 更新类接口（POST …/update/…）：必须携带 version，不使用 requestId。
type OptimisticLockRequest struct {
	Version uint `json:"version" form:"version" validate:"required,min=1"`
}
