/*
 * @Date: 2026-03-23 21:59:35
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-05-08 11:11:01
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
var Conf = new(config)

type config struct {
	Application   *application         `mapstructure:"application" yaml:"application"`
	Mysql         *mysqlConfig         `mapstructure:"mysql" yaml:"mysql"`
	Jwt           *jwt                 `mapstructure:"jwt" yaml:"jwt"`
	Logs          *logs                `mapstructure:"logs" yaml:"logs"`
	RateLimit     *RateLimitConfig     `mapstructure:"rate-limit" yaml:"rate-limit"`
	Rabbitmq      *RabbitmqConfig      `mapstructure:"rabbitmq" yaml:"rabbitmq"`
	ElasticSearch *ElasticSearchConfig `mapstructure:"elastic-search" yaml:"elastic-search"`
	ClickHouse    *ClickHouseConfig    `mapstructure:"click-house" yaml:"click-house"`
}

// 应用配置
type application struct {
	// 端口
	Port int `mapstructure:"port" yaml:"port"`
	// 接口路径前缀
	UrlPathPrefix string `mapstructure:"url_path_prefix" yaml:"url_path_prefix"`
	// rsa公钥文件地址
	RSAPublicKeyPath string `mapstructure:"rsa_public_key_path" yaml:"rsa_public_key_path"`
	// rsa私钥文件地址
	RSAPrivateKeyPath string `mapstructure:"rsa_private_key_path" yaml:"rsa_private_key_path"`
	// rsa公钥和私钥文件内容
	RSAPublicBytes []byte `mapstructure:"-" json:"-"`
	// rsa私钥文件内容
	RSAPrivateBytes []byte `mapstructure:"-" json:"-"`
	// 非空时仅允许列表内 Origin 携带凭证跨域；为空则回显请求 Origin（勿在生产依赖此行为）
	CorsAllowOrigins []string `mapstructure:"cors_allow_origins" yaml:"cors_allow_origins"`
}

// mysql配置
type mysqlConfig struct {
	// 用户名
	Username string `mapstructure:"username" yaml:"username"`
	// 密码
	Password string `mapstructure:"password" yaml:"password"`
	// 主机地址
	Host string `mapstructure:"host" yaml:"host"`
	// 端口
	Port int `mapstructure:"port" yaml:"port"`
	// 数据库名
	Database string `mapstructure:"database" yaml:"database"`
	// 连接字符串参数
	Query string `mapstructure:"query" yaml:"query"`
	// 是否打印日志
	LogMode bool `mapstructure:"log_mode" yaml:"log_mode"`
	// 数据库表前缀
	TablePrefix string `mapstructure:"table_prefix" yaml:"table_prefix"`
	// 编码方式
	Charset string `mapstructure:"charset" yaml:"charset"`
	// 字符集
	Collation string `mapstructure:"collation" yaml:"collation"`
}

// jwt配置
type jwt struct {
	// jwt标识
	Realm string `mapstructure:"realm" yaml:"realm"`
	// 服务端密钥
	Key string `mapstructure:"key" yaml:"key"`
	// token过期时间
	Timeout int `mapstructure:"timeout" yaml:"timeout"`
	// token最大刷新时间
	MaxRefresh int `mapstructure:"max_refresh" yaml:"max_refresh"`
}

// 日志配置
type logs struct {
	// 日志等级
	Level int `mapstructure:"level" yaml:"level"`
	// 日志路径
	Path string `mapstructure:"path" yaml:"path"`
	// 文件最大大小
	MaxSize int `mapstructure:"max-size" yaml:"max-size"`
	// 备份数
	MaxBackups int `mapstructure:"max-backups" yaml:"max-backups"`
	// 存放时间
	MaxAge int `mapstructure:"max-age" yaml:"max-age"`
	// 是否压缩
	Compress bool `mapstructure:"compress" yaml:"compress"`
}

// 限流配置
type RateLimitConfig struct {
	// 填充一个令牌需要的时间间隔
	FillInterval int64 `mapstructure:"fill-interval" json:"fillInterval"`
	// 桶容量
	Capacity int64 `mapstructure:"capacity" json:"capacity"`
}

// rabbitmq配置
type RabbitmqConfig struct {
	// 主机地址
	Host string `mapstructure:"host" yaml:"host"`
	// 端口
	Port int `mapstructure:"port" yaml:"port"`
	// 用户名
	Username string `mapstructure:"username" yaml:"username"`
	// 密码
	Password string `mapstructure:"password" yaml:"password"`
	// 虚拟主机 类似 MySQL 的不同 Database
	VHost string `mapstructure:"vhost" yaml:"vhost"`
	// 队列相关
	QueueName string `mapstructure:"queue-name" yaml:"queue-name"`
	// 交换机名称
	ExchangeName string `mapstructure:"exchange-name" yaml:"exchange-name"`
	// 路由键
	RoutingKey string `mapstructure:"routing-key" yaml:"routing-key"`
	// 交换机类型
	ExchangeType string `mapstructure:"exchange-type" yaml:"exchange-type"`
	// 队列持久化 生产环境必须设为 true
	// 但要注意，仅开启队列持久化是不够的，发送消息时还需要将消息的 DeliveryMode 设置为 2 (持久化)，才能保证消息写入磁盘
	Durable bool `mapstructure:"durable" yaml:"durable"`
	// 当最后一个消费者断开连接后，队列或交换机是否自动删除 核心业务队列应设为 false
	AutoDelete bool `mapstructure:"auto-delete" yaml:"auto-delete"`
	// 设置为 true 时，该队列仅对首次声明它的连接（Connection）可见，通常设为 false
	Exclusive bool `mapstructure:"exclusive" yaml:"exclusive"`
	// 限制每次推送的消息数，防止消费者过载
	PrefetchCount int `mapstructure:"prefetch-count" yaml:"prefetch-count"`
	// 自动确认消息 生产环境强烈建议设为 false,手动去确认消息成功
	AutoAck bool `mapstructure:"auto-ack" yaml:"auto-ack"`
}

// elasticSearch配置
type ElasticSearchConfig struct {
	// 地址
	Address string `mapstructure:"address" yaml:"address"`
	// 索引名称
	IndexName string `mapstructure:"index-name" yaml:"index-name"`
	// 用户名
	Username string `mapstructure:"username" yaml:"username"`
	// 密码
	Password string `mapstructure:"password" yaml:"password"`
}

// ClickHouse配置
type ClickHouseConfig struct {
	// 主机地址
	Host string `mapstructure:"Host" yaml:"Host"`
	// 端口
	Port int `mapstructure:"Port" yaml:"Port"`
	// 数据库
	Database string `mapstructure:"database" yaml:"database"`
	// 表名
	TableName string `mapstructure:"table-name" yaml:"table-name"`
	// 用户名
	Username string `mapstructure:"username" yaml:"username"`
	// 密码
	Password string `mapstructure:"password" yaml:"password"`
}

/**
 * @description: 初始化配置
 * @return: err
 */
