/*
 * @Description: 创建/更新幂等与乐观锁仓库辅助
 */
package repository

import (
	"strings"

	"go-blog/internal/common"

	"gorm.io/gorm"
)

// createIdempotencyResolve 在 request_id 唯一冲突后加载已有记录并判定是否幂等重复。
type createIdempotencyResolve func() (duplicate bool, err error)

// handleCreateIdempotency 处理创建时的唯一约束冲突：resolve 内须比对业务字段，一致才视为幂等重复。
func handleCreateIdempotency(db *gorm.DB, requestID string, createErr error, resolve createIdempotencyResolve) (duplicate bool, err error) {
	if createErr == nil {
		return false, nil
	}
	if !common.IsDuplicateKeyError(createErr) {
		return false, createErr
	}
	if strings.TrimSpace(requestID) == "" {
		return false, createErr
	}
	if resolve == nil {
		return false, createErr
	}
	dup, resolveErr := resolve()
	if resolveErr != nil {
		return false, resolveErr
	}
	if dup {
		return true, nil
	}
	return false, common.ErrOptimisticLockConflict
}

// applyOptimisticLockResult 在乐观锁更新影响行数为 0 时，判断是幂等重复还是冲突。
func applyOptimisticLockResult(alreadyApplied bool, rowsAffected int64, lockErr error) (duplicate bool, err error) {
	if lockErr != nil {
		return false, lockErr
	}
	if rowsAffected > 0 {
		return false, nil
	}
	if alreadyApplied {
		return true, nil
	}
	return false, common.ErrOptimisticLockConflict
}
