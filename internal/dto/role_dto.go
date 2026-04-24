/*
 * @Date: 2026-04-02 22:32:12
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-02 22:37:11
 * @Description: Role DTO
 */
package dto

import (
	"go-blog/internal/model"
)

// 角色DTO
type RoleDto struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Keyword     string `json:"keyword"`
	Description string `json:"description"`
	Status      uint   `json:"status"`
	Sort        uint   `json:"sort"`
	Creator     string `json:"creator"`
	CreatedAt   string `json:"createdAt"` // 年月日 时分秒，如 2006-01-02 15:04:05
	UpdatedAt   string `json:"updatedAt"`
}

const roleDateTimeLayout = "2006-01-02 15:04:05"

/** 将角色列表转换为角色DTO列表
 * @param roleList []*model.Role 角色列表
 * @return []RoleDto
 */
func ToRolesDto(roleList []*model.Role) []RoleDto {
	var roles []RoleDto
	for _, role := range roleList {
		roleDto := RoleDto{
			ID:          role.ID,
			Name:        role.Name,
			Keyword:     role.Keyword,
			Description: *role.Desc,
			Status:      role.Status,
			Sort:        role.Sort,
			Creator:     role.Creator,
			CreatedAt:   role.CreatedAt.Format(roleDateTimeLayout),
			UpdatedAt:   role.UpdatedAt.Format(roleDateTimeLayout),
		}
		roles = append(roles, roleDto)
	}
	return roles
}
