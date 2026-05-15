/*
 * @Date: 2026-04-13 13:34:18
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-05-15
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

/** 将父菜单指针转为前端展示用字符串（nil 表示根） */
func PtrParentIDString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
