/*
 * @Date: 2026-04-22 17:42:02
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-05-15
 * @Description: 文章DTO
 */
package dto

import (
	"go-blog/internal/model"
	"time"
)

// 文章DTO
type BlogDto struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	Summary     string     `json:"summary"`
	CoverImage  string     `json:"coverImage"`
	Category    *string    `json:"category"`
	Tags        *string    `json:"tags"`
	Status      uint       `json:"status"`
	Author      string     `json:"author"`
	PublishedAt *time.Time `json:"publishedAt"`
}

/** 将文章列表转换为文章DTO列表
 * @param blogs 文章列表
 * @return []BlogDto
 */
func ToBlogsDto(blogs []model.Blog) []BlogDto {
	var blogDtos []BlogDto
	for _, blog := range blogs {
		blogDtos = append(blogDtos, BlogDto{
			ID:          blog.ID,
			Title:       blog.Title,
			Content:     blog.Content,
			Summary:     blog.Summary,
			CoverImage:  blog.CoverImage,
			Category:    blog.Category,
			Tags:        blog.Tags,
			Status:      blog.Status,
			Author:      blog.Author,
			PublishedAt: blog.PublishedAt,
		})
	}
	return blogDtos
}

/***
 * @description: 文章源数据DTO 用户ES查询结果
 * @return {BlogPostSource}
 */
type BlogPostSource struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	Summary     string     `json:"summary"`
	CoverImage  string     `json:"cover_image"`
	Category    *string    `json:"category"`
	Tags        *string    `json:"tags"`
	Status      uint       `json:"status"`
	Author      string     `json:"author"`
	PublishedAt *time.Time `json:"published_at"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
	Highlight   struct {
		Title   []string `json:"title,omitempty"`
		Summary []string `json:"summary,omitempty"`
		Content []string `json:"content,omitempty"`
	} `json:"highlight"`
}

/**
 * @description: 搜索文章DTO
 * @return {SearchBlogDto}
 */
type SearchBlogDto struct {
	Took int64 `json:"took"`
	Hits struct {
		Total struct {
			Value int `json:"value"`
		} `json:"total"`
		Hits []struct {
			ID        string              `json:"_id"`
			Source    BlogPostSource      `json:"_source"`
			Highlight map[string][]string `json:"highlight,omitempty"`
		} `json:"hits"`
	} `json:"hits"`
	Suggest struct {
		SpellCheck []struct {
			Options []struct {
				Text string `json:"text"`
			} `json:"options"`
		} `json:"spell_check"`
	} `json:"suggest"`
}

/***
 * @description: 搜索结果DTO
 * @return {SearchResultDTO}
 */
type SearchResultDTO struct {
	Hits       []BlogPostSource `json:"hits"`
	Total      int              `json:"total"`
	Took       int64            `json:"took_ms"`              // ES查询耗时
	Suggestion string           `json:"suggestion,omitempty"` // 拼写建议
}
