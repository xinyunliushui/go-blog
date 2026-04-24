/*
 * @Date: 2026-03-25 11:23:47
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-13 14:26:37
 * @Description: menu model
 */

package model

import (
	"gorm.io/gorm"
)

// 菜单模型结构体
type Menu struct {
	gorm.Model
	Name     string  `gorm:"type:varchar(50);comment:'名称(英文名)'" json:"name"`
	Title    string  `gorm:"type:varchar(50);comment:'标题(中文名)'" json:"title"`
	Icon     *string `gorm:"type:varchar(50);comment:'图标'" json:"icon"`
	Path     string  `gorm:"type:varchar(100);comment:'菜单访问路径'" json:"path"`
	Redirect *string `gorm:"type:varchar(100);comment:'重定向路径'" json:"redirect"`
	Sort     uint    `gorm:"type:int(3) unsigned;default:999;comment:'顺序(1-999)'" json:"sort"`
	Status   uint    `gorm:"type:tinyint(1);default:1;comment:'状态(正常/禁用, 默认正常)'" json:"status"`
	Type     uint    `gorm:"type:tinyint(1);default:1;comment:'类型(1是菜单, 2是页面, 3是按钮, 默认菜单)'" json:"hidden"`
	ParentId *uint   `gorm:"default:0;comment:'父节点编号(编号为0时表示根节点)'" json:"parentId"`
	Creator  string  `gorm:"type:varchar(20);comment:'创建人'" json:"creator"`
	Children []*Menu `gorm:"-" json:"children"`                  // 子菜单集合
	Roles    []*Role `gorm:"many2many:role_menus;" json:"roles"` // 角色菜单多对多关系
}
