/*
 * @Date: 2026-05-15 12:46:03
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-05-15 13:01:23
 * @Description:
 */
package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UUIDModel 替代 gorm.Model：字符串主键 + 软删，由 BeforeCreate 填充 UUID。
type UUIDModel struct {
	ID        string         `gorm:"type:char(36);primaryKey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Version   uint           `gorm:"not null;default:1;comment:'乐观锁版本号'" json:"version"`
	RequestId *string        `gorm:"type:varchar(64);uniqueIndex;comment:'创建幂等请求ID'" json:"requestId,omitempty"`
}

// EnsureUUID 若 id 为空则生成 UUID（供各模型 BeforeCreate 调用）。
func EnsureUUID(id *string) {
	if id != nil && *id == "" {
		*id = uuid.New().String()
	}
}
