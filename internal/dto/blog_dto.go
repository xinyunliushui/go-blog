/*
 * @Date: 2026-04-22 17:42:02
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-24 10:51:03
 * @Description: 文章DTO
 */
package dto

import (
	"go-blog/internal/model"
	"time"
)

// 文章DTO
type BlogDto struct {
	ID          uint       `json:"id"`
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
