/*
 * @Date: 2026-04-13 14:57:25
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-13 14:57:33
 * @Description:
 */
package vo

// 创建接口结构体
type CreateMenuRequest struct {
	Name     string `json:"name" form:"name" validate:"required,min=1,max=50"`
	Title    string `json:"title" form:"title" validate:"required,min=1,max=50"`
	Icon     string `json:"icon" form:"icon" validate:"min=0,max=50"`
	Path     string `json:"path" form:"path" validate:"required,min=1,max=100"`
	Redirect string `json:"redirect" form:"redirect" validate:"min=0,max=100"`
	Sort     uint   `json:"sort" form:"sort" validate:"gte=1,lte=999"`
	Status   uint   `json:"status" form:"status" validate:"oneof=1 2"`
	Type     uint   `json:"Type" form:"Type" validate:"oneof=1 2 3"`
	ParentId string `json:"parentId" form:"parentId"`
}
