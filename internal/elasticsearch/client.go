/*
 * @Date: 2026-04-29 14:32:05
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-05-05 21:48:25
 * @Description: ES客户端
 */
package elasticsearch

import (
	"bytes"
	"context"
	"fmt"
	"go-blog/internal/common"
	"go-blog/internal/config"

	"github.com/elastic/go-elasticsearch/v8"
)

// ES客户端全局变量
var ESClient *elasticsearch.Client

// 初始化ES客户端
func InitESClient() error {
	esCfg := elasticsearch.Config{
		Addresses: []string{config.Conf.ElasticSearch.Address},
		Username:  config.Conf.ElasticSearch.Username,
		Password:  config.Conf.ElasticSearch.Password,
	}
	client, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		common.Log.Errorf("elasticsearch 初始化失败: %v", err)
		return err
	}
	// 测试连接
	res, err := client.Info()
	if err != nil || res.IsError() {
		common.Log.Errorf("elasticsearch 连接失败: %v", err)
		return err
	}
	defer res.Body.Close()

	ESClient = client

	// 检查索引是否存在，不存在则创建（带映射）
	if err := ensureIndexExists(config.Conf.ElasticSearch.IndexName); err != nil {
		common.Log.Errorf("elasticsearch 索引检查/创建失败: %v", err)
		return err
	}

	common.Log.Infof("elasticsearch 连接成功，索引 %s 就绪", config.Conf.ElasticSearch.IndexName)
	return nil
}

/**
 * @description: 判断索引是否存在，不存在则创建
 * @param {string} indexName 索引名称
 * @return {error}
 */
func ensureIndexExists(indexName string) error {
	// 1. 检查索引是否存在
	existsRes, err := ESClient.Indices.Exists([]string{indexName})
	if err != nil {
		return err
	}
	defer existsRes.Body.Close()

	// 状态码 404 表示不存在，其他错误直接返回
	if existsRes.StatusCode == 404 {
		// 2. 创建索引并带上映射
		return createIndexWithMapping(indexName)
	} else if existsRes.StatusCode != 200 {
		return fmt.Errorf("检查索引存在性返回非预期状态码: %d", existsRes.StatusCode)
	}

	return nil
}

/**
 * @description: 创建索引并设置映射
 * @param {string} indexName 索引名称
 * @return {error}
 */
func createIndexWithMapping(indexName string) error {
	// 映射须为合法 JSON（不可用 # 注释）；dynamic strict 需与 IndexBlogToES 文档字段一致。
	mapping := `{
        "settings": {
            "number_of_shards": 1,
            "number_of_replicas": 1,
			"refresh_interval": "5s",
			"analysis": {
				"analyzer": {
					"blog_analyzer": {
						"type": "custom",
						"tokenizer": "ik_max_word",
						"filter": ["lowercase", "stop", "english_keywords"]
					}
				},
				"filter": {
					"english_keywords": {
						"type": "keyword_marker",
						"keywords": ["API", "HTTP", "SQL", "ID", "JSON", "UUID"]
					}
				}
			}
        },
        "mappings": {
            "properties": {
                "id":         { "type": "integer" },
                "title":      {
					"type": "text",
					"analyzer": "blog_analyzer",
					"term_vector": "with_positions_offsets"
				},
                "content":    {
					"type": "text",
					"analyzer": "blog_analyzer",
					"term_vector": "with_positions_offsets"
				},
				"summary":   {
					"type": "text",
					"analyzer": "blog_analyzer",
					"term_vector": "with_positions_offsets"
				},
				"cover_image": {
					"type": "text"
				},
				"tags":       { "type": "keyword" },
				"category":   { "type": "keyword" },
                "author":     { "type": "keyword" },
                "status":     { "type": "integer" },
				"published_at": {
					"type": "date",
					"format": "strict_date_optional_time||epoch_millis"
				},
                "created_at": {
					"type": "date",
					"format": "strict_date_optional_time||epoch_millis"
				},
                "updated_at": {
					"type": "date",
					"format": "strict_date_optional_time||epoch_millis"
				}
            }
        }
    }`

	// v8 使用 Indices.Create 并传入 Body
	createRes, err := ESClient.Indices.Create(
		indexName,
		ESClient.Indices.Create.WithBody(bytes.NewReader([]byte(mapping))),
		ESClient.Indices.Create.WithContext(context.Background()),
	)
	if err != nil {
		return fmt.Errorf("创建索引请求失败: %w", err)
	}
	defer createRes.Body.Close()

	if createRes.IsError() {
		return fmt.Errorf("创建索引返回错误: %s", createRes.String())
	}
	common.Log.Infof("elasticsearch 创建索引成功: %s", indexName)
	return nil
}
