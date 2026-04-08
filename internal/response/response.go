/*
 * @Date: 2026-03-31 17:19:10
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-03-31 17:19:14
 * @Description: standardized response
 */
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 返回前端
func Response(c *gin.Context, httpStatus int, code int, data interface{}, message string) {
	c.JSON(httpStatus, gin.H{"code": code, "data": data, "message": message})
}

// 返回前端-成功
func Success(c *gin.Context, data interface{}, message string) {
	Response(c, http.StatusOK, 200, data, message)
}

// 返回前端-失败
func Fail(c *gin.Context, data interface{}, message string) {
	Response(c, http.StatusBadRequest, 400, data, message)
}
