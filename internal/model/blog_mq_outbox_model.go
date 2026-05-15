/*
 * @Date: 2026-05-07 17:04:46
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-05-15 15:24:02
 * @Description: 博客异步投递 RabbitMQ 失败时的本地补偿记录（Outbox）。
 */
package model

import (
	"gorm.io/gorm"
)

type BlogMQOutbox struct {
	UUIDModel
	BlogID     string `gorm:"type:char(36);not null;index;comment:关联 blogs.id"`
	Payload    []byte `gorm:"type:longtext;not null;comment:Blog JSON，用于重试 Publish"`
	Status     uint8  `gorm:"not null;default:0;index;comment:0待投递 1已成功 2已放弃"`
	RetryCount int    `gorm:"not null;default:0"`
	LastError  string `gorm:"type:varchar(1024);comment:最近一次投递失败原因"`
}

func (b *BlogMQOutbox) BeforeCreate(tx *gorm.DB) error {
	EnsureUUID(&b.ID)
	return nil
}

// TableName 指定表名
func (b *BlogMQOutbox) TableName() string {
	return "blog_mq_outbox"
}
