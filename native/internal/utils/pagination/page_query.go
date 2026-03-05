package pagination

import (
	"fmt"
	"strings"
)

const (
	// DefaultPageNum 默认页码
	DefaultPageNum = 1
	// DefaultPageSize 默认每页大小
	DefaultPageSize = 10
	// MaxPageSize 最大每页大小
	MaxPageSize = 1000
)

// PageQuery 分页查询参数
// 可嵌入到各个请求结构体中，避免重复定义分页参数
type PageQuery struct {
	// PageNum 当前页码，从1开始
	PageNum int `form:"pageNum" json:"pageNum"`
	// PageSize 每页数量
	PageSize int `form:"pageSize" json:"pageSize"`
	// OrderByColumn 排序列，支持多列排序，逗号分隔，如: "id,createTime"
	OrderByColumn string `form:"orderByColumn" json:"orderByColumn"`
	// IsAsc 排序方向，asc或desc，支持多个，逗号分隔，如: "asc,desc"
	IsAsc string `form:"isAsc" json:"isAsc"`
}

// GetPageNum 获取页码，如果未设置或小于1，返回默认值
func (p *PageQuery) GetPageNum() int {
	if p.PageNum <= 0 {
		return DefaultPageNum
	}
	return p.PageNum
}

// GetPageSize 获取每页大小，如果未设置或小于1，返回默认值；如果超过最大值，返回最大值
func (p *PageQuery) GetPageSize() int {
	if p.PageSize <= 0 {
		return DefaultPageSize
	}
	if p.PageSize > MaxPageSize {
		return MaxPageSize
	}
	return p.PageSize
}

// GetLimit 获取查询限制数量（用于SQL LIMIT）
func (p *PageQuery) GetLimit() int {
	return p.GetPageSize()
}

// GetOffset 获取查询偏移量（用于SQL OFFSET）
func (p *PageQuery) GetOffset() int {
	return (p.GetPageNum() - 1) * p.GetPageSize()
}

// GetOrderBy 构建ORDER BY子句
// 返回格式示例: "id ASC, create_time DESC"
// 如果没有指定排序，返回空字符串
func (p *PageQuery) GetOrderBy() string {
	if p.OrderByColumn == "" || p.IsAsc == "" {
		return ""
	}

	// 转换为小写并替换 ascending/descending 为 asc/desc
	isAsc := strings.ToLower(p.IsAsc)
	isAsc = strings.ReplaceAll(isAsc, "ascending", "asc")
	isAsc = strings.ReplaceAll(isAsc, "descending", "desc")

	columns := strings.Split(p.OrderByColumn, ",")
	directions := strings.Split(isAsc, ",")

	// 如果方向只有一个，应用到所有列
	if len(directions) == 1 {
		direction := strings.TrimSpace(directions[0])
		if direction != "asc" && direction != "desc" {
			return "" // 无效的排序方向
		}
		var parts []string
		for _, col := range columns {
			col = strings.TrimSpace(col)
			if col != "" {
				parts = append(parts, fmt.Sprintf("%s %s", toSnakeCase(col), strings.ToUpper(direction)))
			}
		}
		return strings.Join(parts, ", ")
	}

	// 如果方向数量与列数量不匹配，返回空
	if len(directions) != len(columns) {
		return ""
	}

	// 每列有各自的排序方向
	var parts []string
	for i, col := range columns {
		col = strings.TrimSpace(col)
		direction := strings.TrimSpace(directions[i])
		if col != "" && (direction == "asc" || direction == "desc") {
			parts = append(parts, fmt.Sprintf("%s %s", toSnakeCase(col), strings.ToUpper(direction)))
		}
	}

	return strings.Join(parts, ", ")
}

// toSnakeCase 将驼峰命名转换为蛇形命名
// 例如: createTime -> create_time
func toSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, r)
	}
	return strings.ToLower(string(result))
}
