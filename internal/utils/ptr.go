/*
 * @Description: 指针辅助，用于 VO → Model 可空字段映射
 */
package utils

import "strings"

// OptionalString 将请求中的可选字符串转为 *string：空白视为未设置（nil），供 GORM 可空列与 JSON null 使用。
func OptionalString(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}
