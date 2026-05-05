/*
 * @Date: 2026-04-22 16:03:57
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-05-05 20:46:39
 * @Description: 文章接口实现
 */
package repository

import (
	"encoding/json"
	"fmt"
	"go-blog/internal/common"
	"go-blog/internal/config"
	"go-blog/internal/dto"
	"go-blog/internal/elasticsearch"
	"go-blog/internal/model"
	"go-blog/internal/utils"
	"go-blog/internal/vo"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type IBlogRepository interface {
	GetBlogs(req *vo.GetBlogListRequest) ([]model.Blog, int64, error)                      // 获取文章列表
	GetBlogById(blogId uint) (model.Blog, error)                                           // 根据ID获取文章详情
	CreateBlog(blog *model.Blog) error                                                     // 创建文章
	UpdateBlogPublishStatusById(blogId uint, status uint, publishedAt *time.Time) error    // 更新文章状态
	UpdateBlogById(blogId uint, blog *model.Blog) error                                    // 更新文章
	SearchBlogs(req *vo.SearchBlogRequest, ctx *gin.Context) (*dto.SearchResultDTO, error) // 搜索文章
}

type BlogRepository struct {
}

func NewBlogRepository() IBlogRepository {
	return BlogRepository{}
}

/** 获取文章列表
 * @param req *vo.GetBlogListRequest 文章列表请求
 * @return []model.Blog, int64, error
 */
func (br BlogRepository) GetBlogs(req *vo.GetBlogListRequest) ([]model.Blog, int64, error) {
	var blogList []model.Blog
	db := common.DB.Model(&model.Blog{}).Order("created_at DESC")
	status := req.Status
	if status != 0 {
		db = db.Where("status = ?", status)
	}
	//记录总条数
	var total int64
	err := db.Count(&total).Error
	if err != nil {
		return blogList, total, err
	}
	page := int(req.Page)
	pageSize := int(req.PageSize)
	if page > 0 && pageSize > 0 {
		err = db.Offset((page - 1) * pageSize).Limit(pageSize).Find(&blogList).Error
	} else {
		err = db.Find(&blogList).Error
	}
	return blogList, total, err
}

/** 创建文章
 * @param blog 文章信息
 * @return error
 */
func (br BlogRepository) CreateBlog(blog *model.Blog) error {
	err := common.DB.Create(blog).Error
	return err
}

/** 根据ID获取文章详情
 * @param blogId 文章ID
 * @return model.Blog, error
 */
func (br BlogRepository) GetBlogById(blogId uint) (model.Blog, error) {
	var blog model.Blog
	err := common.DB.Where("id = ?", blogId).First(&blog).Error
	return blog, err
}

/** 更新文章状态和发布时间
 * @param blogId 文章ID
 * @param status 文章状态
 * @param publishedAt 发布时间
 * @return error
 */
func (br BlogRepository) UpdateBlogPublishStatusById(blogId uint, status uint, publishedAt *time.Time) error {
	var publishedAtValue interface{}
	if publishedAt == nil {
		publishedAtValue = nil
	} else {
		publishedAtValue = *publishedAt
	}
	blog := map[string]interface{}{"status": status, "published_at": publishedAtValue}
	err := common.DB.Model(&model.Blog{}).Where("id = ?", blogId).Updates(blog).Error
	return err
}

/** 更新文章
 * @param blogId 文章ID
 * @param blog 文章信息
 * @return error
 */
func (br BlogRepository) UpdateBlogById(blogId uint, blog *model.Blog) error {
	err := common.DB.Model(&model.Blog{}).Where("id = ?", blogId).Updates(blog).Error
	return err
}

/** 搜索文章
 * @param req *vo.SearchBlogRequest 搜索文章请求
 * @return dto.SearchBlogDto, error
 */
func (br BlogRepository) SearchBlogs(req *vo.SearchBlogRequest, ctx *gin.Context) (*dto.SearchResultDTO, error) {
	queryDSL, err := BuildBlogSearchQueryDSL(*req)
	if err != nil {
		return nil, err
	}

	// 2. 执行ES请求
	index := config.Conf.ElasticSearch.IndexName
	res, err := elasticsearch.ESClient.Search(
		elasticsearch.ESClient.Search.WithContext(ctx),
		elasticsearch.ESClient.Search.WithIndex(index),
		elasticsearch.ESClient.Search.WithBody(strings.NewReader(queryDSL)),
		elasticsearch.ESClient.Search.WithTrackTotalHits(true),
		elasticsearch.ESClient.Search.WithPretty(),
	)
	if err != nil {
		return nil, fmt.Errorf("ES查询失败: %w", err)
	}
	defer res.Body.Close()

	// 3. 处理响应
	if res.IsError() {
		return nil, fmt.Errorf("ES错误 [%s]: %s", res.Status(), res.String())
	}

	var esResp dto.SearchBlogDto
	if err := json.NewDecoder(res.Body).Decode(&esResp); err != nil {
		return nil, fmt.Errorf("解析ES响应失败: %w", err)
	}

	// 4. 转换结果
	result := &dto.SearchResultDTO{
		Total: esResp.Hits.Total.Value,
		Took:  esResp.Took,
	}

	for _, hit := range esResp.Hits.Hits {
		post := dto.BlogPostSource{
			ID:          hit.Source.ID,
			Title:       utils.HighlightOrFallback(hit.Highlight["title"], hit.Source.Title),
			Summary:     utils.HighlightOrFallback(hit.Highlight["summary"], hit.Source.Summary),
			Content:     utils.HighlightOrFallback(hit.Highlight["content"], hit.Source.Content),
			PublishedAt: hit.Source.PublishedAt,
		}

		// 保留高亮信息
		if hl, ok := hit.Highlight["title"]; ok {
			post.Highlight.Title = hl
		}
		if hl, ok := hit.Highlight["content"]; ok {
			post.Highlight.Content = hl
		}
		result.Hits = append(result.Hits, post)
	}

	// 5. 添加拼写建议
	if len(esResp.Suggest.SpellCheck) > 0 &&
		len(esResp.Suggest.SpellCheck[0].Options) > 0 {
		result.Suggestion = esResp.Suggest.SpellCheck[0].Options[0].Text
	}

	return result, nil
}
