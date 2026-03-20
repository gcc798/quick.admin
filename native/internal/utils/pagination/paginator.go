package pagination

import (
	"gorm.io/gorm"
)

// Page 分页结果（参考 MyBatis Plus Page 结构）
type Page[T any] struct {
	Records []T   `json:"records"` // 查询数据列表
	Total   int64 `json:"total"`   // 总记录数
	Size    int64 `json:"size"`    // 每页显示条数
	Current int64 `json:"current"` // 当前页
	Pages   int64 `json:"pages"`   // 总页数
}

// Paginator 分页器
// 使用泛型支持任意模型类型，提供统一的分页处理
type Paginator[T any] struct {
	db       *gorm.DB
	query    *PageQuery
	page     *Page[T]
	err      error
	preloads []string // 预加载关联
}

// New 创建新的分页器
// db: 已构建好查询条件的 GORM 实例（支持多层链式调用）
// query: 分页查询参数
//
// 示例用法:
//
//	query := db.Where("id > ?", 3).Where("status = ?", "active")
//	pager := pagination.New[User](query, &pageQuery)
//	page, err := pager.Find()
//	if err != nil {
//	    // 处理错误
//	}
func New[T any](db *gorm.DB, query *PageQuery) *Paginator[T] {
	return &Paginator[T]{
		db:    db,
		query: query,
		page: &Page[T]{
			Records: make([]T, 0),
		},
	}
}

// Preload 添加预加载关联
// 支持链式调用，可以多次调用添加多个关联
//
// 示例:
//
//	pager.Preload("Orders").Preload("Profile").Find()
func (p *Paginator[T]) Preload(relations ...string) *Paginator[T] {
	p.preloads = append(p.preloads, relations...)
	return p
}

// Find 执行分页查询
// 返回 Page 结构体和错误信息
func (p *Paginator[T]) Find() (*Page[T], error) {
	// 1. 查询总数
	var total int64
	countDB := p.db.Session(&gorm.Session{})
	if err := countDB.Model(new(T)).Count(&total).Error; err != nil {
		p.err = err
		return nil, err
	}

	// 2. 构建分页查询
	queryDB := p.db

	// 添加预加载
	for _, relation := range p.preloads {
		queryDB = queryDB.Preload(relation)
	}

	// 添加排序（如果 PageQuery 中指定了）
	if orderBy := p.query.GetOrderBy(); orderBy != "" {
		queryDB = queryDB.Order(orderBy)
	}

	// 3. 应用分页
	pageSize := p.query.GetPageSize()
	pageNum := p.query.GetPageNum()
	queryDB = queryDB.
		Offset(p.query.GetOffset()).
		Limit(pageSize)

	// 4. 执行查询
	var records []T
	if err := queryDB.Find(&records).Error; err != nil {
		p.err = err
		return nil, err
	}

	// 5. 计算总页数
	pages := int64(0)
	if pageSize > 0 {
		pages = (total + int64(pageSize) - 1) / int64(pageSize)
	}

	// 6. 构造分页结果
	p.page = &Page[T]{
		Records: records,
		Total:   total,
		Size:    int64(pageSize),
		Current: int64(pageNum),
		Pages:   pages,
	}

	return p.page, nil
}

// GetPage 获取分页结果（兼容旧代码）
func (p *Paginator[T]) GetPage() *Page[T] {
	return p.page
}

// GetError 获取错误信息
func (p *Paginator[T]) GetError() error {
	return p.err
}

// Scopes 应用查询作用域（可选功能，用于复杂场景）
// 示例:
//
//	pager.Scopes(
//	    func(db *gorm.DB) *gorm.DB {
//	        return db.Where("created_at > ?", startTime)
//	    },
//	).Find()
func (p *Paginator[T]) Scopes(funcs ...func(*gorm.DB) *gorm.DB) *Paginator[T] {
	p.db = p.db.Scopes(funcs...)
	return p
}

// ============= 便捷函数 =============

// Paginate 快速分页函数（函数式 API）
// 返回 Page 结构体
//
// 示例:
//
//	page, err := pagination.Paginate[User](db.Where("id > ?", 3), &pageQuery)
func Paginate[T any](db *gorm.DB, query *PageQuery) (*Page[T], error) {
	pager := New[T](db, query)
	return pager.Find()
}

// PaginateWithPreload 带预加载的快速分页函数
// 示例:
//
//	page, err := pagination.PaginateWithPreload[User](
//	    db.Where("status = ?", "active"),
//	    &pageQuery,
//	    "Orders", "Profile",
//	)
func PaginateWithPreload[T any](db *gorm.DB, query *PageQuery, preloads ...string) (*Page[T], error) {
	pager := New[T](db, query)
	for _, preload := range preloads {
		pager.Preload(preload)
	}
	return pager.Find()
}
