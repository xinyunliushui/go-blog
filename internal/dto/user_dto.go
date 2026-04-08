/*
 * @Date: 2026-03-25 22:01:32
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-03 14:59:22
 * @Description: user dto
 */
package dto

import "go-blog/internal/model"

// 返回给前端的当前用户信息
type UserInfoDto struct {
	ID           uint   `json:"id"`
	Username     string `json:"username"`
	Mobile       string `json:"mobile"`
	Avatar       string `json:"avatar"`
	Nickname     string `json:"nickname"`
	Introduction string `json:"introduction"`
	RoleIds      []uint `json:"roleIds"`
}

func ToUserInfoDto(user model.User) UserInfoDto {
	// 角色色处理
	roleIds := make([]uint, 0)
	for _, role := range user.Roles {
		roleIds = append(roleIds, role.ID)
	}
	return UserInfoDto{
		ID:           user.ID,
		Username:     user.Username,
		Mobile:       user.Mobile,
		Avatar:       user.Avatar,
		Nickname:     user.Nickname,
		Introduction: user.Introduction,
		RoleIds:      roleIds,
	}
}

// 返回给前端的用户列表
type UsersDto struct {
	ID           uint   `json:"id"`
	Username     string `json:"username"`
	Mobile       string `json:"mobile"`
	Avatar       string `json:"avatar"`
	Nickname     string `json:"nickname"`
	Introduction string `json:"introduction"`
	Status       uint   `json:"status"`
	Creator      string `json:"creator"`
	RoleIds      []uint `json:"roleIds"`
}

func ToUsersDto(userList []model.User) []UsersDto {
	var users []UsersDto
	for _, user := range userList {
		userDto := UsersDto{
			ID:           user.ID,
			Username:     user.Username,
			Mobile:       user.Mobile,
			Avatar:       user.Avatar,
			Nickname:     user.Nickname,
			Introduction: user.Introduction,
			Status:       user.Status,
			Creator:      user.Creator,
		}
		roleIds := make([]uint, 0)
		for _, role := range user.Roles {
			roleIds = append(roleIds, role.ID)
		}
		userDto.RoleIds = roleIds
		users = append(users, userDto)
	}
	return users
}
