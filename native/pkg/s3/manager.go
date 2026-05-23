package s3

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gcc798/quick.admin/internal/logger"
	"go.uber.org/zap"
)

// Config S3配置
type Config struct {
	Enabled         bool   // 是否启用
	Endpoint        string // MinIO/S3服务器地址
	AccessKeyID     string // 访问密钥ID
	SecretAccessKey string // 访问密钥
	Region          string // 区域
	Bucket          string // 默认存储桶
	UseSSL          bool   // 是否使用SSL
	ForcePathStyle  bool   // 强制路径样式（MinIO需要）
}

// StorageProvider 存储提供者接口
type StorageProvider interface {
	UploadFile(ctx context.Context, key string, reader io.Reader, contentType string) error
	DownloadFile(ctx context.Context, key string) (io.ReadCloser, error)
	DeleteFile(ctx context.Context, key string) error
	FileExists(ctx context.Context, key string) (bool, error)
	GetPresignedURL(ctx context.Context, key string, expiration time.Duration) (string, error)
	ListFiles(ctx context.Context, prefix string, maxKeys int32) ([]string, error)
}

// UploadCallback 上传结果回调
type UploadCallback func(ctx context.Context, key string, err error)

// Manager S3存储管理器
type Manager struct {
	client          *s3.Client
	config          *Config
	logger          logger.Logger
	onUploadSuccess []UploadCallback
	onUploadFailure []UploadCallback
}

// AddOnUploadSuccess 添加上传成功回调
func (m *Manager) AddOnUploadSuccess(cb UploadCallback) {
	m.onUploadSuccess = append(m.onUploadSuccess, cb)
}

// AddOnUploadFailure 添加上传失败回调
func (m *Manager) AddOnUploadFailure(cb UploadCallback) {
	m.onUploadFailure = append(m.onUploadFailure, cb)
}

// NewManager 创建S3管理器
func NewManager(cfg *Config, log logger.Logger) (*Manager, error) {
	if !cfg.Enabled {
		log.Info("S3 storage is disabled")
		return &Manager{config: cfg, logger: log}, nil
	}

	// 构建endpoint URL
	endpoint := cfg.Endpoint
	if cfg.UseSSL {
		endpoint = "https://" + endpoint
	} else {
		endpoint = "http://" + endpoint
	}

	// 创建自定义解析器
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               endpoint,
			HostnameImmutable: true,
			SigningRegion:     cfg.Region,
		}, nil
	})

	// 加载AWS配置
	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.Region),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// 创建S3客户端
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = cfg.ForcePathStyle
	})

	log.Info("S3 storage initialized successfully",
		zap.String("endpoint", cfg.Endpoint),
		zap.String("bucket", cfg.Bucket))

	return &Manager{
		client: client,
		config: cfg,
		logger: log,
	}, nil
}

// IsEnabled 检查是否启用
func (m *Manager) IsEnabled() bool {
	return m.config.Enabled
}

// UploadFile 上传文件
// key: 文件在S3中的路径（例如: "images/avatar.jpg"）
// reader: 文件内容读取器
// contentType: 文件MIME类型（例如: "image/jpeg"）
func (m *Manager) UploadFile(ctx context.Context, key string, reader io.Reader, contentType string) error {
	if !m.config.Enabled {
		return fmt.Errorf("S3 storage is disabled")
	}

	// 使用 S3 Manager 进行分片上传，支持大文件和流式上传
	uploader := manager.NewUploader(m.client, func(u *manager.Uploader) {
		// 设置分片大小为 10MB
		u.PartSize = 10 * 1024 * 1024
		// 设置并发数
		u.Concurrency = 5
	})

	input := &s3.PutObjectInput{
		Bucket:      aws.String(m.config.Bucket),
		Key:         aws.String(key),
		Body:        reader,
		ContentType: aws.String(contentType),
	}

	_, err := uploader.Upload(ctx, input)
	if err != nil {
		m.logger.Error("failed to upload file to S3",
			zap.String("key", key),
			zap.Error(err))

		// 触发失败回调
		for _, cb := range m.onUploadFailure {
			cb(ctx, key, err)
		}

		return fmt.Errorf("failed to upload file: %w", err)
	}

	m.logger.Info("file uploaded successfully",
		zap.String("key", key),
		zap.String("bucket", m.config.Bucket))

	// 触发成功回调
	for _, cb := range m.onUploadSuccess {
		cb(ctx, key, nil)
	}

	return nil
}

