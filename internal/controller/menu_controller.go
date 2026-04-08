package controller

import (
	"go-blog/internal/repository"
	"go-blog/internal/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type IMenuController interface {
	GetUserMenuTreeByUserId(c *gin.Context) // 获取用户的可访问菜单树
}

type MenuController struct {
	MenuRepository repository.IMenuRepository
}

func NewMenuController() IMenuController {
	return MenuController{
		MenuRepository: repository.NewMenuRepository(),
	}
}

// 根据用户ID获取用户的可访问菜单树
func (mc MenuController) GetUserMenuTreeByUserId(c *gin.Context) {
	// 获取路径中的userId
	userId, _ := strconv.Atoi(c.Param("userId"))
	if userId <= 0 {
		response.Fail(c, nil, "用户ID不正确")
		return
	}

	menuTree, err := mc.MenuRepository.GetUserMenuTreeByUserId(uint(userId))
	if err != nil {
		response.Fail(c, nil, "获取用户的可访问菜单树失败: "+err.Error())
		return
	}
	response.Success(c, menuTree, "获取用户的可访问菜单树成功")
}
