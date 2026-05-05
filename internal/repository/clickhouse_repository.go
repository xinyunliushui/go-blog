/*
 * @Date: 2026-05-04 16:35:42
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-05-04 20:39:59
 * @Description: 将博客插入到ClickHouse
 */
package repository

import (
	"fmt"
	"go-blog/internal/clickhouse"
	"go-blog/internal/model"
)

/**
 * @description: 插入博客到ClickHouse
 * @param {model.Blog} blog
 * @return {error}
 */
func InsertBlogToClickHouse(blog *model.Blog) error {
	if clickhouse.ClickHouseDB == nil {
		return fmt.Errorf("ClickHouse 未初始化，请先调用 InitClickHouse")
	}
	// 直接使用 GORM 的 Create 方法插入一条记录
	result := clickhouse.ClickHouseDB.Create(blog)
	if result.Error != nil {
		return fmt.Errorf("插入 ClickHouse 失败: %w", result.Error)
	}
	return nil
}
