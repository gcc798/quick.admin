package configx

import (
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

func MustLoadYAML(path string, out any) {
	content, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	if err := yaml.Unmarshal(content, out); err != nil {
		panic(err)
	}
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
