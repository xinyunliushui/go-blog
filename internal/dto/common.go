/*
 * @Date: 2026-04-13 13:34:18
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-24 10:51:34
 * @Description: 公共DTO
 */
package dto

/** 将指针字符串转换为字符串
 * @param p *string 指针字符串
 * @return string
 */
func PtrStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

/** 将指针uint转换为uint
 * @param p *uint 指针uint
 * @return uint
 */
func PtrUint(p *uint) uint {
	if p == nil {
		return 0
	}
	return *p
}
