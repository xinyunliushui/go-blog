package model

import (
	"gorm.io/gorm"
)

// BlogMQOutbox 博客异步投递 RabbitMQ 失败时的本地补偿记录（Outbox）。
type BlogMQOutbox struct {
	gorm.Model
	BlogID     uint   `gorm:"not null;index;comment:关联 blogs.id"`
	Payload    []byte `gorm:"type:longtext;not null;comment:Blog JSON，用于重试 Publish"`
	Status     uint8  `gorm:"not null;default:0;index;comment:0待投递 1已成功 2已放弃"`
	RetryCount int    `gorm:"not null;default:0"`
	LastError  string `gorm:"type:varchar(1024);comment:最近一次投递失败原因"`
}

// TableName 指定表名
func (BlogMQOutbox) TableName() string {
	return "blog_mq_outbox"
}
