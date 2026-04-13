package dto

import "go-blog/internal/model"

type MenuDto struct {
	ID       uint      `json:"id"`
	Name     string    `json:"name"`
	Title    string    `json:"title"`
	Icon     string    `json:"icon"`
	Path     string    `json:"path"`
	Redirect string    `json:"redirect"`
	Sort     uint      `json:"sort"`
	Status   uint      `json:"status"`
	Type     uint      `json:"type"`
	ParentId uint      `json:"parentId"`
	Creator  string    `json:"creator"`
	Children []MenuDto `json:"children"`
}

// ToMenuTreeDto 将 model tree菜单树转为前端 DTO；递归处理 Children。
func ToMenuTreeDto(menus []*model.Menu) []MenuDto {
	if len(menus) == 0 {
		return nil
	}
	tree := make([]MenuDto, 0, len(menus))
	for _, menu := range menus {
		if menu == nil {
			continue
		}
		newMenu := MenuDto{
			ID:       menu.ID,
			Name:     menu.Name,
			Title:    menu.Title,
			Icon:     PtrStr(menu.Icon),
			Path:     menu.Path,
			Redirect: PtrStr(menu.Redirect),
			Sort:     menu.Sort,
			Status:   menu.Status,
			Type:     menu.Type,
			ParentId: PtrUint(menu.ParentId),
			Creator:  menu.Creator,
		}
		if len(menu.Children) > 0 {
			newMenu.Children = ToMenuTreeDto(menu.Children)
		}
		tree = append(tree, newMenu)
	}
	return tree
}

// ToMenuDto 将 model 菜单转为前端 DTO。
func ToMenuDto(menu *model.Menu) MenuDto {
	return MenuDto{
		ID:       menu.ID,
		Name:     menu.Name,
		Title:    menu.Title,
		Icon:     PtrStr(menu.Icon),
		Path:     menu.Path,
		Redirect: PtrStr(menu.Redirect),
		Sort:     menu.Sort,
		Status:   menu.Status,
		Type:     menu.Type,
		ParentId: PtrUint(menu.ParentId),
		Creator:  menu.Creator,
	}
}

func ToMenuListDto(menus []*model.Menu) []MenuDto {
	list := make([]MenuDto, 0)
	for _, menu := range menus {
		newMenu := ToMenuDto(menu)
		list = append(list, newMenu)
	}
	return list
}
