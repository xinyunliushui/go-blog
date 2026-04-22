/*
 * @Date: 2026-03-23 23:08:26
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-22 17:05:46
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

//	Transaction 执行数据库事务的辅助函数。
//
// 当业务需要保证多个数据库操作的原子性时（要么全部成功，要么全部回滚），应使用本函数包裹相关操作。
// 回调函数中返回非 nil 错误时，GORM 将自动执行回滚；返回 nil 时自动提交。
//
// 使用注意：
//   - 回调内必须使用参数 gdb 执行数据库操作，不能使用 common.DB
//   - 勿在回调内发起 HTTP 请求或执行耗时外部调用，以免长时间持有连接
//
// 示例：
//
//	err := common.Transaction(func(gdb *gorm.DB) error {
//		if err := gdb.Create(&user).Error; err != nil {
//			return err
//		}
//		if err := gdb.Model(&user).Association("Roles").Replace(user.Roles); err != nil {
//			return err
//		}
//		return nil
//	})
func Transaction(fc func(gdb *gorm.DB) error) error {
	return DB.Transaction(fc)
}
