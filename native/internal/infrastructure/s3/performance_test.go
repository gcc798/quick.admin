package s3

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/force-c/nai-tizi/internal/logger"
)

// LargeFileReader 大文件流式读取器（避免内存溢出）
type LargeFileReader struct {
	size      int64 // 文件大小
	read      int64 // 已读取字节数
	chunkSize int   // 每次读取的块大小
}

// NewLargeFileReader 创建大文件读取器
func NewLargeFileReader(size int64) *LargeFileReader {
	return &LargeFileReader{
		size:      size,
		read:      0,
		chunkSize: 1024 * 1024, // 1MB 块
	}
}

// Read 实现 io.Reader 接口
func (r *LargeFileReader) Read(p []byte) (n int, err error) {
	if r.read >= r.size {
		return 0, io.EOF
	}

	// 计算本次应该读取的字节数
	remaining := r.size - r.read
	toRead := int64(len(p))
	if toRead > remaining {
		toRead = remaining
	}

	// 生成随机数据
	n64, err := rand.Read(p[:toRead])
	n = int(n64)
	r.read += int64(n)

	return n, err
}

// TestConcurrentUpload 并发上传性能测试
func TestConcurrentUpload(t *testing.T) {
	// 配置
	const (
		fileSize      = 5 * 1024 * 1024 * 1024 // 5GB
		fileSizeMB    = 5 * 1024               // 5GB in MB
		concurrency   = 3                      // 并发数
		testFileCount = concurrency            // 测试文件数量
	)

	log, err := logger.NewLogger("development")
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

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

	t.Logf("=== 并发上传性能测试 ===")
	t.Logf("文件大小: %d MB (%.2f GB)", fileSizeMB, float64(fileSizeMB)/1024)
	t.Logf("并发数: %d", concurrency)
	t.Logf("测试文件数: %d", testFileCount)
	t.Logf("总数据量: %.2f GB", float64(fileSizeMB*testFileCount)/1024)

	// 用于同步和收集结果
	var wg sync.WaitGroup
	type uploadResult struct {
		fileNum  int
		duration time.Duration
		err      error
	}
	resultChan := make(chan uploadResult, testFileCount)

	// 记录总开始时间
	totalStartTime := time.Now()

	// 启动并发上传
	for i := 0; i < testFileCount; i++ {
		wg.Add(1)
		go func(fileNum int) {
			defer wg.Done()

			key := fmt.Sprintf("perf-test/large-file-%d.bin", fileNum)
			reader := NewLargeFileReader(fileSize)

			t.Logf("[协程 %d] 开始上传文件: %s", fileNum, key)
			startTime := time.Now()

			err := manager.UploadFile(ctx, key, reader, "application/octet-stream")

			duration := time.Since(startTime)
			resultChan <- uploadResult{
				fileNum:  fileNum,
				duration: duration,
				err:      err,
			}

			if err != nil {
				t.Logf("[协程 %d] ❌ 上传失败: %v", fileNum, err)
			} else {
				speedMBps := float64(fileSizeMB) / duration.Seconds()
				t.Logf("[协程 %d] ✓ 上传成功 | 耗时: %s | 速度: %.2f MB/s",
					fileNum, duration.Round(time.Millisecond), speedMBps)
			}
		}(i)
	}

	// 等待所有上传完成
	wg.Wait()
	close(resultChan)

	totalDuration := time.Since(totalStartTime)

	// 收集结果
	var successCount, failCount int
	var totalUploadTime time.Duration
	var minDuration, maxDuration time.Duration

	for result := range resultChan {
		if result.err != nil {
			failCount++
		} else {
			successCount++
			totalUploadTime += result.duration

			if minDuration == 0 || result.duration < minDuration {
				minDuration = result.duration
			}
			if result.duration > maxDuration {
				maxDuration = result.duration
			}
		}
	}

	// 输出性能统计
	t.Logf("\n=== 性能统计 ===")
	t.Logf("总耗时: %s", totalDuration.Round(time.Millisecond))
	t.Logf("成功数: %d / %d", successCount, testFileCount)
	t.Logf("失败数: %d", failCount)

	if successCount > 0 {
		avgDuration := totalUploadTime / time.Duration(successCount)
		totalDataMB := float64(fileSizeMB * successCount)
		avgSpeedMBps := totalDataMB / totalDuration.Seconds()
		peakSpeedMBps := float64(fileSizeMB) / minDuration.Seconds()

		t.Logf("\n平均单文件耗时: %s", avgDuration.Round(time.Millisecond))
		t.Logf("最快上传耗时: %s", minDuration.Round(time.Millisecond))
		t.Logf("最慢上传耗时: %s", maxDuration.Round(time.Millisecond))
		t.Logf("\n平均上传速度: %.2f MB/s", avgSpeedMBps)
		t.Logf("峰值上传速度: %.2f MB/s", peakSpeedMBps)
		t.Logf("总数据量: %.2f GB", totalDataMB/1024)
		t.Logf("并发效率: %.2f%%", (avgDuration.Seconds()/totalDuration.Seconds())*100*float64(concurrency))
	}

	//清理测试文件
	t.Logf("\n=== 清理测试文件 ===")
	for i := 0; i < testFileCount; i++ {
		key := fmt.Sprintf("perf-test/large-file-%d.bin", i)
		if err := manager.DeleteFile(ctx, key); err != nil {
			t.Logf("清理文件失败 %s: %v", key, err)
		} else {
			t.Logf("✓ 已删除: %s", key)
		}
	}

	if failCount > 0 {
		t.Errorf("有 %d 个文件上传失败", failCount)
	}
}

