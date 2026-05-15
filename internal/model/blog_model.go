/*
 * @Date: 2026-04-22 15:33:08
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-05-15
 * @Description: 博客模型
 */
package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	BlogStatusDraft     = 1 // 草稿
	BlogStatusPublished = 2 // 发布
	BlogStatusPrivate   = 3 // 私密
)

// 博客模型结构体
type Blog struct {
	UUIDModel
	Title       string     `gorm:"type:varchar(100);not null;comment:'标题'"  json:"title"`
	Content     string     `gorm:"type:mediumtext;not null;comment:'内容'" json:"content"`
	Summary     string     `gorm:"type:varchar(500);not null;comment:'摘要'" json:"summary"`
	CoverImage  string     `gorm:"type:varchar(255);not null;comment:'封面图片'" json:"coverImage"`
	Category    *string    `gorm:"type:varchar(255);comment:'分类'" json:"category"`
	Tags        *string    `gorm:"type:varchar(255);comment:'标签'" json:"tags"`
	Status      uint       `gorm:"type:tinyint(1);default:1;not null;comment:'1草稿, 2发布, 3私密'" json:"status"`
	Author      string     `gorm:"type:varchar(100);not null;comment:'文章作者'" json:"author"`
	PublishedAt *time.Time `gorm:"type:datetime;comment:'发布时间'" json:"publishedAt"`
}

func (b *Blog) BeforeCreate(tx *gorm.DB) error {
	EnsureUUID(&b.ID)
	return nil
}
