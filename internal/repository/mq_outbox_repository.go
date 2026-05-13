/*
 * @Date: 2026-05-07 17:04:47
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-05-13 15:23:57
 * @Description: 博客MQ投递失败时落库并定时重试仓储
 */
package repository

import (
	"encoding/json"
	"errors"
	"go-blog/internal/common"
	"go-blog/internal/model"
)

const (
	OutboxStatusPending uint8 = 0 // 待投递
	OutboxStatusSent    uint8 = 1 // 已发送
	OutboxStatusDead    uint8 = 2 // 已放弃
)

/**
 * @description: 博客MQ投递失败时落库并定时重试仓储接口
 * @return {IMQOutboxRepository}
 */
type IMQOutboxRepository interface {
	EnqueueBlogPublish(blog *model.Blog, publishErr string) error      // 将博客投递到MQ
	ListPendingForRetry(limit int) ([]model.BlogMQOutbox, error)       // 查询待重试的博客
	MarkSent(id uint) error                                            // 标记博客已发送
	MarkRetry(id uint, retryCount int, errMsg string, dead bool) error // 标记博客重试
}

type MQOutboxRepository struct{}

func NewMQOutboxRepository() IMQOutboxRepository {
	return &MQOutboxRepository{}
}

/**
 * @description: 博客MQ投递失败时落库并定时重试仓储接口
 * @param {blog *model.Blog} blog
 * @param {publishErr string} publishErr
 * @return {error}
 */
func (*MQOutboxRepository) EnqueueBlogPublish(blog *model.Blog, publishErr string) error {
	if blog == nil || blog.ID == 0 {
		return errors.New("blog 无效")
	}
	body, err := json.Marshal(blog)
	if err != nil {
		return err
	}
	row := model.BlogMQOutbox{
		BlogID:     blog.ID,
		Payload:    body,
		Status:     OutboxStatusPending,
		RetryCount: 0,
		LastError:  publishErr,
	}
	return common.DB.Create(&row).Error
}

/**
 * @description: 博客MQ投递失败时落库并定时重试仓储接口
 * @param {limit int} limit
 * @return {[]model.BlogMQOutbox, error}
 */
func (*MQOutboxRepository) ListPendingForRetry(limit int) ([]model.BlogMQOutbox, error) {
	if limit <= 0 {
		limit = 50
	}
	var rows []model.BlogMQOutbox
	err := common.DB.Where("status = ?", OutboxStatusPending).
		Order("id ASC").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

/*
*
/**
  - @description: 博客MQ投递失败时落库并定时重试仓储接口
  - @param {id uint} id
  - @return {error}
*/
func (*MQOutboxRepository) MarkSent(id uint) error {
	return common.DB.Model(&model.BlogMQOutbox{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status": OutboxStatusSent,
		}).Error
}

/**
 * @description: 博客MQ投递失败时落库并定时重试仓储接口
 * @param {id uint} id
 * @param {retryCount int} retryCount
 * @param {errMsg string} errMsg
 * @param {dead bool} dead
 * @return {error}
 */
func (*MQOutboxRepository) MarkRetry(id uint, retryCount int, errMsg string, dead bool) error {
	st := OutboxStatusPending
	if dead {
		st = OutboxStatusDead
	}
	return common.DB.Model(&model.BlogMQOutbox{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"retry_count": retryCount,
			"last_error":  errMsg,
			"status":      st,
		}).Error
}
