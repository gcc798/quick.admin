package configx

import (
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	_ "github.com/go-kratos/kratos/v2/encoding/yaml"
)

func MustLoad(path string, out any) {
	if err := Load(path, out); err != nil {
		panic(err)
	}
}

func Load(path string, out any) error {
	reader := config.New(config.WithSource(file.NewSource(path)))
	defer reader.Close()
	if err := reader.Load(); err != nil {
		return err
	}
	if err := reader.Scan(out); err != nil {
		return err
	}
	return nil
}

func ParseDurationOrDefault(value string, defaultValue time.Duration) time.Duration {
	value = strings.TrimSpace(value)
	if value == "" {
		return defaultValue
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}
	return duration
}
