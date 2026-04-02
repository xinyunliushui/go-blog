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
