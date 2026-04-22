package vo

// 创建文章请求结构体
type CreateBlogRequest struct {
	Title      string `json:"title" form:"title" validate:"required,min=1,max=100"`
	Content    string `json:"content" form:"content" validate:"required,min=1,max=20000"`
	Summary    string `json:"summary" form:"summary" validate:"required,min=1,max=500"`
	CoverImage string `json:"cover_image" form:"cover_image" validate:"required,min=1,max=255"`
	Category   string `json:"category" form:"category" validate:"max=255"`
	Tags       string `json:"tags" form:"tags" validate:"max=255"`
}

// 获取文章列表
type GetBlogListRequest struct {
	PaginationRequest
	Status uint `json:"status" form:"status"`
}
