package utils

import (
	"os"
	"strings"
)

// 获取环境变量
func GetEnv(key string, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return strings.ToLower(val)
	}
	return defaultVal
}
