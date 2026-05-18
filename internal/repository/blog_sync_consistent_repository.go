/*
 * @Date: 2026-05-18 14:32:42
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-05-18 17:32:34
 * @Description: ES/CH 存在性探测与幂等写入，用于补偿对账
 */

package repository

import (
	"context"
	"fmt"
	"go-blog/internal/clickhouse"
	"go-blog/internal/config"
	"go-blog/internal/elasticsearch"
	"go-blog/internal/model"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

/**
 * @description: 判断文档是否已索引。
 * @param blogID string blogID
 * @return bool 是否存在
 * @return error 错误
 */
func BlogExistsInES(blogID string) (bool, error) {
	if elasticsearch.ESClient == nil {
		return false, fmt.Errorf("Elasticsearch 未初始化")
	}
	req := esapi.ExistsRequest{
		Index:      config.Conf.ElasticSearch.IndexName,
		DocumentID: blogID,
	}
	res, err := req.Do(context.Background(), elasticsearch.ESClient)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()
	return res.StatusCode == 200, nil
}

/**
 * @description: 判断分析库是否已有该博客记录。
 * @param blogID string blogID
 * @return bool 是否存在
 * @return error 错误
 */
func BlogExistsInClickHouse(blogID string) (bool, error) {
	if clickhouse.ClickHouseDB == nil {
		return false, fmt.Errorf("ClickHouse 未初始化")
	}
	var count int64
	err := clickhouse.ClickHouseDB.Model(&model.Blog{}).Where("id = ?", blogID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

/**
 * @description: 按 blog_id 幂等写入（有则更新，无则插入）。
 * @param blog *model.Blog 博客
 * @return error 错误
 */
func UpsertBlogToClickHouse(blog *model.Blog) error {
	if clickhouse.ClickHouseDB == nil {
		return fmt.Errorf("ClickHouse 未初始化")
	}
	exists, err := BlogExistsInClickHouse(blog.ID)
	if err != nil {
		return err
	}
	if !exists {
		return InsertBlogToClickHouse(blog)
	}
	result := clickhouse.ClickHouseDB.Model(&model.Blog{}).Where("id = ?", blog.ID).Updates(blog)
	if result.Error != nil {
		return fmt.Errorf("更新 ClickHouse 失败: %w", result.Error)
	}
	return nil
}

/**
 * @description: 对照 ES/CH 实际状态，计算仍待同步的位图。
 * @param blogID string blogID
 * @param currentMask uint8 当前位图
 * @return uint8 仍待同步位图（0=均已存在；ES=1、CH=2、3=两侧都缺）
 * @return error 错误
 */
func ReconcileSyncPendingMask(blogID string, currentMask uint8) (uint8, error) {
	if blogID == "" || blogID == "unknown" {
		return currentMask, nil
	}

	var pending uint8
	// 检查 ES 是否存在
	esOK, esErr := BlogExistsInES(blogID)
	if esErr != nil {
		pending |= model.SyncPendingES
	} else if !esOK {
		pending |= model.SyncPendingES
	}

	// 检查 CH 是否存在
	chOK, chErr := BlogExistsInClickHouse(blogID)
	if chErr != nil {
		pending |= model.SyncPendingCH
	} else if !chOK {
		pending |= model.SyncPendingCH
	}
	// pending 为 0：探测成功且 ES、CH 均已存在，无待同步项
	if pending == 0 {
		return 0, nil
	}
	// 与任务记录合并，避免对账结果漏掉任一侧
	if currentMask != 0 {
		pending = model.MergeSyncMask(pending, currentMask)
	}
	return pending, nil
}