// DownloadFile 下载文件
func (m *Manager) DownloadFile(ctx context.Context, key string) (io.ReadCloser, error) {
	if !m.config.Enabled {
		return nil, fmt.Errorf("S3 storage is disabled")
	}

	input := &s3.GetObjectInput{
		Bucket: aws.String(m.config.Bucket),
		Key:    aws.String(key),
	}

	result, err := m.client.GetObject(ctx, input)
	if err != nil {
		m.logger.Error("failed to download file from S3",
			zap.String("key", key),
			zap.Error(err))
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	return result.Body, nil
}

// DeleteFile 删除文件
func (m *Manager) DeleteFile(ctx context.Context, key string) error {
	if !m.config.Enabled {
		return fmt.Errorf("S3 storage is disabled")
	}

	input := &s3.DeleteObjectInput{
		Bucket: aws.String(m.config.Bucket),
		Key:    aws.String(key),
	}

	_, err := m.client.DeleteObject(ctx, input)
	if err != nil {
		m.logger.Error("failed to delete file from S3",
			zap.String("key", key),
			zap.Error(err))
		return fmt.Errorf("failed to delete file: %w", err)
	}

	m.logger.Info("file deleted successfully",
		zap.String("key", key))

	return nil
}

// FileExists 检查文件是否存在
func (m *Manager) FileExists(ctx context.Context, key string) (bool, error) {
	if !m.config.Enabled {
		return false, fmt.Errorf("S3 storage is disabled")
	}

	input := &s3.HeadObjectInput{
		Bucket: aws.String(m.config.Bucket),
		Key:    aws.String(key),
	}

	_, err := m.client.HeadObject(ctx, input)
	if err != nil {
		// 文件不存在不算错误
		return false, nil
	}

	return true, nil
}

// GetPresignedURL 获取预签名URL（用于临时访问）
func (m *Manager) GetPresignedURL(ctx context.Context, key string, expiration time.Duration) (string, error) {
	if !m.config.Enabled {
		return "", fmt.Errorf("S3 storage is disabled")
	}

	presignClient := s3.NewPresignClient(m.client)

	input := &s3.GetObjectInput{
		Bucket: aws.String(m.config.Bucket),
		Key:    aws.String(key),
	}

	result, err := presignClient.PresignGetObject(ctx, input, func(opts *s3.PresignOptions) {
		opts.Expires = expiration
	})
	if err != nil {
		m.logger.Error("failed to generate presigned URL",
			zap.String("key", key),
			zap.Error(err))
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return result.URL, nil
}

// ListFiles 列出文件
func (m *Manager) ListFiles(ctx context.Context, prefix string, maxKeys int32) ([]string, error) {
	if !m.config.Enabled {
		return nil, fmt.Errorf("S3 storage is disabled")
	}

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(m.config.Bucket),
		Prefix: aws.String(prefix),
	}

	if maxKeys > 0 {
		input.MaxKeys = aws.Int32(maxKeys)
	}

	result, err := m.client.ListObjectsV2(ctx, input)
	if err != nil {
		m.logger.Error("failed to list files from S3",
			zap.String("prefix", prefix),
			zap.Error(err))
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	var keys []string
	for _, obj := range result.Contents {
		keys = append(keys, *obj.Key)
	}

	return keys, nil
}

// CopyFile 复制文件
func (m *Manager) CopyFile(ctx context.Context, sourceKey, destKey string) error {
	if !m.config.Enabled {
		return fmt.Errorf("S3 storage is disabled")
	}

	copySource := fmt.Sprintf("%s/%s", m.config.Bucket, sourceKey)
	input := &s3.CopyObjectInput{
		Bucket:     aws.String(m.config.Bucket),
		CopySource: aws.String(copySource),
		Key:        aws.String(destKey),
	}

	_, err := m.client.CopyObject(ctx, input)
	if err != nil {
		m.logger.Error("failed to copy file",
			zap.String("source", sourceKey),
			zap.String("dest", destKey),
			zap.Error(err))
		return fmt.Errorf("failed to copy file: %w", err)
	}

	m.logger.Info("file copied successfully",
		zap.String("source", sourceKey),
		zap.String("dest", destKey))

	return nil
}

// CreateBucket 创建存储桶（如果不存在）
func (m *Manager) CreateBucket(ctx context.Context, bucketName string) error {
	if !m.config.Enabled {
		return fmt.Errorf("S3 storage is disabled")
	}

	input := &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	}

	_, err := m.client.CreateBucket(ctx, input)
	if err != nil {
		m.logger.Error("failed to create bucket",
			zap.String("bucket", bucketName),
			zap.Error(err))
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	m.logger.Info("bucket created successfully",
		zap.String("bucket", bucketName))

	return nil
}

// GetDefaultBucket 获取默认存储桶名称
func (m *Manager) GetDefaultBucket() string {
	return m.config.Bucket
}
