package storage

import (
	"context"
	"io"
	"time"
)

// Storage 统一存储接口
type Storage interface {
	// Upload 上传文件
	Upload(ctx context.Context, key string, reader io.Reader, size int64) error

	// Download 下载文件
	Download(ctx context.Context, key string) (io.ReadCloser, error)

	// Delete 删除文件
	Delete(ctx context.Context, key string) error

	// GetURL 获取访问 URL
	// expires: 过期时间，0 表示永久 URL（公开文件）
	GetURL(ctx context.Context, key string, expires time.Duration) (string, error)

	// Exists 检查文件是否存在
	Exists(ctx context.Context, key string) (bool, error)

	// List 列出文件
	List(ctx context.Context, prefix string) ([]string, error)

	// GetInfo 获取文件信息
	GetInfo(ctx context.Context, key string) (*FileInfo, error)
}

// FileInfo 文件信息
type FileInfo struct {
	Key          string    // 文件 Key
	Size         int64     // 文件大小
	LastModified time.Time // 最后修改时间
	ContentType  string    // 内容类型
	ETag         string    // ETag
}

// StorageFactory 存储工厂接口
type StorageFactory interface {
	// Create 根据配置创建 Storage 实例
	Create(config map[string]interface{}) (Storage, error)
}
