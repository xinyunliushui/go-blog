/*
 * @Date: 2026-03-23 21:59:35
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-01 08:58:16
 * @Description:
 */
package config

import (
	"fmt"

	"github.com/jinzhu/configor"
)

var Config = struct {
	Application struct {
		Port            int    `default:"8080"`
		UrlPathPrefix   string `yaml:"url_path_prefix" default:"/"`
		RSAPublicBytes  string `yaml:"rsa_public_key"`
		RSAPrivateBytes string `yaml:"rsa_private_key"`
		// 非空时仅允许列表内 Origin 携带凭证跨域；为空则回显请求 Origin（勿在生产依赖此行为）
		CorsAllowOrigins []string `yaml:"cors_allow_origins"`
	}
	Mysql struct {
		Username    string
		Password    string
		Host        string
		Port        int
		Database    string
		Query       string
		LogMode     bool   `yaml:"log_mode"`
		TablePrefix string `yaml:"table_prefix" default:""`
		Charset     string
		Collation   string
	}
	Casbin struct {
		ModelPath string `yaml:"model_path"`
	}
	Jwt struct {
		Realm      string
		Key        string
		Timeout    int
		MaxRefresh int `yaml:"max_refresh"`
	}
	Logs struct {
		Level      int
		Path       string
		MaxSize    int
		MaxBackups int
		MaxAge     int
		Compress   bool
	}
}{}

/**
 * @description: 初始化配置
 * @return: void
 */
func InitConfig() {
	err := configor.Load(&Config, "config.yml")
	if err != nil {
		fmt.Printf("初始化配置失败: %s\n", err)
	}
	fmt.Printf("初始化配置成功: %#v\n", Config)
}
