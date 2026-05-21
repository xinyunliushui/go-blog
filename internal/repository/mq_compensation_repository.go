/*
 * @Date: 2026-05-18 15:47:03
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-05-21 23:41:57
 * @Description: 博客 MQ 补偿仓储（PUBLISH 补 MQ / CONSUME 补 ES+CH）
 */
package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"go-blog/internal/common"
	"go-blog/internal/model"
	"strings"

	"gorm.io/gorm"
)

const (
	CompensationStatusPending uint8 = 0 // 补偿状态：待重试
	CompensationStatusSent    uint8 = 1 // 补偿状态：已发送
	CompensationStatusDead    uint8 = 2 // 补偿状态：死亡
)

/**
 * @description: 博客 MQ 补偿仓储接口
 * @return IMQCompensationRepository 博客 MQ 补偿仓储接口
 */
type IMQCompensationRepository interface {
	EnqueueBlogPublish(blog *model.Blog, publishErr, traceID string) error                                     // 发布补偿入库
	EnqueueBlogConsume(blog *model.Blog, pendingMask uint8, syncErr, traceID string) error                     // 消费补偿入库
	EnqueueBlogConsumePayload(blogID string, payload []byte, pendingMask uint8, syncErr, traceID string) error // 消费补偿入库（payload 版）
	ListPendingForRetry(limit int) ([]model.BlogMQCompensation, error)                                         // 查询待重试的补偿记录
	MarkSent(id string) error                                                                                  // 标记补偿已发送
	MarkRetry(id string, retryCount int, errMsg string, dead bool, pendingMask uint8) error                    // 标记补偿重试
}

type MQCompensationRepository struct{}

func NewMQCompensationRepository() IMQCompensationRepository {
	return &MQCompensationRepository{}
}

/**
 * @description: 序列化博客为 JSON
 * @param blog *model.Blog 博客
 * @return []byte 序列化后的博客
 * @return error 错误
 */
func marshalBlogPayload(blog *model.Blog) ([]byte, error) {
	if blog == nil || blog.ID == "" {
		return nil, errors.New("blog 无效")
	}
	return json.Marshal(blog)
}

/**
 * @description: 入队发布补偿
 * @param blog *model.Blog 博客
 * @param publishErr string 发布错误
 * @return error 错误
 */
func (*MQCompensationRepository) EnqueueBlogPublish(blog *model.Blog, publishErr, traceID string) error {
	body, err := marshalBlogPayload(blog)
	if err != nil {
		return err
	}
	return upsertPendingCompensation(blog.ID, model.TaskTypePublish, 0, body, publishErr, traceID)
}

/**
 * @description: 入队消费补偿
 * @param blog *model.Blog 博客
 * @param pendingMask uint8 待同步位图
 * @param syncErr string 同步错误
 * @return error 错误
 */
func (*MQCompensationRepository) EnqueueBlogConsume(blog *model.Blog, pendingMask uint8, syncErr, traceID string) error {
	body, err := marshalBlogPayload(blog)
	if err != nil {
		return err
	}
	return upsertPendingCompensation(blog.ID, model.TaskTypeConsume, normalizeSyncMask(pendingMask), body, syncErr, traceID)
}

/**
 * @description: 入队消费补偿（payload 版）
 * @param blogID string 博客ID
 * @param payload []byte 消息体
 * @param pendingMask uint8 待同步位图
 * @param syncErr string 同步错误
 * @return error 错误
 */
func (*MQCompensationRepository) EnqueueBlogConsumePayload(blogID string, payload []byte, pendingMask uint8, syncErr, traceID string) error {
	if len(payload) == 0 {
		return errors.New("payload 为空")
	}
	if blogID == "" {
		blogID = "unknown"
	}
	return upsertPendingCompensation(blogID, model.TaskTypeConsume, normalizeSyncMask(pendingMask), payload, syncErr, traceID)
}

/**
 * @description: 更新补偿记录
 * @param blogID string 博客ID
 * @param taskType string 任务类型
 * @param pendingMask uint8 待同步位图
 * @param payload []byte 消息体
 * @param errMsg string 错误信息
 * @return error 错误
 */
func upsertPendingCompensation(blogID, taskType string, pendingMask uint8, payload []byte, errMsg, traceID string) error {
	traceID = common.ResolveTraceIDForCompensation(traceID)

	var existing model.BlogMQCompensation
	err := common.DB.Where(
		"blog_id = ? AND task_type = ? AND status = ?",
		blogID, taskType, CompensationStatusPending,
	).First(&existing).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		row := model.BlogMQCompensation{
			BlogID:      blogID,
			TraceID:     traceID,
			TaskType:    taskType,
			PendingMask: pendingMask,
			Payload:     payload,
			Status:      CompensationStatusPending,
			LastError:   errMsg,
		}
		return common.DB.Create(&row).Error
	}
	if err != nil {
		return err
	}

	updates := map[string]interface{}{
		"payload":    payload,
		"last_error": errMsg,
	}
	if strings.TrimSpace(existing.TraceID) == "" {
		updates["trace_id"] = traceID
	}
	if taskType == model.TaskTypeConsume {
		updates["pending_mask"] = model.MergeSyncMask(existing.EffectivePendingMask(), pendingMask)
	}
	// 更新补偿记录
	return common.DB.Model(&existing).Updates(updates).Error
}

/**
 * @description: 归一化同步位图
 * @param mask uint8 待同步位图
 * @return uint8 归一化后的同步位图
 */
func normalizeSyncMask(mask uint8) uint8 {
	if mask == 0 {
		return model.SyncPendingAll
	}
	return mask & model.SyncPendingAll
}

/**
 * @description: 查询待重试的补偿记录
 * @param limit int 查询数量
 * @return []model.BlogMQCompensation 补偿记录
 * @return error 错误
 */
func (*MQCompensationRepository) ListPendingForRetry(limit int) ([]model.BlogMQCompensation, error) {
	if limit <= 0 {
		limit = 50
	}
	var rows []model.BlogMQCompensation
	err := common.DB.Where("status = ?", CompensationStatusPending).
		Order("created_at ASC").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

/**
 * @description: 标记补偿已发送
 * @param id string 补偿ID
 * @return error 错误
 */
func (*MQCompensationRepository) MarkSent(id string) error {
	return common.DB.Model(&model.BlogMQCompensation{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       CompensationStatusSent,
			"pending_mask": 0,
		}).Error
}

/**
 * @description: 标记补偿重试
 * @param id string 补偿ID
 * @param retryCount int 重试次数
 * @param errMsg string 错误信息
 * @param dead bool 是否死亡
 * @param pendingMask uint8 待同步位图
 * @return error 错误
 */
func (*MQCompensationRepository) MarkRetry(id string, retryCount int, errMsg string, dead bool, pendingMask uint8) error {
	st := CompensationStatusPending
	if dead {
		st = CompensationStatusDead
	}
	updates := map[string]interface{}{
		"retry_count": retryCount,
		"last_error":  errMsg,
		"status":      st,
	}
	if pendingMask != 0 {
		updates["pending_mask"] = normalizeSyncMask(pendingMask)
	}
	res := common.DB.Model(&model.BlogMQCompensation{}).Where("id = ?", id).Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("补偿记录不存在: %s", id)
	}
	return nil
}
