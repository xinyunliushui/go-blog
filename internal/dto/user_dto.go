/*
 * @Date: 2026-03-25 22:01:32
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-05-15
 * @Description: user dto
 */
package dto

import "go-blog/internal/model"

// 当前用户信息DTO
type UserInfoDto struct {
	ID           string   `json:"id"`
	Username     string   `json:"username"`
	Mobile       string   `json:"mobile"`
	Avatar       string   `json:"avatar"`
	Nickname     string   `json:"nickname"`
	Introduction string   `json:"introduction"`
	RoleIds      []string `json:"roleIds"`
}

/** 将用户转换为当前用户信息DTO
 * @param user model.User 用户
 * @return UserInfoDto
 */
func ToUserInfoDto(user model.User) UserInfoDto {
	// 角色色处理
	roleIds := make([]string, 0)
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
	ID           string   `json:"id"`
	Username     string   `json:"username"`
	Mobile       string   `json:"mobile"`
	Avatar       string   `json:"avatar"`
	Nickname     string   `json:"nickname"`
	Introduction string   `json:"introduction"`
	Status       uint     `json:"status"`
	Creator      string   `json:"creator"`
	RoleIds      []string `json:"roleIds"`
}

/** 将用户列表转换为用户DTO列表
 * @param userList []model.User 用户列表
 * @return []UsersDto
 */
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
		roleIds := make([]string, 0)
		for _, role := range user.Roles {
			roleIds = append(roleIds, role.ID)
		}
		userDto.RoleIds = roleIds
		users = append(users, userDto)
	}
	return users
}
