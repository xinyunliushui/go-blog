/*
 * @Date: 2026-03-31 17:19:10
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-24 10:56:57
 * @Description: standardized response
 */
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

/** 返回前端
 * @param c *gin.Context 上下文
 * @param httpStatus int HTTP状态码
 * @param code int 响应码
 * @param data interface{} 响应数据
 * @param message string 响应消息
 * @return void
 */
func Response(c *gin.Context, httpStatus int, code int, data interface{}, message string) {
	c.JSON(httpStatus, gin.H{"code": code, "data": data, "message": message})
}

/** 返回前端-成功
 * @param c *gin.Context 上下文
 * @param data interface{} 响应数据
 * @param message string 响应消息
 * @return void
 */
func Success(c *gin.Context, data interface{}, message string) {
	Response(c, http.StatusOK, 200, data, message)
}

/** 返回前端-失败
 * @param c *gin.Context 上下文
 * @param data interface{} 响应数据
 * @param message string 响应消息
 * @return void
 */
func Fail(c *gin.Context, data interface{}, message string) {
	Response(c, http.StatusBadRequest, 500, data, message)
}
