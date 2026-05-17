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

	"go-blog/internal/common"

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

/** 返回前端-失败并记录底层错误
 * @param c *gin.Context 上下文
 * @param data interface{} 响应数据
 * @param message string 响应消息
 * @param err error 底层错误
 * @return void
 */
func FailErr(c *gin.Context, data interface{}, message string, err error) {
	if err != nil {
		logAPIFail(c, message, err)
	}
	Response(c, http.StatusBadRequest, 500, data, message)
}

/** 记录API失败日志
 * @param c *gin.Context 上下文
 * @param message string 响应消息
 * @param err error 底层错误
 * @return void
 */
func logAPIFail(c *gin.Context, message string, err error) {
	if common.Log == nil {
		return
	}
	path := c.FullPath()
	if path == "" {
		path = c.Request.URL.Path
	}
	common.Log.Errorf("[%s] %s - %s: %v", c.Request.Method, path, message, err)
}
