/*
 * @Date: 2026-03-25 22:01:32
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-03-25 22:01:42
 * @Description: data transfer object for user
 */
package dto

import "go-blog/internal/model"

type UserInfoDto struct {
	ID           uint          `json:"id"`
	Username     string        `json:"username"`
	Mobile       string        `json:"mobile"`
	Avatar       string        `json:"avatar"`
	Nickname     string        `json:"nickname"`
	Introduction string        `json:"introduction"`
	Roles        []*model.Role `json:"roles"`
}
