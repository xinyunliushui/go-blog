/*
 * @Date: 2026-03-23 21:59:35
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-21 17:19:15
 * @Description:
 */
package config

import (
	"fmt"
	"go-blog/internal/utils"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// 全局配置变量
var Config = new(config)

type config struct {
	Application *application
	Mysql       *mysqlConfig
	Jwt         *jwt
	Logs        *logs
}

// 应用配置
type application struct {
	Port          int    `mapstructure:"port" yaml:"port"`
	UrlPathPrefix string `mapstructure:"url_path_prefix" yaml:"url_path_prefix"`
	// rsa公钥和私钥文件地址
	RSAPublicKeyPath  string `mapstructure:"rsa_public_key_path" yaml:"rsa_public_key_path"`
	RSAPrivateKeyPath string `mapstructure:"rsa_private_key_path" yaml:"rsa_private_key_path"`
	// rsa公钥和私钥文件内容
	RSAPublicBytes  []byte `mapstructure:"-" json:"-"`
	RSAPrivateBytes []byte `mapstructure:"-" json:"-"`
	// 非空时仅允许列表内 Origin 携带凭证跨域；为空则回显请求 Origin（勿在生产依赖此行为）
	CorsAllowOrigins []string `mapstructure:"cors_allow_origins" yaml:"cors_allow_origins"`
}

// mysql配置
type mysqlConfig struct {
	Username    string `mapstructure:"username" yaml:"username"`
	Password    string `mapstructure:"password" yaml:"password"`
	Host        string `mapstructure:"host" yaml:"host"`
	Port        int    `mapstructure:"port" yaml:"port"`
	Database    string `mapstructure:"database" yaml:"database"`
	Query       string `mapstructure:"query" yaml:"query"`
	LogMode     bool   `mapstructure:"log_mode" yaml:"log_mode"`
	TablePrefix string `mapstructure:"table_prefix" yaml:"table_prefix"`
	Charset     string `mapstructure:"charset" yaml:"charset"`
	Collation   string `mapstructure:"collation" yaml:"collation"`
}

// jwt配置
type jwt struct {
	Realm      string `mapstructure:"realm" yaml:"realm"`
	Key        string `mapstructure:"key" yaml:"key"`
	Timeout    int    `mapstructure:"timeout" yaml:"timeout"`
	MaxRefresh int    `mapstructure:"max_refresh" yaml:"max_refresh"`
}

// 日志配置
type logs struct {
	Level      int    `mapstructure:"level" yaml:"level"`
	Path       string `mapstructure:"path" yaml:"path"`
	MaxSize    int    `mapstructure:"max-size" yaml:"max-size"`
	MaxBackups int    `mapstructure:"max-backups" yaml:"max-backups"`
	MaxAge     int    `mapstructure:"max-age" yaml:"max-age"`
	Compress   bool   `mapstructure:"compress" yaml:"compress"`
}

/**
 * @description: 初始化配置
 * @return: err
 */
func InitConfig() error {
	// 1. 获取默认环境
	env := getEnv("APP_ENV", "dev")

	// 2. 初始化 viper
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./internal/config") // 配置目录

	// 3. 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("读取公共配置失败: %w", err)
	}

	// 读取环境配置文件
	envConfigName := fmt.Sprintf("config.%s", env)
	v.SetConfigName(envConfigName)
	if err := v.MergeInConfig(); err != nil {
		// 增量配置文件不是必需的，如果没有就继续
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("合并环境配置失败 %s: %w", env, err)
		}
		return fmt.Errorf("环境配置文件不存在: %s", envConfigName)
	}

	// 4. 合并环境覆盖配置
	if err := mergeEnvConfig(v, env); err != nil {
		return fmt.Errorf("合并环境配置失败: %w", err)
	}

	// 5. 支持环境变量覆盖（优先级高于配置文件）
	// 环境变量，数据库密码等可以放在环境变量中
	v.AutomaticEnv()
	// 自动将环境变量映射到配置键。例如 DATABASE_PASSWORD 会覆盖 database.password
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 6. 反序列化到结构体
	if err := v.Unmarshal(Config); err != nil {
		return fmt.Errorf("反序列化配置失败: %w", err)
	}

	// 7. 验证配置
	if err := ValidateConfig(Config); err != nil {
		return fmt.Errorf("验证配置失败: %w", err)
	}

	fmt.Printf("初始化配置成功: %#v\n", Config)
	return nil
}

// 获取环境变量
func getEnv(key string, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return strings.ToLower(val)
	}
	return defaultVal
}

// 合并环境配置文件
func mergeEnvConfig(v *viper.Viper, env string) error {
	file := filepath.Join("internal", "config", fmt.Sprintf("config.%s.yaml", env))

	v.SetConfigFile(file)
	return v.MergeInConfig()
}

/**
 * @description: 验证配置
 * @param: cfg *config
 * @return: error
 */
func ValidateConfig(cfg *config) error {
	// 读取rsa公钥和私钥文件内容
	Config.Application.RSAPublicBytes = utils.RSAReadKeyFromFile(Config.Application.RSAPublicKeyPath)
	Config.Application.RSAPrivateBytes = utils.RSAReadKeyFromFile(Config.Application.RSAPrivateKeyPath)
	// 更多校验...
	return nil
}
