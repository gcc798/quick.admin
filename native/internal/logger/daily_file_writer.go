package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

type dailyFileWriter struct {
	mu          sync.Mutex
	cfg         FileConfig
	now         func() time.Time
	currentDate string
	current     *lumberjack.Logger
}

func newDailyFileWriter(cfg FileConfig) *dailyFileWriter {
	return &dailyFileWriter{
		cfg: cfg,
		now: time.Now,
	}
}

func (w *dailyFileWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.rotateIfNeededLocked(); err != nil {
		return 0, err
	}

	return w.current.Write(p)
}

func (w *dailyFileWriter) Sync() error {
	return nil
}

func (w *dailyFileWriter) rotateIfNeededLocked() error {
	currentDate := w.now().Format("2006-01-02")
	if w.current != nil && w.currentDate == currentDate {
		return nil
	}

	if err := os.MkdirAll(w.cfg.Path, 0o755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	if w.current != nil {
		if err := w.current.Close(); err != nil {
			return fmt.Errorf("failed to close previous log file: %w", err)
		}
	}

	w.currentDate = currentDate
	w.current = &lumberjack.Logger{
		Filename:   filepath.Join(w.cfg.Path, fmt.Sprintf("%s-%s.log", w.cfg.Filename, currentDate)),
		MaxSize:    w.cfg.MaxSize,
		MaxBackups: w.cfg.MaxBackups,
		MaxAge:     w.cfg.MaxAge,
		Compress:   w.cfg.Compress,
		LocalTime:  true,
	}

	return nil
}
