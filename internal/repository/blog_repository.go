/*
 * @Date: 2026-04-22 16:03:57
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-22 17:39:09
 * @Description:
 */
package repository

import (
	"go-blog/internal/common"
	"go-blog/internal/model"
	"go-blog/internal/vo"
)

type IBlogRepository interface {
	GetBlogs(req *vo.GetBlogListRequest) ([]model.Blog, int64, error) // 获取文章列表
	CreateBlog(blog *model.Blog) error                                // 创建文章
}

type BlogRepository struct {
}

func NewBlogRepository() IBlogRepository {
	return BlogRepository{}
}

// 获取文章列表
func (br BlogRepository) GetBlogs(req *vo.GetBlogListRequest) ([]model.Blog, int64, error) {
	// var blogs []model.Blog
	// err := common.DB.Find(&blogs).Error
	// return blogs, err
	var list []model.Blog
	db := common.DB.Model(&model.Blog{}).Order("created_at DESC")
	status := req.Status
	if status != 0 {
		db = db.Where("status = ?", status)
	}
	// 当pageNum > 0 且 pageSize > 0 才分页
	//记录总条数
	var total int64
	err := db.Count(&total).Error
	if err != nil {
		return list, total, err
	}
	page := int(req.Page)
	pageSize := int(req.PageSize)
	if page > 0 && pageSize > 0 {
		err = db.Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	} else {
		err = db.Find(&list).Error
	}
	return list, total, err
}

// 创建文章
func (br BlogRepository) CreateBlog(blog *model.Blog) error {
	err := common.DB.Create(blog).Error
	return err
}
