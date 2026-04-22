package dto

import (
	"go-blog/internal/model"
	"time"
)

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
