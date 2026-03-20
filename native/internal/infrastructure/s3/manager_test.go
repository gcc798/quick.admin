package s3

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/force-c/nai-tizi/internal/logger"
)

// TestManager 测试S3 Manager功能
func TestManager(t *testing.T) {
	// 创建logger
	log, err := logger.NewLogger("development")
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	// 创建S3 Manager
	config := &Config{
		Enabled:         true,
		Endpoint:        "localhost:9000",
		AccessKeyID:     "IOx2ruSY4QfWcA3CvqRK",
		SecretAccessKey: "34MkLbW6sf5da7jHrlTFVJ209KDPqunBQmtSxOIg",
		Region:          "auto",
		Bucket:          "door",
		UseSSL:          false,
		ForcePathStyle:  true,
	}

	manager, err := NewManager(config, log)
	if err != nil {
		t.Fatalf("failed to create S3 manager: %v", err)
	}

	ctx := context.Background()

	// 测试用的文件内容
	testContent := "Hello, RustFS! This is a test file."
	testKey := "test/unit_test_file.txt"

	// 1. 测试上传文件
	t.Run("UploadFile", func(t *testing.T) {
		reader := strings.NewReader(testContent)
		err := manager.UploadFile(ctx, testKey, reader, "text/plain")
		if err != nil {
			t.Errorf("failed to upload file: %v", err)
		}
		t.Logf("✓ 文件上传成功: %s", testKey)
	})

	// 2. 测试文件是否存在
	t.Run("FileExists", func(t *testing.T) {
		exists, err := manager.FileExists(ctx, testKey)
		if err != nil {
			t.Errorf("failed to check file existence: %v", err)
		}
		if !exists {
			t.Error("file should exist but doesn't")
		}
		t.Logf("✓ 文件存在检查通过")
	})

	// 3. 测试下载文件
	t.Run("DownloadFile", func(t *testing.T) {
		reader, err := manager.DownloadFile(ctx, testKey)
		if err != nil {
			t.Errorf("failed to download file: %v", err)
			return
		}
		defer reader.Close()

		// 读取内容
		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, reader)
		if err != nil {
			t.Errorf("failed to read downloaded content: %v", err)
			return
		}

		downloadedContent := buf.String()
		if downloadedContent != testContent {
			t.Errorf("content mismatch: expected %q, got %q", testContent, downloadedContent)
		}
		t.Logf("✓ 文件下载成功，内容匹配")
	})

	// 4. 测试获取预签名URL
	t.Run("GetPresignedURL", func(t *testing.T) {
		url, err := manager.GetPresignedURL(ctx, testKey, 1*time.Hour)
		if err != nil {
			t.Errorf("failed to get presigned URL: %v", err)
			return
		}
		if url == "" {
			t.Error("presigned URL is empty")
		}
		t.Logf("✓ 预签名URL生成成功: %s", url)
	})

	// 5. 测试列出文件
	t.Run("ListFiles", func(t *testing.T) {
		files, err := manager.ListFiles(ctx, "test/", 10)
		if err != nil {
			t.Errorf("failed to list files: %v", err)
			return
		}
		found := false
		for _, file := range files {
			if file == testKey {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("uploaded file not found in list")
		}
		t.Logf("✓ 文件列表查询成功，找到 %d 个文件", len(files))
	})

	// 6. 测试复制文件
	t.Run("CopyFile", func(t *testing.T) {
		destKey := "test/unit_test_file_copy.txt"
		err := manager.CopyFile(ctx, testKey, destKey)
		if err != nil {
			t.Errorf("failed to copy file: %v", err)
			return
		}

		// 验证复制的文件存在
		exists, err := manager.FileExists(ctx, destKey)
		if err != nil {
			t.Errorf("failed to check copied file: %v", err)
			return
		}
		if !exists {
			t.Error("copied file should exist")
		}
		t.Logf("✓ 文件复制成功: %s -> %s", testKey, destKey)

		// 清理复制的文件
		_ = manager.DeleteFile(ctx, destKey)
	})

	// 7. 测试删除文件
	t.Run("DeleteFile", func(t *testing.T) {
		err := manager.DeleteFile(ctx, testKey)
		if err != nil {
			t.Errorf("failed to delete file: %v", err)
			return
		}

		// 验证文件已被删除
		exists, _ := manager.FileExists(ctx, testKey)
		if exists {
			t.Error("file should be deleted but still exists")
		}
		t.Logf("✓ 文件删除成功")
	})

	// 8. 测试不存在的文件
	t.Run("NonExistentFile", func(t *testing.T) {
		exists, err := manager.FileExists(ctx, "non-existent-file.txt")
		if err != nil {
			t.Errorf("failed to check non-existent file: %v", err)
			return
		}
		if exists {
			t.Error("non-existent file should not exist")
		}
		t.Logf("✓ 不存在文件检查通过")
	})
}

