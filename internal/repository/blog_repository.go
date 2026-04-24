/*
 * @Date: 2026-04-22 16:03:57
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-24 10:45:25
 * @Description: 文章接口实现
 */
package repository

import (
	"go-blog/internal/common"
	"go-blog/internal/model"
	"go-blog/internal/vo"
	"time"
)

type IBlogRepository interface {
	GetBlogs(req *vo.GetBlogListRequest) ([]model.Blog, int64, error)                   // 获取文章列表
	GetBlogById(blogId uint) (model.Blog, error)                                        // 根据ID获取文章详情
	CreateBlog(blog *model.Blog) error                                                  // 创建文章
	UpdateBlogPublishStatusById(blogId uint, status uint, publishedAt *time.Time) error // 更新文章状态
	UpdateBlogById(blogId uint, blog *model.Blog) error                                 // 更新文章
}

type BlogRepository struct {
}

func NewBlogRepository() IBlogRepository {
	return BlogRepository{}
}

/** 获取文章列表
 * @param req *vo.GetBlogListRequest 文章列表请求
 * @return []model.Blog, int64, error
 */
func (br BlogRepository) GetBlogs(req *vo.GetBlogListRequest) ([]model.Blog, int64, error) {
	var blogList []model.Blog
	db := common.DB.Model(&model.Blog{}).Order("created_at DESC")
	status := req.Status
	if status != 0 {
		db = db.Where("status = ?", status)
	}
	//记录总条数
	var total int64
	err := db.Count(&total).Error
	if err != nil {
		return blogList, total, err
	}
	page := int(req.Page)
	pageSize := int(req.PageSize)
	if page > 0 && pageSize > 0 {
		err = db.Offset((page - 1) * pageSize).Limit(pageSize).Find(&blogList).Error
	} else {
		err = db.Find(&blogList).Error
	}
	return blogList, total, err
}

/** 创建文章
 * @param blog 文章信息
 * @return error
 */
func (br BlogRepository) CreateBlog(blog *model.Blog) error {
	err := common.DB.Create(blog).Error
	return err
}

/** 根据ID获取文章详情
 * @param blogId 文章ID
 * @return model.Blog, error
 */
func (br BlogRepository) GetBlogById(blogId uint) (model.Blog, error) {
	var blog model.Blog
	err := common.DB.Where("id = ?", blogId).First(&blog).Error
	return blog, err
}

/** 更新文章状态和发布时间
 * @param blogId 文章ID
 * @param status 文章状态
 * @param publishedAt 发布时间
 * @return error
 */
func (br BlogRepository) UpdateBlogPublishStatusById(blogId uint, status uint, publishedAt *time.Time) error {
	var publishedAtValue interface{}
	if publishedAt == nil {
		publishedAtValue = nil
	} else {
		publishedAtValue = *publishedAt
	}
	blog := map[string]interface{}{"status": status, "published_at": publishedAtValue}
	err := common.DB.Model(&model.Blog{}).Where("id = ?", blogId).Updates(blog).Error
	return err
}

/** 更新文章
 * @param blogId 文章ID
 * @param blog 文章信息
 * @return error
 */
func (br BlogRepository) UpdateBlogById(blogId uint, blog *model.Blog) error {
	err := common.DB.Model(&model.Blog{}).Where("id = ?", blogId).Updates(blog).Error
	return err
}
