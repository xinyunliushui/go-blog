/*
 * @Date: 2026-03-25 21:56:11
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-03-31 23:02:07
 * @Description:
 */
package common

import (
	"errors"
	"fmt"
	"go-blog/internal/model"

	"gorm.io/gorm"
)

func InitMysqlData() {
	newUsers := make([]model.User, 0)
	users := []model.User{
		{
			Model:        gorm.Model{ID: 1},
			Username:     "theShy",
			Password:     "654321",
			Mobile:       "18888888888",
			Avatar:       "https://wpimg.wallstcn.com/f778738c-e4f8-4870-b634-56703b4acafe.gif",
			Nickname:     "theShy",
			Introduction: "lpl",
			Status:       1,
			Creator:      "系统",
		},
		{
			Model:        gorm.Model{ID: 2},
			Username:     "faker",
			Password:     "123456",
			Mobile:       "18888888899",
			Avatar:       "https://wpimg.wallstcn.com/f778738c-e4f8-4870-b634-56703b4acafe.gif",
			Nickname:     "faker",
			Introduction: "lck",
			Status:       1,
			Creator:      "系统",
		},
	}

	for _, user := range users {
		err := DB.First(&user, user.ID).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newUsers = append(newUsers, user)
		}
	}

	if len(newUsers) > 0 {
		err := DB.Create(&newUsers).Error
		if err != nil {
			fmt.Println("error: 写入用户数据失败：%v\n", err)
		}
	}
}
