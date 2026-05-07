/*
 * @Date: 2026-03-23 23:08:26
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-05-07 15:15:50
 * @Description: database
 */
package common

import (
	"fmt"
	"go-blog/internal/config"
	"go-blog/internal/model"

	"go-blog/internal/utils"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

/**
 * @description: 初始化mysql数据库
 * @return: void
 */
func InitMysql() {
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
		panic(fmt.Errorf("初始化mysql数据库异常: %v\n", err))
	}

	// 开启mysql日志
	if config.Conf.Mysql.LogMode {
		db.Debug()
	}

	DB = db

	// 迁移表结构
	if env := utils.GetEnv("APP_ENV", "dev"); env == "dev" {
		dbAutoMigrate()
	}
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
