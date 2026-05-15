/*
 * @Date: 2026-05-04 16:35:28
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-05-15
 * @Description: 将博客索引到Elasticsearch
 */
package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go-blog/internal/config"
	"go-blog/internal/elasticsearch"
	"go-blog/internal/model"
	"go-blog/internal/vo"
	"time"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

/**
 * @description: 将博客索引到Elasticsearch
 * @param {model.Blog} blog
 * @return {error}
 */
func IndexBlogToES(blog *model.Blog) error {
	// 转换为ES文档结构
	doc := map[string]interface{}{
		"id":           blog.ID,
		"title":        blog.Title,
		"content":      blog.Content,
		"summary":      blog.Summary,
		"cover_image":  blog.CoverImage,
		"category":     blog.Category,
		"tags":         blog.Tags,
		"status":       blog.Status,
		"author":       blog.Author,
		"published_at": blog.PublishedAt,
		"created_at":   blog.CreatedAt,
		"updated_at":   blog.UpdatedAt,
	}

	// 序列化文章为 JSON
	body, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("序列化文章失败: %w", err)
	}

	req := esapi.IndexRequest{
		Index:      config.Conf.ElasticSearch.IndexName,
		DocumentID: blog.ID,
		Body:       bytes.NewReader(body),
		Refresh:    "false", // 生产环境按需调整刷新策略，写入性能优先
	}

	res, err := req.Do(context.Background(), elasticsearch.ESClient)
	if err != nil {
		return fmt.Errorf("执行索引请求失败: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		var errResp map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&errResp); err != nil {
			return fmt.Errorf("索引请求返回非成功状态码: %s, 且解析错误响应失败: %w", res.Status(), err)
		}
		return fmt.Errorf("索引请求返回非成功状态码: %s, 响应: %v", res.Status(), errResp)
	}

	return nil
}

/***
 * @description: 构建博客搜索查询DSL
 * @param {vo.SearchBlogRequest} req
 * @return {string, error}
 */
func BuildBlogSearchQueryDSL(req vo.SearchBlogRequest) (string, error) {
	// 1. 多字段加权搜索
	multiMatch := map[string]interface{}{
		"query":    req.Keyword,
		"analyzer": "blog_analyzer",
		"type":     "best_fields",
		"fields": []string{
			"title^3.0",   // 标题权重最高
			"summary^2.0", // 摘要次之
			"content^1.0", // 内容基础权重
		},
		"minimum_should_match": "75%", // 至少匹配75%的词项
	}

	// 完整DSL构建
	from := (req.Page - 1) * req.PageSize
	dsl := map[string]interface{}{
		"from": from,
		"size": req.PageSize,
		"query": map[string]interface{}{
			"function_score": map[string]interface{}{
				"query": map[string]interface{}{
					"bool": map[string]interface{}{
						"must": []interface{}{
							map[string]interface{}{"multi_match": multiMatch},
							map[string]interface{}{
								"term": map[string]interface{}{
									"status": model.BlogStatusPublished, // 只搜索发布状态的文章
								},
							},
						},
					},
				},
				// 时间衰减函数：新内容优先
				"functions": []map[string]interface{}{
					{
						"exp": map[string]interface{}{
							"created_at": map[string]interface{}{
								"origin": "now",
								"offset": "7d",
								"scale":  "30d",
								"decay":  0.33,
							},
						},
						"weight": 1.5,
					},
				},
				"boost_mode": "multiply",
			},
		},
		// 高亮配置（title/summary 与 content 均使用 <mark>，与前端 .hitHtml 一致）
		"highlight": map[string]interface{}{
			"fields": map[string]interface{}{
				"title": map[string]interface{}{
					"type":          "fvh",
					"fragment_size": 120,
					"pre_tags":      []string{"<mark class='hl'>"},
					"post_tags":     []string{"</mark>"},
				},
				"summary": map[string]interface{}{
					"type":          "fvh",
					"fragment_size": 280,
					"pre_tags":      []string{"<mark class='hl'>"},
					"post_tags":     []string{"</mark>"},
				},
				"content": map[string]interface{}{
					"type":                "fvh",
					"number_of_fragments": 2,   // 内容分段数
					"fragment_size":       150, // 适合代码片段的长度
					"pre_tags":            []string{"<mark class='hl'>"},
					"post_tags":           []string{"</mark>"},
				},
			},
		},
		// 拼写检查
		"suggest": map[string]interface{}{
			"spell_check": map[string]interface{}{
				"text": req.Keyword,
				"term": map[string]interface{}{
					"field": "content",
				},
			},
		},
	}
	// 转为JSON
	body, err := json.Marshal(dsl)
	if err != nil {
		return "", fmt.Errorf("构建博客搜索查询DSL失败: %w", err)
	}
	return string(body), nil
}

/***
 * @description: 更新博客正文变更后同步 ES（partial update），字段与 IndexBlogToES 中可被编辑部分一致；不改变 status/author/published_at。
 * @param {string} blogID
 * @param {map[string]interface{}} doc
 * @return {error}
 */
func UpdateBlogFieldsInES(blogID string, doc map[string]interface{}) error {
	payload := map[string]interface{}{"doc": doc}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化 ES partial update 失败: %w", err)
	}

	req := esapi.UpdateRequest{
		Index:      config.Conf.ElasticSearch.IndexName,
		DocumentID: blogID,
		Body:       bytes.NewReader(body),
		Refresh:    "false",
	}

	res, err := req.Do(context.Background(), elasticsearch.ESClient)
	if err != nil {
		return fmt.Errorf("执行 ES Update 请求失败: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		var errResp map[string]interface{}
		if decErr := json.NewDecoder(res.Body).Decode(&errResp); decErr != nil {
			return fmt.Errorf("ES Update 失败 [%s]，解析响应失败: %w", res.Status(), decErr)
		}
		return fmt.Errorf("ES Update 失败 [%s]: %v", res.Status(), errResp)
	}

	return nil
}

/***
 * @description: 更新博客发布状态到ES，仅更新发布状态和更新时间
 * @param {string} blogID
 * @param {uint} status
 * @param {*time.Time} publishedAt
 * @return {error}
 */
func UpdateBlogPublishStatusInES(blogID string, status uint, publishedAt *time.Time) error {
	if blogID == "" {
		return fmt.Errorf("blogID 无效")
	}
	doc := map[string]interface{}{
		"status": status,
	}
	if publishedAt != nil {
		doc["published_at"] = publishedAt
	} else {
		doc["published_at"] = nil
	}

	return UpdateBlogFieldsInES(blogID, doc)
}

/***
 * @description: 更新博客正文变更后同步 ES（partial update），字段与 IndexBlogToES 中可被编辑部分一致；不改变 status/author/published_at。
 * @param {string} blogID
 * @param {*model.Blog} blog
 * @return {error}
 */
func UpdateBlogFieldsInESByBlog(blogID string, blog *model.Blog) error {
	if blog == nil || blogID == "" {
		return fmt.Errorf("blogID 或 blog 无效")
	}
	doc := map[string]interface{}{
		"title":       blog.Title,
		"content":     blog.Content,
		"summary":     blog.Summary,
		"cover_image": blog.CoverImage,
		"category":    blog.Category,
		"tags":        blog.Tags,
	}
	return UpdateBlogFieldsInES(blogID, doc)
}
