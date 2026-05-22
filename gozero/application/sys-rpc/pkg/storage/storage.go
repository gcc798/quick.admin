package storage

import (
	"context"
	"io"
	"time"
)

// Storage defines the unified interface for file storage operations.
type Storage interface {
	Upload(ctx context.Context, key string, reader io.Reader, size int64) error
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	GetURL(ctx context.Context, key string, expires time.Duration) (string, error)
	Exists(ctx context.Context, key string) (bool, error)
}
