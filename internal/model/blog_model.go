package model

import (
	"time"

	"gorm.io/gorm"
)

type Blog struct {
	gorm.Model
	Title       string     `gorm:"type:varchar(100);not null;comment:'标题'"  json:"title"`
	Content     string     `gorm:"type:mediumtext;not null;comment:'内容'" json:"content"`
	Summary     string     `gorm:"type:varchar(500);not null;comment:'摘要'" json:"summary"`
	CoverImage  string     `gorm:"type:varchar(255);not null;comment:'封面图片'" json:"cover_image"`
	Category    *string    `gorm:"type:varchar(255);comment:'分类'" json:"category"`
	Tags        *string    `gorm:"type:varchar(255);comment:'标签'" json:"tags"`
	Status      uint       `gorm:"type:tinyint(1);default:1;not null;comment:'1草稿, 2发布, 3私密'" json:"status"`
	IsTop       bool       `gorm:"type:tinyint(1);default:1;not null;comment:'1正常, 2置顶'" json:"is_top"`
	Author      string     `gorm:"type:varchar(100);not null;comment:'文章作者'" json:"author"`
	PublishedAt *time.Time `gorm:"type:datetime;comment:'发布时间'" json:"published_at"`
}
