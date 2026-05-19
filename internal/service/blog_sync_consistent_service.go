/*
 * @Date: 2026-05-18 14:15:24
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-05-18 22:29:15
 * @Description: MQ 消费失败后后 ES/CH 同步、对账与可补偿错误
 */
package service

import (
	"encoding/json"
	"errors"
	"go-blog/internal/common"
	"go-blog/internal/model"
	"go-blog/internal/repository"
)

type BlogSyncConsistentError struct {
	Blog        *model.Blog
	Body        []byte
	PendingMask uint8
	Cause       error
}

func (e *BlogSyncConsistentError) Error() string {
	if e.Cause != nil {
		return e.Cause.Error()
	}
	return "blog sync failed"
}

func (e *BlogSyncConsistentError) Unwrap() error {
	return e.Cause
}

/**
 * @description: 创建同步错误
 * @param blog *model.Blog 博客
 * @param mask uint8 位图
 * @param cause error 错误
 * @return *BlogSyncConsistentError 同步错误
 */
func newSyncError(blog *model.Blog, mask uint8, cause error) *BlogSyncConsistentError {
	return &BlogSyncConsistentError{
		Blog:        blog,
		PendingMask: mask,
		Cause:       cause,
	}
}

/**
 * @description: 对账并同步博客
 * @param blog *model.Blog 博客
 * @param pendingMask uint8 位图
 * @return error 错误
 */
func ReconcileAndSyncBlog(blog *model.Blog, pendingMask uint8) error {
	if blog == nil || blog.ID == "" {
		return newSyncError(nil, model.SyncPendingAll, errors.New("blog 无效"))
	}
	// 对账：返回仍待同步位图（ES=1、CH=2）；为 0 表示两侧均已存在，无需写入
	stillPending, err := repository.ReconcileSyncPendingMask(blog.ID, pendingMask)
	if err != nil {
		common.Log.Errorf("[同步对账] blog_id=%s 对账失败，沿用任务位图: %v", blog.ID, err)
		stillPending = pendingMask
	}
	if stillPending == 0 {
		common.Log.Infof("[同步对账] blog_id=%s ES/CH 均已存在，跳过写入", blog.ID)
		return nil
	}
	// 同步写入 ES/CH（错误内已携带剩余待同步位图）
	return syncBlogWithPendingMask(blog, stillPending)
}

/**
 * @description: 同步博客
 * @param blog *model.Blog 博客
 * @param mask uint8 位图
 * @return error 错误（*BlogSyncConsistentError.PendingMask 为剩余待同步位）
 */
func syncBlogWithPendingMask(blog *model.Blog, mask uint8) error {
	remaining := mask
	if model.NeedsSyncES(remaining) {
		if err := repository.IndexBlogToES(blog); err != nil {
			return newSyncError(blog, remaining, err)
		}
		remaining = model.MarkESSynced(remaining)
	}
	if model.NeedsSyncCH(remaining) {
		if err := repository.UpsertBlogToClickHouse(blog); err != nil {
			return newSyncError(blog, remaining, err)
		}
	}
	return nil
}

/**
 * @description: 补偿消费失败，重新入队
 * @param repo repository.IMQCompensationRepository 补偿仓库
 * @param syncErr *BlogSyncConsistentError 补偿错误
 * @return error 错误
 */
func enqueueBlogConsumeCompensation(repo repository.IMQCompensationRepository, syncErr *BlogSyncConsistentError) error {
	if syncErr == nil {
		return errors.New("syncErr 为空")
	}
	// 获取补偿位图
	mask := syncErr.PendingMask
	if mask == 0 {
		mask = model.SyncPendingAll
	}
	errMsg := syncErr.Error()
	if syncErr.Blog != nil && syncErr.Blog.ID != "" {
		return repo.EnqueueBlogConsume(syncErr.Blog, mask, errMsg)
	}
	if len(syncErr.Body) > 0 {
		var blog model.Blog
		if json.Unmarshal(syncErr.Body, &blog) == nil && blog.ID != "" {
			return repo.EnqueueBlogConsume(&blog, mask, errMsg)
		}
		return repo.EnqueueBlogConsumePayload("", syncErr.Body, mask, errMsg)
	}
	return errors.New("无法构造 CONSUME 补偿任务：缺少 blog 与 payload")
}

/**
 * @description: 日志同步对账
 * @param blogID string blogID
 * @param before uint8 之前位图
 * @param after uint8 之后位图
 */
func LogSyncReconcile(blogID string, before, after uint8) {
	if before == after {
		return
	}
	common.Log.Infof("[同步对账] blog_id=%s pending %s -> %s",
		blogID, model.MaskToLabel(before), model.MaskToLabel(after))
}

/**
 * @description: 格式化同步位图
 * @param mask uint8 位图
 * @return string 格式化后的位图
 */
func FormatSyncMask(mask uint8) string {
	s := model.MaskToLabel(mask)
	if s != "" {
		return s
	}
	return "DONE"
}

/**
 * @description: 获取同步错误位图
 * @param err error 错误
 * @return uint8 位图
 */
func SyncErrorPendingMask(err error) uint8 {
	var syncErr *BlogSyncConsistentError
	if errors.As(err, &syncErr) && syncErr.PendingMask != 0 {
		return syncErr.PendingMask
	}
	return 0
}
