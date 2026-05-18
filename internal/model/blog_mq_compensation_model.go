/*
 * @Description: 博客 MQ 异步补偿（推送失败补 MQ / 消费后补 ES+CH）
 */
package model

import (
	"gorm.io/gorm"
)

const (
	TaskTypePublish = "PUBLISH"
	TaskTypeConsume = "CONSUME"
)

const (
	SyncPendingES  uint8                           = 1 << iota // 0
	SyncPendingCH                                              // 1
	SyncPendingAll = SyncPendingES | SyncPendingCH             // 按位或，表示 ES 和 CH 都待同步
)

const (
	SyncLabelES  = "ES"
	SyncLabelCH  = "CH"
	SyncLabelAll = "ALL"
)

type BlogMQCompensation struct {
	UUIDModel
	BlogID      string `gorm:"type:char(36);not null;index:idx_blog_task_status,priority:1;comment:关联 blogs.id" json:"blogId"`
	TaskType    string `gorm:"type:varchar(16);not null;default:PUBLISH;index:idx_blog_task_status,priority:2;comment:PUBLISH|CONSUME" json:"taskType"`
	PendingMask uint8  `gorm:"not null;default:0;comment:待同步位 ES=1 CH=2 ALL=3" json:"pendingMask"`
	Payload     []byte `gorm:"type:longtext;not null;comment:Blog消息体(JSON格式)" json:"payload"`
	Status      uint8  `gorm:"not null;default:0;index:idx_blog_task_status,priority:3;comment:0待处理 1已成功 2已放弃" json:"status"`
	RetryCount  int    `gorm:"not null;default:0;comment:重试次数" json:"retryCount"`
	LastError   string `gorm:"type:varchar(1024);comment:最近一次失败原因" json:"lastError"`
}

func (b *BlogMQCompensation) BeforeCreate(tx *gorm.DB) error {
	EnsureUUID(&b.ID)
	if b.TaskType == "" {
		b.TaskType = TaskTypePublish
	}
	if b.TaskType == TaskTypeConsume && b.PendingMask == 0 {
		b.PendingMask = SyncPendingAll
	}
	return nil
}

func (b *BlogMQCompensation) TableName() string {
	return "blog_mq_compensation"
}

/**
 * 返回有效待同步位图
 */
func (b *BlogMQCompensation) EffectivePendingMask() uint8 {
	if b.PendingMask != 0 {
		return b.PendingMask
	}
	if b.TaskType == TaskTypeConsume {
		return SyncPendingAll
	}
	return 0
}

/**
 * 将位图转换为可读标签，用于日志中
 */
func MaskToLabel(mask uint8) string {
	switch mask {
	case SyncPendingES:
		return SyncLabelES
	case SyncPendingCH:
		return SyncLabelCH
	case SyncPendingAll:
		return SyncLabelAll
	default:
		return ""
	}
}

/**
 * 合并待同步位图
 */
func MergeSyncMask(existing, incoming uint8) uint8 {
	return existing | incoming
}

// NeedsSyncES 是否仍需同步 Elasticsearch。
func NeedsSyncES(mask uint8) bool {
	return mask == SyncPendingES || mask == SyncPendingAll
}

// NeedsSyncCH 是否仍需同步 ClickHouse。
func NeedsSyncCH(mask uint8) bool {
	return mask == SyncPendingCH || mask == SyncPendingAll
}

// MarkESSynced 将 ES 标为已同步，返回剩余待同步位图。
func MarkESSynced(mask uint8) uint8 {
	switch mask {
	case SyncPendingES:
		return 0
	case SyncPendingAll:
		return SyncPendingCH
	default:
		return mask
	}
}

// MarkCHSynced 将 CH 标为已同步，返回剩余待同步位图。
func MarkCHSynced(mask uint8) uint8 {
	switch mask {
	case SyncPendingCH:
		return 0
	case SyncPendingAll:
		return SyncPendingES
	default:
		return mask
	}
}
