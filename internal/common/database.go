/*
 * @Date: 2026-03-23 23:08:26
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-24 11:08:59
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
		config.Config.Mysql.Username,
		config.Config.Mysql.Password,
		config.Config.Mysql.Host,
		config.Config.Mysql.Port,
		config.Config.Mysql.Database,
		config.Config.Mysql.Charset,
		config.Config.Mysql.Collation,
		config.Config.Mysql.Query,
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		// NamingStrategy: schema.NamingStrategy{
		// 	TablePrefix: config.Config.Mysql.TablePrefix + "_",
		// },
	})

	if err != nil {
		panic(fmt.Errorf("初始化mysql数据库异常: %v/n", err))
	}

	// 开启mysql日志
	if config.Config.Mysql.LogMode {
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
	)
}

/** 执行数据库事务
 * @param fc func(gdb *gorm.DB) error 事务回调函数
 * @return error 错误
 */
func Transaction(fc func(gdb *gorm.DB) error) error {
	return DB.Transaction(fc)
}
