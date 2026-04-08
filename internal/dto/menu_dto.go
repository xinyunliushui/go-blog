package dto

import "go-blog/internal/model"

type MenuDto struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Title     string `json:"title"`
	Icon      string `json:"icon"`
	Path      string `json:"path"`
	Redirect  string `json:"redirect"`
	Component string `json:"component"`
	Sort      uint   `json:"sort"`
	Status    uint   `json:"status"`
}

func ToMenuDto(menu model.Menu) MenuDto {
	return MenuDto{
		ID:    menu.ID,
		Name:  menu.Name,
		Title: menu.Title,
		// Icon:      menu.Icon,
		Path: menu.Path,
		// Redirect:  menu.Redirect,
		Component: menu.Component,
		Sort:      menu.Sort,
		Status:    menu.Status,
	}
}
