/*
 * @Date: 2026-04-22 16:01:43
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-05-08 17:43:08
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
	"go-blog/internal/vo"
	"strconv"
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
		return
	}
	response.Success(ctx, gin.H{"content": dto.ToBlogsDto(blogs), "total": total, "page": req.Page, "pageSize": req.PageSize}, "获取文章列表成功")
}

// 获取文章详情
func (bc BlogController) GetBlogById(ctx *gin.Context) {
	blogId, _ := strconv.Atoi(ctx.Param("blogId"))
	if blogId <= 0 {
		response.Fail(ctx, nil, "文章ID不正确")
		return
	}
	blog, err := bc.BlogRepository.GetBlogById(uint(blogId))
	if err != nil {
		response.Fail(ctx, nil, "获取文章详情失败")
		return
	}
	response.Success(ctx, blog, "获取文章详情成功")
}

func (bc BlogController) CreateBlog(ctx *gin.Context) {
	var req vo.CreateAndUpdateBlogRequest
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

	if err := rabbitmq.PublishMessage(config.Conf.Rabbitmq.QueueName, &blog); err != nil {
		common.Log.Errorf("[MQ_OUTBOX_ALERT] blog_id=%d RabbitMQ 首次投递失败，写入本地补偿表: %v", blog.ID, err)
		outboxRepo := repository.NewMQOutboxRepository()
		if encErr := outboxRepo.EnqueueBlogPublish(&blog, err.Error()); encErr != nil {
			common.Log.Errorf("[MQ_OUTBOX_ALERT] blog_id=%d 补偿表写入失败（需人工核对 MySQL 与 ES/CH）: %v", blog.ID, encErr)
			response.Fail(ctx, gin.H{"blogId": blog.ID}, "文章已保存，但消息队列不可用且补偿记录失败，请联系管理员")
			return
		}
		response.Success(ctx, gin.H{"blogId": blog.ID}, "文章已保存，同步至检索系统将自动重试")
		return
	}
	response.Success(ctx, gin.H{"blogId": blog.ID}, "创建文章成功")
}

// 更新文章状态
func (bc BlogController) UpdateBlogPublishStatusById(ctx *gin.Context) {
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
	// 获取路径中的blogId
	blogId, _ := strconv.Atoi(ctx.Param("blogId"))
	if blogId <= 0 {
		response.Fail(ctx, nil, "文章ID不正确")
		return
	}
	var publishedAt *time.Time
	if req.Status == 2 {
		now := time.Now()
		publishedAt = &now
	}
	err := bc.BlogRepository.UpdateBlogPublishStatusById(uint(blogId), req.Status, publishedAt)
	// 更新 ES 发布状态
	if err := repository.UpdateBlogPublishStatusInES(uint(blogId), req.Status, publishedAt); err != nil {
		common.Log.Errorf("blog_id=%d 更新 ES 发布状态失败: %v", blogId, err)
	}
	if err != nil {
		response.Fail(ctx, nil, "更新文章状态失败")
		return
	}
	response.Success(ctx, nil, "更新文章状态成功")
}

// 更新文章
func (bc BlogController) UpdateBlogById(ctx *gin.Context) {
	var req vo.CreateAndUpdateBlogRequest
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
	// 获取路径中的blogId
	blogId, _ := strconv.Atoi(ctx.Param("blogId"))
	if blogId <= 0 {
		response.Fail(ctx, nil, "文章ID不正确")
		return
	}
	blog := model.Blog{
		Title:      req.Title,
		Content:    req.Content,
		Summary:    req.Summary,
		CoverImage: req.CoverImage,
		Category:   &req.Category,
		Tags:       &req.Tags,
	}
	err := bc.BlogRepository.UpdateBlogById(uint(blogId), &blog)
	if err != nil {
		response.Fail(ctx, nil, "更新文章失败")
		return
	}
	if err := repository.UpdateBlogFieldsInESByBlog(uint(blogId), &blog); err != nil {
		common.Log.Errorf("blog_id=%d 同步 ES 可检索字段失败: %v", blogId, err)
	}
	response.Success(ctx, nil, "更新文章成功")
}

// 搜索文章
func (bc BlogController) SearchBlogs(ctx *gin.Context) {
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
		common.Log.Errorf("搜索文章失败: %v", err)
		response.Fail(ctx, nil, "搜索文章失败")
		return
	}
	response.Success(ctx, searchBlogDto, "搜索文章成功")
}