// TestConcurrentUploadWithConfig 可配置的并发上传测试
func TestConcurrentUploadWithConfig(t *testing.T) {
	// 允许通过环境变量或测试参数配置
	type testConfig struct {
		fileSizeGB  int // 单文件大小(GB)
		concurrency int // 并发数
	}

	configs := []testConfig{
		{fileSizeGB: 1, concurrency: 2}, // 1GB x 2
		{fileSizeGB: 1, concurrency: 3}, // 1GB x 3
		{fileSizeGB: 2, concurrency: 2}, // 2GB x 2
		{fileSizeGB: 5, concurrency: 1}, // 5GB x 1 (基准)
	}

	for _, cfg := range configs {
		testName := fmt.Sprintf("%dGBx%d", cfg.fileSizeGB, cfg.concurrency)
		t.Run(testName, func(t *testing.T) {
			runConcurrentUploadTest(t, int64(cfg.fileSizeGB)*1024*1024*1024, cfg.concurrency)
		})
	}
}

// runConcurrentUploadTest 执行并发上传测试的辅助函数
func runConcurrentUploadTest(t *testing.T, fileSize int64, concurrency int) {
	log, err := logger.NewLogger("development")
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

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
	fileSizeMB := fileSize / (1024 * 1024)

	t.Logf("文件大小: %d MB | 并发数: %d", fileSizeMB, concurrency)

	var wg sync.WaitGroup
	totalStartTime := time.Now()

	// 并发上传
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(fileNum int) {
			defer wg.Done()

			key := fmt.Sprintf("perf-test/config-test-%d.bin", fileNum)
			reader := NewLargeFileReader(fileSize)

			startTime := time.Now()
			err := manager.UploadFile(ctx, key, reader, "application/octet-stream")
			duration := time.Since(startTime)

			if err != nil {
				t.Logf("[%d] ❌ 失败: %v", fileNum, err)
			} else {
				speedMBps := float64(fileSizeMB) / duration.Seconds()
				t.Logf("[%d] ✓ 成功 | %s | %.2f MB/s", fileNum, duration.Round(time.Millisecond), speedMBps)
			}

			// 清理
			_ = manager.DeleteFile(ctx, key)
		}(i)
	}

	wg.Wait()
	totalDuration := time.Since(totalStartTime)

	totalDataGB := float64(fileSizeMB*int64(concurrency)) / 1024
	avgSpeedMBps := float64(fileSizeMB*int64(concurrency)) / totalDuration.Seconds()

	t.Logf("总耗时: %s | 总数据: %.2f GB | 平均速度: %.2f MB/s",
		totalDuration.Round(time.Millisecond), totalDataGB, avgSpeedMBps)
}

// TestStreamUpload 测试流式上传（小内存占用）
func TestStreamUpload(t *testing.T) {
	log, _ := logger.NewLogger("development")

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

	manager, _ := NewManager(config, log)
	ctx := context.Background()

	// 测试 100MB 流式上传
	fileSize := int64(100 * 1024 * 1024) // 100MB
	key := "perf-test/stream-upload.bin"

	t.Logf("测试流式上传 100MB 文件...")
	reader := NewLargeFileReader(fileSize)

	startTime := time.Now()
	err := manager.UploadFile(ctx, key, reader, "application/octet-stream")
	duration := time.Since(startTime)

	if err != nil {
		t.Errorf("上传失败: %v", err)
	} else {
		speedMBps := 100.0 / duration.Seconds()
		t.Logf("✓ 上传成功 | 耗时: %s | 速度: %.2f MB/s",
			duration.Round(time.Millisecond), speedMBps)
	}

	// 清理
	_ = manager.DeleteFile(ctx, key)
}