func InitConfig() error {
	// 1. 获取默认环境
	env := utils.GetEnv("APP_ENV", "dev")

	// 2. 初始化 viper
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./internal/config") // 配置目录

	// 3. 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("读取公共配置失败: %w", err)
	}

	// 4. 可选合并环境覆盖配置：存在 config.{env}.yaml 则合并，不存在则仅使用公共配置
	if err := mergeEnvConfig(v, env); err != nil {
		return fmt.Errorf("合并环境配置失败: %w", err)
	}

	// 5. 支持环境变量覆盖（优先级高于配置文件）
	// 环境变量，数据库密码等可以放在环境变量中
	v.AutomaticEnv()
	// 自动将环境变量映射到配置键。例如 DATABASE_PASSWORD 会覆盖 database.password
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 6. 反序列化到结构体
	if err := v.Unmarshal(Conf); err != nil {
		return fmt.Errorf("反序列化配置失败: %w", err)
	}

	// 7. 验证配置
	if err := ValidateConfig(Conf); err != nil {
		return fmt.Errorf("验证配置失败: %w", err)
	}

	return nil
}

// mergeEnvConfig 合并 internal/config/config.{env}.yaml；文件不存在时不报错，与「环境增量配置可选」一致。
func mergeEnvConfig(v *viper.Viper, env string) error {
	file := filepath.Join("internal", "config", fmt.Sprintf("config.%s.yaml", env))
	st, err := os.Stat(file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if st.IsDir() {
		return nil
	}
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
	Conf.Application.RSAPublicBytes = utils.RSAReadKeyFromFile(Conf.Application.RSAPublicKeyPath)
	Conf.Application.RSAPrivateBytes = utils.RSAReadKeyFromFile(Conf.Application.RSAPrivateKeyPath)
	// 更多校验...
	return nil
}
