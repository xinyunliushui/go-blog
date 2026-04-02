/*
 * @Date: 2026-04-01 10:05:45
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-01 10:29:16
 * @Description:
 */
package middleware

import (
	"fmt"
	"go-blog/internal/common"
	"go-blog/internal/config"
	"go-blog/internal/repository"
	"go-blog/internal/response"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

var checkLock sync.Mutex

func CasbinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ur := repository.NewUserRepository()
		user, err := ur.GetCurrentUser(c)
		if err != nil {
			response.Fail(c, nil, fmt.Sprintf("%v", err))
			c.Abort()
			return
		}
		// 获得用户的全部角色
		roles := user.Roles
		// 获得用户全部未被禁用的角色的Keyword
		var subs []string
		for _, role := range roles {
			if role.Status == 1 {
				subs = append(subs, role.Keyword)
			}
		}
		// 获得请求路径URL
		//obj := strings.Replace(c.Request.URL.Path, "/"+config.Conf.System.UrlPathPrefix, "", 1)
		obj := strings.TrimPrefix(c.FullPath(), "/"+config.Config.Application.UrlPathPrefix)
		// 获取请求方式
		act := c.Request.Method

		isPass := check(subs, obj, act)
		if !isPass {
			response.Response(c, 403, 403, nil, "禁止访问")
			c.Abort()
			return
		}
		c.Next()
	}
}

func check(subs []string, obj string, act string) bool {
	// 同一时间只允许一个请求执行校验, 否则可能会校验失败
	checkLock.Lock()
	defer checkLock.Unlock()
	isPass := false
	for _, sub := range subs {
		pass, _ := common.CasbinEnforcer.Enforce(sub, obj, act)
		if pass {
			isPass = true
			break
		}
	}
	return isPass
}
