/*
 * @Date: 2026-04-22 16:45:14
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-24 10:59:42
 * @Description: 环境变量工具
 */
package utils

import (
	"os"
	"strings"
)

/** 获取环境变量
 * @param key string 环境变量key
 * @param defaultVal string 默认值
 * @return string 环境变量值
 */
func GetEnv(key string, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return strings.ToLower(val)
	}
	return defaultVal
}
