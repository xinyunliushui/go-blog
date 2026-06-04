/*
 * @Date: 2026-04-22 16:01:43
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-05-15
 * @Description: 文章控制器接口实现
 */
package controller

import (
	"go-blog/internal/common"
	"go-blog/internal/config"
	"go-blog/internal/dto"
	"go-blog/internal/model"
	"go-blog/internal/rabbitmq"
	"go-blog/internal/repository"
	"go-blog/internal/response"
	"go-blog/internal/utils"
	"go-blog/internal/vo"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type IBlogController interface {
	GetBlogs(c *gin.Context)                    // 获取文章列表
	GetBlogById(c *gin.Context)                 // 获取文章详情
	CreateBlog(c *gin.Context)                  // 创建文章
	UpdateBlogPublishStatusById(c *gin.Context) // 更新文章状态
	UpdateBlogById(c *gin.Context)              // 更新文章
	SearchBlogs(c *gin.Context)                 // 搜索文章
}

type BlogController struct {
	BlogRepository repository.IBlogRepository
}

func NewBlogController() IBlogController {
	return &BlogController{
		BlogRepository: repository.NewBlogRepository(),
	}
}

// 获取文章列表
func (bc *BlogController) GetBlogs(ctx *gin.Context) {
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
		response.FailErr(ctx, nil, "获取文章列表失败", err)
		return
	}
	response.Success(ctx, gin.H{"content": dto.ToBlogsDto(blogs), "total": total, "page": req.Page, "pageSize": req.PageSize}, "获取文章列表成功")
}

// 获取文章详情
func (bc *BlogController) GetBlogById(ctx *gin.Context) {
	blogId := strings.TrimSpace(ctx.Param("blogId"))
	if blogId == "" {
		response.Fail(ctx, nil, "文章ID不正确")
		return
	}
	blog, err := bc.BlogRepository.GetBlogById(blogId)
	if err != nil {
		response.FailErr(ctx, nil, "获取文章详情失败", err)
		return
	}
	response.Success(ctx, dto.ToBlogDto(blog), "获取文章详情成功")
}

func (bc *BlogController) CreateBlog(ctx *gin.Context) {
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
		response.FailErr(ctx, nil, "获取当前用户信息失败", err)
		return
	}
	requestID, err := common.ResolveRequestID(ctx, req.RequestId)
	if err != nil {
		response.Fail(ctx, nil, err.Error())
		return
	}
	blog := &model.Blog{
		UUIDModel:  model.UUIDModel{RequestId: common.RequestIDPtr(requestID)},
		Title:      req.Title,
		Content:    req.Content,
		Summary:    req.Summary,
		CoverImage: req.CoverImage,
		Category:   utils.OptionalString(req.Category),
		Tags:       utils.OptionalString(req.Tags),
		Author:     ctxUser.Username, // 文章作者
	}
	duplicate, err := bc.BlogRepository.CreateBlog(blog)
	if err != nil {
		handleWriteError(ctx, "创建文章失败", err)
		return
	}
	if duplicate {
		response.Duplicate(ctx, gin.H{"blogId": blog.ID}, msgCreateDuplicate)
		return
	}

	if err := rabbitmq.PublishMessage(ctx.Request.Context(), config.Conf.Rabbitmq.QueueName, blog); err != nil {
		common.LoggerFromGin(ctx).Errorf("[消息推送补偿] blog_id=%s RabbitMQ 首次投递失败，写入本地补偿表: %v", blog.ID, err)
		compRepo := repository.NewMQCompensationRepository()
		if encErr := compRepo.EnqueueBlogPublish(blog, err.Error(), common.TraceIDFromGin(ctx)); encErr != nil {
			common.LoggerFromGin(ctx).Errorf("[消息推送补偿] blog_id=%s 补偿表写入失败（需人工核对 MySQL 与 ES/CH）: %v", blog.ID, encErr)
			response.Fail(ctx, gin.H{"blogId": blog.ID}, "文章已保存，但消息队列不可用且补偿记录失败，请联系管理员")
			return
		}
		response.Success(ctx, gin.H{"blogId": blog.ID}, "文章已保存，同步至检索系统将自动重试")
		return
	}
	response.Success(ctx, gin.H{"blogId": blog.ID}, "创建文章成功")
}

// 更新文章状态
func (bc *BlogController) UpdateBlogPublishStatusById(ctx *gin.Context) {
	var req vo.UpdateBlogPublishStatusRequest
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
	blogId := strings.TrimSpace(ctx.Param("blogId"))
	if blogId == "" {
		response.Fail(ctx, nil, "文章ID不正确")
		return
	}
	if !requireVersion(ctx, req.Version) {
		return
	}
	var publishedAt *time.Time
	if req.Status == 2 {
		now := time.Now()
		publishedAt = &now
	}
	duplicate, err := bc.BlogRepository.UpdateBlogPublishStatusById(blogId, req.Version, req.Status, publishedAt)
	// 更新 ES 发布状态
	if err == nil && !duplicate {
		if err := repository.UpdateBlogPublishStatusInES(blogId, req.Status, publishedAt); err != nil {
			common.Log.Errorf("blog_id=%s 更新 ES 发布状态失败: %v", blogId, err)
		}
	}
	if err != nil {
		handleWriteError(ctx, "更新文章状态失败", err)
		return
	}
	if duplicate {
		response.Duplicate(ctx, nil, msgUpdateDuplicate)
		return
	}
	response.Success(ctx, nil, "更新文章状态成功")
}

// 更新文章
func (bc *BlogController) UpdateBlogById(ctx *gin.Context) {
	var req vo.UpdateBlogRequest
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
	blogId := strings.TrimSpace(ctx.Param("blogId"))
	if blogId == "" {
		response.Fail(ctx, nil, "文章ID不正确")
		return
	}
	if !requireVersion(ctx, req.Version) {
		return
	}
	blog := &model.Blog{
		Title:      req.Title,
		Content:    req.Content,
		Summary:    req.Summary,
		CoverImage: req.CoverImage,
		Category:   utils.OptionalString(req.Category),
		Tags:       utils.OptionalString(req.Tags),
	}
	duplicate, err := bc.BlogRepository.UpdateBlogById(blogId, req.Version, blog)
	if err != nil {
		handleWriteError(ctx, "更新文章失败", err)
		return
	}
	if duplicate {
		response.Duplicate(ctx, nil, msgUpdateDuplicate)
		return
	}
	if err := repository.UpdateBlogFieldsInESByBlog(blogId, blog); err != nil {
		common.Log.Errorf("blog_id=%s 同步 ES 可检索字段失败: %v", blogId, err)
	}
	response.Success(ctx, nil, "更新文章成功")
}

// 搜索文章
func (bc *BlogController) SearchBlogs(ctx *gin.Context) {
	var req vo.SearchBlogRequest
	// 参数绑定
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.Fail(ctx, nil, common.ValidationErrString(err))
		return
	}
	// 参数校验
	if err := common.Validate.Struct(&req); err != nil {
		response.Fail(ctx, nil, common.ValidationErrString(err))
		return
	}
	searchBlogDto, err := bc.BlogRepository.SearchBlogs(&req, ctx)
	if err != nil {
		response.FailErr(ctx, nil, "搜索文章失败", err)
		return
	}
	response.Success(ctx, searchBlogDto, "搜索文章成功")
}
