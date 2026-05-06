package clickhouse

import (
	"fmt"
	"go-blog/internal/common"
	"go-blog/internal/config"
	"go-blog/internal/model"
	"go-blog/internal/utils"

	"gorm.io/driver/clickhouse"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var ClickHouseDB *gorm.DB

func InitClickHouse() error {
	dsn := fmt.Sprintf("clickhouse://%s:%s@%s:%d/%s?dial_timeout=5s&compress=lz4",
		config.Conf.ClickHouse.Username,
		config.Conf.ClickHouse.Password,
		config.Conf.ClickHouse.Host,
		config.Conf.ClickHouse.Port,
		config.Conf.ClickHouse.Database,
	)

	db, err := gorm.Open(clickhouse.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // 可按需调整日志级别
	})
	if err != nil {
		common.Log.Errorf("连接 ClickHouse 失败: %v", err)
		return err
	}

	ClickHouseDB = db

	// dev环境下自动迁移表结构
	env := utils.GetEnv("APP_ENV", "dev")
	if env == "dev" {
		if err := ClickHouseDB.AutoMigrate(&model.Blog{}); err != nil {
			common.Log.Errorf("自动迁移失败: %v", err)
			return err
		}
	}
	return nil
}
