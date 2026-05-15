/*
 * @Date: 2026-03-23 23:08:26
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-05-15 11:32:37
 * @Description: database
 */
package common

import (
	"fmt"
	"go-blog/internal/config"
	"go-blog/internal/model"
	"time"

	"go-blog/internal/utils"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

/**
 * @description: 初始化mysql数据库
 * @return error 失败时 DB 保持 nil，由调用方记录日志；就绪探针可反映未就绪
 */
func InitMysql() error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&collation=%s&%s",
		config.Conf.Mysql.Username,
		config.Conf.Mysql.Password,
		config.Conf.Mysql.Host,
		config.Conf.Mysql.Port,
		config.Conf.Mysql.Database,
		config.Conf.Mysql.Charset,
		config.Conf.Mysql.Collation,
		config.Conf.Mysql.Query,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		// NamingStrategy: schema.NamingStrategy{
		// 	TablePrefix: config.Conf.Mysql.TablePrefix + "_",
		// },
	})
	if err != nil {
		Log.Errorf("初始化 mysql 数据库失败: %v", err)
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		Log.Errorf("获取底层 sql.DB 失败: %v", err)
		return err
	}

	// 设置了数据库连接的最大存活时间为 30 分钟
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	// 设置了空闲连接的最大存活时间为 10 分钟
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)
	// 设置了最大打开连接数为 50
	sqlDB.SetMaxOpenConns(50)
	// 设置了最大空闲连接数为 10
	sqlDB.SetMaxIdleConns(10)

	// 开启mysql日志
	if config.Conf.Mysql.LogMode {
		db.Debug()
	}

	DB = db

	// 迁移表结构
	if env := utils.GetEnv("APP_ENV", "dev"); env == "dev" {
		dbAutoMigrate()
	}
	return nil
}

/**
 * @description: 自动迁移表结构
 * @return: void
 */
func dbAutoMigrate() {
	DB.AutoMigrate(
		&model.User{},
		&model.Role{},
		&model.Menu{},
		&model.Blog{},
		&model.BlogMQOutbox{},
	)
}

/** 执行数据库事务
 * @param fc func(gdb *gorm.DB) error 事务回调函数
 * @return error 错误
 */
func Transaction(fc func(gdb *gorm.DB) error) error {
	return DB.Transaction(fc)
}
