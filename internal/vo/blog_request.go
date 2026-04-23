/*
 * @Date: 2026-04-22 16:19:27
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-23 16:58:22
 * @Description:
 */
package vo

// 创建文章请求结构体
type CreateAndUpdateBlogRequest struct {
	Title      string `json:"title" validate:"required,min=1,max=100"`
	Content    string `json:"content" validate:"required,min=1,max=20000"`
	Summary    string `json:"summary" validate:"required,min=1,max=500"`
	CoverImage string `json:"cover_image" validate:"required,min=1,max=255"`
	Category   string `json:"category" validate:"max=255"`
	Tags       string `json:"tags" validate:"max=255"`
}

// 获取文章列表
type GetBlogListRequest struct {
	PaginationRequest
	Status uint `json:"status" form:"status"`
}

// 更新文章状态
type UpdateBlogPublishStatusRequest struct {
	BlogId uint `json:"blogId" validate:"required"`
	// 1草稿, 2发布, 3私密，草稿是默认态
	Status uint `json:"status" validate:"required,oneof=2 3"`
}
