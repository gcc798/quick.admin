package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LocalStorage 本地存储实现
type LocalStorage struct {
	basePath  string // 本地存储根目录
	urlPrefix string // URL 前缀
}

// LocalConfig 本地存储配置
type LocalConfig struct {
	BasePath  string `json:"basePath"`
	URLPrefix string `json:"urlPrefix"`
}

// NewLocalStorage 创建本地存储实例
func NewLocalStorage(config LocalConfig) (*LocalStorage, error) {
	// 确保目录存在
	if err := os.MkdirAll(config.BasePath, 0755); err != nil {
		return nil, fmt.Errorf("创建存储目录失败: %w", err)
	}

	return &LocalStorage{
		basePath:  config.BasePath,
		urlPrefix: config.URLPrefix,
	}, nil
}

// Upload 上传文件
func (l *LocalStorage) Upload(ctx context.Context, key string, reader io.Reader, size int64) error {
	// 构建完整路径
	fullPath := filepath.Join(l.basePath, key)

	// 确保目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 创建文件
	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	// 写入文件
	_, err = io.Copy(file, reader)
	if err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}

// Download 下载文件
func (l *LocalStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	fullPath := filepath.Join(l.basePath, key)

	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("文件不存在")
		}
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}

	return file, nil
}

// Delete 删除文件
func (l *LocalStorage) Delete(ctx context.Context, key string) error {
	fullPath := filepath.Join(l.basePath, key)

	err := os.Remove(fullPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除文件失败: %w", err)
	}

	return nil
}

// GetURL 获取访问 URL
func (l *LocalStorage) GetURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	// 本地存储返回静态 URL（不支持临时 URL）
	return fmt.Sprintf("%s/%s", strings.TrimRight(l.urlPrefix, "/"), key), nil
}

// Exists 检查文件是否存在
func (l *LocalStorage) Exists(ctx context.Context, key string) (bool, error) {
	fullPath := filepath.Join(l.basePath, key)

	_, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("检查文件是否存在失败: %w", err)
	}

	return true, nil
}

// List 列出文件
func (l *LocalStorage) List(ctx context.Context, prefix string) ([]string, error) {
	var keys []string
	searchPath := filepath.Join(l.basePath, prefix)

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			// 获取相对路径
			relPath, err := filepath.Rel(l.basePath, path)
			if err != nil {
				return err
			}
			// 转换为 Unix 风格路径
			relPath = filepath.ToSlash(relPath)
			keys = append(keys, relPath)
		}

		return nil
	})

	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("列出文件失败: %w", err)
	}

	return keys, nil
}

// GetInfo 获取文件信息
func (l *LocalStorage) GetInfo(ctx context.Context, key string) (*FileInfo, error) {
	fullPath := filepath.Join(l.basePath, key)

	stat, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("文件不存在")
		}
		return nil, fmt.Errorf("获取文件信息失败: %w", err)
	}

	return &FileInfo{
		Key:          key,
		Size:         stat.Size(),
		LastModified: stat.ModTime(),
		ContentType:  "application/octet-stream",
		ETag:         "",
	}, nil
}

// LocalStorageFactory 本地存储工厂
type LocalStorageFactory struct{}

func NewLocalStorageFactory() *LocalStorageFactory {
	return &LocalStorageFactory{}
}

func (f *LocalStorageFactory) Create(config map[string]interface{}) (Storage, error) {
	localConfig := LocalConfig{
		BasePath:  getStringValue(config, "basePath", "/tmp/storage"),
		URLPrefix: getStringValue(config, "urlPrefix", "http://localhost:8080/files"),
	}

	return NewLocalStorage(localConfig)
}
