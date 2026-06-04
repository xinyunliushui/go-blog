/*
 * @Date: 2026-04-22 16:03:57
 * @Author: zhongwenhao
 * @LastEditors: zhongwh 746227367@qq.com
 * @LastEditTime: 2026-05-17 14:54:54
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
	GetBlogById(blogId string) (model.Blog, error)                                         // 根据ID获取文章详情
	CreateBlog(blog *model.Blog) (duplicate bool, err error)                               // 创建文章
	UpdateBlogPublishStatusById(blogId string, version uint, status uint, publishedAt *time.Time) (duplicate bool, err error) // 更新文章状态
	UpdateBlogById(blogId string, version uint, blog *model.Blog) (duplicate bool, err error) // 更新文章
	SearchBlogs(req *vo.SearchBlogRequest, ctx *gin.Context) (*dto.SearchResultDTO, error) // 搜索文章
}

type BlogRepository struct {
}

func NewBlogRepository() IBlogRepository {
	return &BlogRepository{}
}

/** 获取文章列表
 * @param req *vo.GetBlogListRequest 文章列表请求
 * @return []model.Blog, int64, error
 */
func (*BlogRepository) GetBlogs(req *vo.GetBlogListRequest) ([]model.Blog, int64, error) {
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
 * @return duplicate, error
 */
func (*BlogRepository) CreateBlog(blog *model.Blog) (bool, error) {
	intended := *blog
	err := common.DB.Create(blog).Error
	requestID := ""
	if blog.RequestId != nil {
		requestID = *blog.RequestId
	}
	return handleCreateIdempotency(common.DB, requestID, err, func() (bool, error) {
		var existing model.Blog
		if loadErr := common.DB.Where("request_id = ?", requestID).First(&existing).Error; loadErr != nil {
			return false, err
		}
		if !blogCreateFieldsMatch(&existing, &intended) {
			return false, nil
		}
		*blog = existing
		return true, nil
	})
}

func blogFieldsMatch(current, target *model.Blog) bool {
	return current.Title == target.Title &&
		current.Content == target.Content &&
		current.Summary == target.Summary &&
		current.CoverImage == target.CoverImage &&
		common.StringPtrEqual(current.Category, target.Category) &&
		common.StringPtrEqual(current.Tags, target.Tags)
}

func blogCreateFieldsMatch(current, target *model.Blog) bool {
	return blogFieldsMatch(current, target) && current.Author == target.Author
}

func blogPublishStatusMatch(current *model.Blog, status uint, publishedAt *time.Time) bool {
	if current.Status != status {
		return false
	}
	if status != model.BlogStatusPublished {
		return true
	}
	return current.PublishedAt != nil
}

/** 根据ID获取文章详情
 * @param blogId 文章ID
 * @return model.Blog, error
 */
func (*BlogRepository) GetBlogById(blogId string) (model.Blog, error) {
	var blog model.Blog
	err := common.DB.Where("id = ?", blogId).First(&blog).Error
	return blog, err
}

/** 更新文章状态和发布时间
 * @param blogId 文章ID
 * @param version 乐观锁版本号
 * @param status 文章状态
 * @param publishedAt 发布时间
 * @return duplicate, error
 */
func (*BlogRepository) UpdateBlogPublishStatusById(blogId string, version uint, status uint, publishedAt *time.Time) (bool, error) {
	var publishedAtValue interface{}
	if publishedAt == nil {
		publishedAtValue = nil
	} else {
		publishedAtValue = *publishedAt
	}
	updates := map[string]interface{}{
		"status":       status,
		"published_at": publishedAtValue,
		"version":      version + 1,
	}
	result := common.DB.Model(&model.Blog{}).Where("id = ? AND version = ?", blogId, version).Updates(updates)
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected > 0 {
		return false, nil
	}
	var current model.Blog
	if err := common.DB.Where("id = ?", blogId).First(&current).Error; err != nil {
		return false, err
	}
	return applyOptimisticLockResult(blogPublishStatusMatch(&current, status, publishedAt), result.RowsAffected, nil)
}

/** 更新文章
 * @param blogId 文章ID
 * @param version 乐观锁版本号
 * @param blog 文章信息
 * @return duplicate, error
 */
func (*BlogRepository) UpdateBlogById(blogId string, version uint, blog *model.Blog) (bool, error) {
	updates := map[string]interface{}{
		"title":       blog.Title,
		"content":     blog.Content,
		"summary":     blog.Summary,
		"cover_image": blog.CoverImage,
		"category":    blog.Category,
		"tags":        blog.Tags,
		"version":     version + 1,
	}
	result := common.DB.Model(&model.Blog{}).Where("id = ? AND version = ?", blogId, version).Updates(updates)
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected > 0 {
		return false, nil
	}
	var current model.Blog
	if err := common.DB.Where("id = ?", blogId).First(&current).Error; err != nil {
		return false, err
	}
	return applyOptimisticLockResult(blogFieldsMatch(&current, blog), result.RowsAffected, nil)
}

/** 搜索文章
 * @param req *vo.SearchBlogRequest 搜索文章请求
 * @return dto.SearchBlogDto, error
 */
func (*BlogRepository) SearchBlogs(req *vo.SearchBlogRequest, ctx *gin.Context) (*dto.SearchResultDTO, error) {
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
			CoverImage:  hit.Source.CoverImage,
			Category:    hit.Source.Category,
			Tags:        hit.Source.Tags,
			Status:      hit.Source.Status,
			Author:      hit.Source.Author,
			CreatedAt:   hit.Source.CreatedAt,
			UpdatedAt:   hit.Source.UpdatedAt,
			Content:     utils.HighlightOrFallback(hit.Highlight["content"], hit.Source.Content),
			PublishedAt: hit.Source.PublishedAt,
		}

		// 保留高亮信息
		if hl, ok := hit.Highlight["title"]; ok {
			post.Highlight.Title = hl
		}
		if hl, ok := hit.Highlight["summary"]; ok {
			post.Highlight.Summary = hl
		}
		if hl, ok := hit.Highlight["content"]; ok {
			post.Highlight.Content = hl
		}
		result.Hits = append(result.Hits, post)
	}

	enrichBlogSearchVersions(result.Hits)

	// 5. 添加拼写建议
	if len(esResp.Suggest.SpellCheck) > 0 &&
		len(esResp.Suggest.SpellCheck[0].Options) > 0 {
		result.Suggestion = esResp.Suggest.SpellCheck[0].Options[0].Text
	}

	return result, nil
}

// enrichBlogSearchVersions 为 ES 搜索结果补全 MySQL 中的乐观锁 version（ES 索引不含该字段）。
func enrichBlogSearchVersions(hits []dto.BlogPostSource) {
	if len(hits) == 0 {
		return
	}
	ids := make([]string, 0, len(hits))
	seen := make(map[string]struct{}, len(hits))
	for _, h := range hits {
		id := strings.TrimSpace(h.ID)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return
	}
	var rows []struct {
		ID      string `gorm:"column:id"`
		Version uint   `gorm:"column:version"`
	}
	if err := common.DB.Model(&model.Blog{}).Select("id", "version").Where("id IN ?", ids).Find(&rows).Error; err != nil {
		common.Log.Errorf("补全搜索结果的 version 失败: %v", err)
		return
	}
	verByID := make(map[string]uint, len(rows))
	for _, r := range rows {
		v := r.Version
		if v < 1 {
			v = 1
		}
		verByID[r.ID] = v
	}
	for i := range hits {
		if v, ok := verByID[hits[i].ID]; ok {
			hits[i].Version = v
		} else if hits[i].Version < 1 {
			hits[i].Version = 1
		}
	}
}
