/*
 * @Date: 2026-04-13 17:49:06
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-13 17:49:12
 * @Description:
 */
package controller

import (
	"go-blog/internal/common"
	"go-blog/internal/response"

	"github.com/gin-gonic/gin"
)

func BindAndValidate(ctx *gin.Context, req interface{}) {
	if err := ctx.ShouldBind(&req); err != nil {
		response.Fail(ctx, nil, common.ValidationErrString(err))
		return
	}
	// 参数校验
	if err := common.Validate.Struct(&req); err != nil {
		response.Fail(ctx, nil, common.ValidationErrString(err))
		return
	}
}
