/*
 * @Date: 2026-03-31 17:07:35
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-21 15:56:31
 * @Description: gin-jwt认证中间件
 */
package middleware

import (
	"errors"
	"go-blog/internal/config"
	"go-blog/internal/model"
	"go-blog/internal/repository"
	"go-blog/internal/response"
	"go-blog/internal/utils"
	"go-blog/internal/vo"
	"net/http"
	"strings"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

// 初始化jwt中间件
func InitAuth() (*jwt.GinJWTMiddleware, error) {
	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:           config.Config.Jwt.Realm,                                 // jwt标识
		Key:             []byte(config.Config.Jwt.Key),                           // 服务端密钥
		Timeout:         time.Hour * time.Duration(config.Config.Jwt.Timeout),    // token过期时间
		MaxRefresh:      time.Hour * time.Duration(config.Config.Jwt.MaxRefresh), // token最大刷新时间(RefreshToken过期时间=Timeout+MaxRefresh)
		PayloadFunc:     payloadFunc,                                             // 有效载荷处理
		IdentityHandler: identityHandler,                                         // 解析Claims
		Authenticator:   login,                                                   // 校验token的正确性, 处理登录逻辑
		Authorizator:    authorizator,                                            // 用户登录校验成功处理
		Unauthorized:    unauthorized,                                            // 用户登录校验失败处理
		LoginResponse:   loginResponse,                                           // 登录成功后的响应
		LogoutResponse:  logoutResponse,                                          // 登出后的响应
		RefreshResponse: refreshResponse,                                         // 刷新token后的响应
		TokenLookup:     "header: Authorization, query: jwt, cookie: jwt",        // 自动在这几个地方寻找请求中的token
		TokenHeadName:   "Bearer",                                                // header名称
		TimeFunc:        time.Now,
	})
	return authMiddleware, err
}

// 有效载荷处理，自定义JWT Claims
func payloadFunc(data interface{}) jwt.MapClaims {
	if v, ok := data.(map[string]interface{}); ok {
		var user model.User
		utils.JsonI2Struct(v["user"], &user)
		return jwt.MapClaims{
			jwt.IdentityKey: user.ID,
			"user":          v["user"],
		}
	}
	return jwt.MapClaims{}
}

// 身份处理：解析jwt的Claims并提取用户信息，返回值类型map[string]interface{}与payloadFunc和authorizator的data类型必须一致, 否则会导致授权失败还不容易找到原因
func identityHandler(c *gin.Context) interface{} {
	claims := jwt.ExtractClaims(c)
	return map[string]interface{}{
		"IdentityKey": claims[jwt.IdentityKey],
		"user":        claims["user"],
	}
}

// 认证器即登录：验证用户名/密码
func login(c *gin.Context) (interface{}, error) {
	var req vo.RegisterAndLoginRequest
	// 请求json绑定
	if err := c.ShouldBindJSON(&req); err != nil {
		return "", jwt.ErrMissingLoginValues
	}

	// 密码通过RSA解密
	decodeData, err := utils.RSADecrypt([]byte(req.Password), config.Config.Application.RSAPrivateBytes)
	if err != nil {
		return nil, err
	}

	u := &model.User{
		Username: req.Username,
		Password: string(decodeData),
	}

	// 密码校验
	userRepository := repository.NewUserRepository()
	user, err := userRepository.Login(u)
	if err != nil {
		return nil, errors.New("用户名或密码不正确")
	}
	// 重要 将用户以json格式写入, payloadFunc/authorizator会使用到
	return map[string]interface{}{
		"user": utils.Struct2Json(user),
	}, nil
}

// 用户登录校验成功处理
func authorizator(data interface{}, c *gin.Context) bool {
	if v, ok := data.(map[string]interface{}); ok {
		userStr := v["user"].(string)
		var user model.User
		// 将用户json转为结构体
		utils.Json2Struct(userStr, &user)
		// 将用户保存到context, api调用时取数据方便
		c.Set("user", user)
		return true
	}
	return false
}

// 用户登录校验失败处理
func unauthorized(c *gin.Context, code int, message string) {
	// common.Log.Debugf("JWT认证失败, 错误码: %d, 错误信息: %s", code, message)
	response.Response(c, code, code, nil, message)
}

func requestIsHTTPS(c *gin.Context) bool {
	if c.Request.TLS != nil {
		return true
	}
	return strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https")
}

// 登录成功后的响应
func loginResponse(c *gin.Context, code int, token string, expires time.Time) {
	maxAgeSeconds := int((time.Hour * time.Duration(config.Config.Jwt.Timeout)).Seconds())
	secure := requestIsHTTPS(c)
	// 与 SPA 同机开发时请用 localhost 访问前后端，避免 localhost 与 127.0.0.1 混用导致浏览器不写 Cookie
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("jwt", token, maxAgeSeconds, "/", "", secure, true)
	response.Success(c,
		gin.H{
			"token":   token,
			"expires": expires.Unix(),
		},
		"登录成功")
}

// 登出后的响应
func logoutResponse(c *gin.Context, code int) {
	response.Success(c, nil, "退出成功")
}

// 刷新token后的响应
func refreshResponse(c *gin.Context, code int, token string, expires time.Time) {
	response.Response(c, code, code,
		gin.H{
			"token":   token,
			"expires": expires,
		},
		"刷新token成功")
}
