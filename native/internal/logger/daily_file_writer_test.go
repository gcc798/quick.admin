package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDailyFileWriterRotateOnDateChange(t *testing.T) {
	dir := t.TempDir()
	writer := newDailyFileWriter(FileConfig{
		Path:       dir,
		Filename:   "lsh-api",
		MaxSize:    100,
		MaxBackups: 30,
		MaxAge:     7,
	})

	firstDay := time.Date(2026, 4, 23, 23, 59, 59, 0, time.Local)
	secondDay := firstDay.Add(2 * time.Second)

	writer.now = func() time.Time { return firstDay }
	if _, err := writer.Write([]byte("day1\n")); err != nil {
		t.Fatalf("first write failed: %v", err)
	}

	writer.now = func() time.Time { return secondDay }
	if _, err := writer.Write([]byte("day2\n")); err != nil {
		t.Fatalf("second write failed: %v", err)
	}

	firstContent, err := os.ReadFile(filepath.Join(dir, "lsh-api-2026-04-23.log"))
	if err != nil {
		t.Fatalf("read first file failed: %v", err)
	}
	if !strings.Contains(string(firstContent), "day1") {
		t.Fatalf("first file missing first day log: %q", string(firstContent))
	}
	if strings.Contains(string(firstContent), "day2") {
		t.Fatalf("first file should not contain second day log: %q", string(firstContent))
	}

	secondContent, err := os.ReadFile(filepath.Join(dir, "lsh-api-2026-04-24.log"))
	if err != nil {
		t.Fatalf("read second file failed: %v", err)
	}
	if !strings.Contains(string(secondContent), "day2") {
		t.Fatalf("second file missing second day log: %q", string(secondContent))
	}
}

func TestDailyFileWriterReuseSameDateFile(t *testing.T) {
	dir := t.TempDir()
	writer := newDailyFileWriter(FileConfig{
		Path:       dir,
		Filename:   "lsh-api",
		MaxSize:    100,
		MaxBackups: 30,
		MaxAge:     7,
	})

	now := time.Date(2026, 4, 24, 8, 12, 12, 0, time.Local)
	writer.now = func() time.Time { return now }

	if _, err := writer.Write([]byte("first\n")); err != nil {
		t.Fatalf("first write failed: %v", err)
	}
	if _, err := writer.Write([]byte("second\n")); err != nil {
		t.Fatalf("second write failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(dir, "lsh-api-2026-04-24.log"))
	if err != nil {
		t.Fatalf("read log file failed: %v", err)
	}

	got := string(content)
	if !strings.Contains(got, "first") || !strings.Contains(got, "second") {
		t.Fatalf("same-day logs should stay in one file: %q", got)
	}
}
