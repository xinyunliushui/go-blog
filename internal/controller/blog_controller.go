/*
 * @Date: 2026-04-22 16:01:43
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-22 17:41:44
 * @Description: blog controller
 */
package controller

import (
	"go-blog/internal/common"
	"go-blog/internal/dto"
	"go-blog/internal/model"
	"go-blog/internal/repository"
	"go-blog/internal/response"
	"go-blog/internal/vo"

	"github.com/gin-gonic/gin"
)

type IBlogController interface {
	GetBlogs(c *gin.Context)   // 获取文章列表
	CreateBlog(c *gin.Context) // 创建文章
	// GetBlogById(c *gin.Context)    // 获取文章详情
	// UpdateBlogById(c *gin.Context) // 更新文章
	// DeleteBlogById(c *gin.Context) // 删除文章
}

type BlogController struct {
	BlogRepository repository.IBlogRepository
}

func NewBlogController() IBlogController {
	return BlogController{
		BlogRepository: repository.NewBlogRepository(),
	}
}

// 获取文章列表
func (bc BlogController) GetBlogs(ctx *gin.Context) {
	var req vo.GetBlogListRequest
	// 参数绑定
	if err := ctx.ShouldBind(&req); err != nil {
		response.Fail(ctx, nil, common.ValidationErrString(err))
		return
	}
	// 参数校验
	if err := common.Validate.Struct(&req); err != nil {
		response.Fail(ctx, nil, common.ValidationErrString(err))
		return
	}
	blogs, total, err := bc.BlogRepository.GetBlogs(&req)
	if err != nil {
		response.Fail(ctx, nil, "获取文章列表失败")
	}
	response.Success(ctx, gin.H{"content": dto.ToBlogsDto(blogs), "total": total, "page": req.Page, "pageSize": req.PageSize}, "获取用户列表成功")
}

func (bc BlogController) CreateBlog(ctx *gin.Context) {
	var req vo.CreateBlogRequest
	// 参数绑定
	if err := ctx.ShouldBind(&req); err != nil {
		response.Fail(ctx, nil, common.ValidationErrString(err))
		return
	}
	// 参数校验
	if err := common.Validate.Struct(&req); err != nil {
		response.Fail(ctx, nil, common.ValidationErrString(err))
		return
	}
	// 获取当前用户
	ur := repository.NewUserRepository()
	ctxUser, err := ur.GetCurrentUser(ctx)
	if err != nil {
		response.Fail(ctx, nil, "获取当前用户信息失败")
		return
	}
	blog := model.Blog{
		Title:      req.Title,
		Content:    req.Content,
		Summary:    req.Summary,
		CoverImage: req.CoverImage,
		Category:   &req.Category,
		Tags:       &req.Tags,
		Author:     ctxUser.Username, // 文章作者
	}
	err = bc.BlogRepository.CreateBlog(&blog)
	if err != nil {
		response.Fail(ctx, nil, "创建文章失败")
		return
	}
	response.Success(ctx, nil, "创建文章成功")
}
