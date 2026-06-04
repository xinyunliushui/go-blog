/*
 * @Description: 写操作（创建/更新）统一响应处理
 */
package controller

import (
	"errors"

	"go-blog/internal/common"
	"go-blog/internal/response"

	"github.com/gin-gonic/gin"
)

// msgStaleDataRetry 乐观锁冲突（HTTP 409）时返回给前端的提示。
const msgStaleDataRetry = "数据已被修改，请刷新页面后重试"

// msgCreateDuplicate 创建幂等重复（HTTP 200）时的提示，与 stale 文案区分以便前端正常视为成功。
const msgCreateDuplicate = "创建成功（重复请求）"

// msgUpdateDuplicate 更新幂等重复（HTTP 200）时的提示。
const msgUpdateDuplicate = "更新成功（重复请求）"

func requireVersion(c *gin.Context, version uint) bool {
	if version == 0 {
		response.Fail(c, nil, "version不能为空，请先刷新页面获取最新version后再提交")
		return false
	}
	return true
}

func handleWriteError(c *gin.Context, failMessage string, err error) {
	if errors.Is(err, common.ErrOptimisticLockConflict) {
		response.Conflict(c, nil, msgStaleDataRetry)
		return
	}
	response.FailErr(c, nil, failMessage, err)
}