// TestManagerDisabled 测试禁用状态
func TestManagerDisabled(t *testing.T) {
	log, _ := logger.NewLogger("development")

	config := &Config{
		Enabled: false,
	}

	manager, err := NewManager(config, log)
	if err != nil {
		t.Fatalf("failed to create disabled manager: %v", err)
	}

	ctx := context.Background()

	// 测试禁用状态下的操作应该返回错误
	t.Run("UploadWhenDisabled", func(t *testing.T) {
		err := manager.UploadFile(ctx, "test.txt", strings.NewReader("test"), "text/plain")
		if err == nil {
			t.Error("should return error when S3 is disabled")
		}
		t.Logf("✓ 禁用状态检查通过: %v", err)
	})
}

// TestBatchOperations 测试批量操作
func TestBatchOperations(t *testing.T) {
	log, _ := logger.NewLogger("development")

	config := &Config{
		Enabled:         true,
		Endpoint:        "localhost:9000",
		AccessKeyID:     "etH5Yf4NQkBZoSF3IUXu",
		SecretAccessKey: "htgqTA74sEuel8JVPrFjZcpbU6QG3CNMvmY19X5L",
		Region:          "auto",
		Bucket:          "door",
		UseSSL:          false,
		ForcePathStyle:  true,
	}

	manager, err := NewManager(config, log)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	ctx := context.Background()

	// 批量上传测试文件
	t.Run("BatchUpload", func(t *testing.T) {
		fileCount := 5
		for i := 0; i < fileCount; i++ {
			key := strings.Replace("test/batch/file_{i}.txt", "{i}", string(rune('0'+i)), 1)
			content := strings.Replace("Test content {i}", "{i}", string(rune('0'+i)), 1)
			err := manager.UploadFile(ctx, key, strings.NewReader(content), "text/plain")
			if err != nil {
				t.Errorf("failed to upload file %d: %v", i, err)
			}
		}
		t.Logf("✓ 批量上传 %d 个文件成功", fileCount)
	})

	// 列出批量上传的文件
	t.Run("ListBatchFiles", func(t *testing.T) {
		files, err := manager.ListFiles(ctx, "test/batch/", 10)
		if err != nil {
			t.Errorf("failed to list batch files: %v", err)
			return
		}
		t.Logf("✓ 找到 %d 个批量上传的文件", len(files))
		for _, file := range files {
			t.Logf("  - %s", file)
		}
	})

	// 清理批量上传的文件
	t.Run("BatchCleanup", func(t *testing.T) {
		files, _ := manager.ListFiles(ctx, "test/batch/", 10)
		for _, file := range files {
			_ = manager.DeleteFile(ctx, file)
		}
		t.Logf("✓ 清理 %d 个测试文件", len(files))
	})
}

// BenchmarkUploadFile 性能测试：文件上传
func BenchmarkUploadFile(b *testing.B) {
	log, _ := logger.NewLogger("development")

	config := &Config{
		Enabled:         true,
		Endpoint:        "localhost:9000",
		AccessKeyID:     "etH5Yf4NQkBZoSF3IUXu",
		SecretAccessKey: "htgqTA74sEuel8JVPrFjZcpbU6QG3CNMvmY19X5L",
		Region:          "auto",
		Bucket:          "door",
		UseSSL:          false,
		ForcePathStyle:  true,
	}

	manager, _ := NewManager(config, log)
	ctx := context.Background()
	content := "Benchmark test content"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := strings.Replace("benchmark/file_{i}.txt", "{i}", string(rune('0'+i%10)), 1)
		_ = manager.UploadFile(ctx, key, strings.NewReader(content), "text/plain")
	}
}

// BenchmarkDownloadFile 性能测试：文件下载
func BenchmarkDownloadFile(b *testing.B) {
	log, _ := logger.NewLogger("development")

	config := &Config{
		Enabled:         true,
		Endpoint:        "localhost:9000",
		AccessKeyID:     "etH5Yf4NQkBZoSF3IUXu",
		SecretAccessKey: "htgqTA74sEuel8JVPrFjZcpbU6QG3CNMvmY19X5L",
		Region:          "auto",
		Bucket:          "door",
		UseSSL:          false,
		ForcePathStyle:  true,
	}

	manager, _ := NewManager(config, log)
	ctx := context.Background()

	// 先上传一个文件
	testKey := "benchmark/download_test.txt"
	_ = manager.UploadFile(ctx, testKey, strings.NewReader("Download test"), "text/plain")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader, _ := manager.DownloadFile(ctx, testKey)
		if reader != nil {
			io.Copy(io.Discard, reader)
			reader.Close()
		}
	}
}
