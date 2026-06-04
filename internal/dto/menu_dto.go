package dto

import "go-blog/internal/model"

// 菜单DTO
type MenuDto struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Title    string    `json:"title"`
	Icon     string    `json:"icon"`
	Path     string    `json:"path"`
	Redirect string    `json:"redirect"`
	Sort     uint      `json:"sort"`
	Status   uint      `json:"status"`
	Type     uint      `json:"type"`
	ParentId string    `json:"parentId"`
	Creator  string    `json:"creator"`
	Version  uint      `json:"version"`
	Children []MenuDto `json:"children"`
}

/** 将菜单树转换为菜单DTO树
 * @param menus []*model.Menu 菜单列表
 * @return []MenuDto
 */
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
			ParentId: PtrParentIDString(menu.ParentId),
			Creator:  menu.Creator,
			Version:  menu.Version,
		}
		if len(menu.Children) > 0 {
			newMenu.Children = ToMenuTreeDto(menu.Children)
		}
		tree = append(tree, newMenu)
	}
	return tree
}

/** 将菜单转换为菜单DTO
 * @param menu *model.Menu 菜单
 * @return MenuDto
 */
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
		ParentId: PtrParentIDString(menu.ParentId),
		Creator:  menu.Creator,
		Version:  menu.Version,
	}
}

/** 将菜单列表转换为菜单DTO列表
 * @param menus []*model.Menu 菜单列表
 * @return []MenuDto
 */
func ToMenuListDto(menus []*model.Menu) []MenuDto {
	list := make([]MenuDto, 0)
	for _, menu := range menus {
		newMenu := ToMenuDto(menu)
		list = append(list, newMenu)
	}
	return list
}
