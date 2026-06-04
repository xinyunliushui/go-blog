/*
 * @Description: 幂等请求 ID 解析、唯一约束与乐观锁辅助
 */
package common

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

const (
	// HeaderIdempotencyKey 客户端幂等键请求头（优先于 X-Request-Id）。
	HeaderIdempotencyKey = "X-Idempotency-Key"
)

// ErrOptimisticLockConflict 表示乐观锁冲突（数据已被他人修改且与本次提交不一致）。
var ErrOptimisticLockConflict = errors.New("optimistic lock conflict")

// IsDuplicateKeyError 判断是否为 MySQL 唯一约束冲突（1062）。
func IsDuplicateKeyError(err error) bool {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1062
	}
	return false
}

// ResolveRequestID 解析创建类接口的幂等请求 ID：body > X-Idempotency-Key > X-Request-Id。
func ResolveRequestID(c *gin.Context, bodyRequestID string) (string, error) {
	if id := strings.TrimSpace(bodyRequestID); id != "" {
		return id, nil
	}
	if c != nil {
		if id := strings.TrimSpace(c.GetHeader(HeaderIdempotencyKey)); id != "" {
			return id, nil
		}
		if id := strings.TrimSpace(c.GetHeader(HeaderTraceIDFallback)); id != "" {
			return id, nil
		}
	}
	return "", errors.New("requestId 不能为空，请在请求体 requestId 或请求头 X-Idempotency-Key 中传递")
}

// RequestIDPtr 将非空 requestId 转为可写入数据库的指针。
func RequestIDPtr(requestID string) *string {
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return nil
	}
	return &requestID
}

// FindByRequestID 按 request_id 查询已创建记录，用于唯一约束冲突后的幂等回放。
func FindByRequestID(db *gorm.DB, requestID string, dest interface{}) error {
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return gorm.ErrRecordNotFound
	}
	return db.Where("request_id = ?", requestID).First(dest).Error
}

// StringPtrEqual 比较两个可空字符串指针是否语义相等。
func StringPtrEqual(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

// StringSliceEqualUnordered 比较两个字符串切片是否包含相同元素（忽略顺序）。
func StringSliceEqualUnordered(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	counts := make(map[string]int, len(a))
	for _, s := range a {
		counts[s]++
	}
	for _, s := range b {
		counts[s]--
		if counts[s] < 0 {
			return false
		}
	}
	return true
}
