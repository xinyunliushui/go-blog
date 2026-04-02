package controller

import (
	"go-blog/internal/common"
	"go-blog/internal/dto"
	"go-blog/internal/repository"
	"go-blog/internal/response"
	"go-blog/internal/vo"

	"github.com/gin-gonic/gin"
)

type IRoleController interface {
	GetRoles(ctx *gin.Context)
}

type RoleController struct {
	RoleRepository repository.IRoleRepository
}

func NewRoleController() IRoleController {
	return RoleController{
		RoleRepository: repository.NewRoleRepository(),
	}
}

func (rc RoleController) GetRoles(ctx *gin.Context) {
	var req vo.RoleListRequest
	// 参数绑定
	if err := ctx.ShouldBind(&req); err != nil {
		response.Fail(ctx, nil, common.ValidationErrString(err))
		return
	}
	// 参数校验
	if err := common.Validate.Struct(&req); err != nil {
		response.Fail(ctx, nil, common.ValidationErrString(err))
		return
	}

	// 获取角色列表
	roles, total, err := rc.RoleRepository.GetRoles(&req)
	if err != nil {
		response.Fail(ctx, nil, "获取角色列表失败："+err.Error())
		return
	}
	response.Success(ctx, gin.H{"content": dto.ToRolesDto(roles), "total": total, "page": req.Page, "pageSize": req.PageSize}, "获取角色列表成功")
}
