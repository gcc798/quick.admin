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

// LocalStorage implements Storage using the local filesystem.
type LocalStorage struct {
	basePath  string
	urlPrefix string
}

// LocalConfig holds configuration for local filesystem storage.
type LocalConfig struct {
	BasePath  string `json:"basePath"`
	URLPrefix string `json:"urlPrefix"`
}

// NewLocalStorage creates a LocalStorage instance, ensuring the base directory exists.
func NewLocalStorage(config LocalConfig) (*LocalStorage, error) {
	if err := os.MkdirAll(config.BasePath, 0o755); err != nil {
		return nil, fmt.Errorf("create storage dir failed: %w", err)
	}
	return &LocalStorage{basePath: config.BasePath, urlPrefix: config.URLPrefix}, nil
}

func (l *LocalStorage) Upload(ctx context.Context, key string, reader io.Reader, size int64) error {
	fullPath := filepath.Join(l.basePath, key)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create dir failed: %w", err)
	}
	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("create file failed: %w", err)
	}
	defer file.Close()
	if _, err := io.Copy(file, reader); err != nil {
		return fmt.Errorf("write file failed: %w", err)
	}
	return nil
}

func (l *LocalStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	fullPath := filepath.Join(l.basePath, key)
	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found")
		}
		return nil, fmt.Errorf("open file failed: %w", err)
	}
	return file, nil
}

func (l *LocalStorage) Delete(ctx context.Context, key string) error {
	fullPath := filepath.Join(l.basePath, key)
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete file failed: %w", err)
	}
	return nil
}

func (l *LocalStorage) GetURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	return fmt.Sprintf("%s/%s", strings.TrimRight(l.urlPrefix, "/"), key), nil
}

func (l *LocalStorage) Exists(ctx context.Context, key string) (bool, error) {
	fullPath := filepath.Join(l.basePath, key)
	_, err := os.Stat(fullPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
