/*
 * @Date: 2026-04-13 17:49:06
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-13 17:49:12
 * @Description: 公共方法
 */
package controller

import (
	"go-blog/internal/common"
	"go-blog/internal/response"

	"github.com/gin-gonic/gin"
)

/** 参数绑定和校验
 * @param ctx gin.Context上下文
 * @param req 请求参数结构体
 * @return void
 */
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
