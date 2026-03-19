package data

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/structpb"
)

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

func buildPage(pageNum, pageSize int64) (int64, int64) {
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return pageNum, pageSize
}

func pageBounds(total int, pageNum, pageSize int64) (int, int) {
	start := int((pageNum - 1) * pageSize)
	if start > total {
		start = total
	}
	end := start + int(pageSize)
	if end > total {
		end = total
	}
	return start, end
}

func stringifyJSON(data map[string]any) string {
	if len(data) == 0 {
		return ""
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return string(raw)
}

func rawJSONToProtoValue(data []byte) *structpb.Value {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return nil
	}
	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		return structpb.NewStringValue(string(data))
	}
	protoValue, err := structpb.NewValue(value)
	if err != nil {
		return nil
	}
	return protoValue
}

func protoValueToRawJSON(value *structpb.Value) []byte {
	if value == nil {
		return []byte("null")
	}
	raw, err := json.Marshal(value.AsInterface())
	if err != nil {
		return []byte("null")
	}
	return raw
}

func protoStructToMap(value *structpb.Struct) map[string]any {
	if value == nil {
		return map[string]any{}
	}
	return value.AsMap()
}

func mapToProtoStruct(data map[string]any) *structpb.Struct {
	if len(data) == 0 {
		return nil
	}
	value, err := structpb.NewStruct(data)
	if err != nil {
		return nil
	}
	return value
}

func parseStorageConfig(value string) map[string]any {
	value = strings.TrimSpace(value)
	if value == "" {
		return map[string]any{}
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(value), &out); err != nil {
		return map[string]any{"raw": value}
	}
	return out
}

func statusString(v int64) string {
	return strconv.FormatInt(v, 10)
}

func storageConfigString(config map[string]any, keys []string, defaultValue string) string {
	for _, key := range keys {
		value, ok := config[key]
		if !ok {
			continue
		}
		switch typed := value.(type) {
		case string:
			if strings.TrimSpace(typed) != "" {
				return strings.TrimSpace(typed)
			}
		case json.Number:
			return typed.String()
		}
	}
	return defaultValue
}

func storageConfigBool(config map[string]any, keys []string, defaultValue bool) bool {
	for _, key := range keys {
		value, ok := config[key]
		if !ok {
			continue
		}
		switch typed := value.(type) {
		case bool:
			return typed
		case string:
			typed = strings.TrimSpace(strings.ToLower(typed))
			if typed == "true" || typed == "1" {
				return true
			}
			if typed == "false" || typed == "0" {
				return false
			}
		}
	}
	return defaultValue
}

func parseRawMetadata(value string) map[string]any {
	value = strings.TrimSpace(value)
	if value == "" {
		return map[string]any{}
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(value), &out); err != nil {
		return map[string]any{"raw": value}
	}
	return out
}

func protoStructToJSONString(value *structpb.Struct) string {
	if value == nil {
		return ""
	}
	raw, err := json.Marshal(value.AsMap())
	if err != nil {
		return ""
	}
	return string(raw)
}

func parseQueryTime(value string) *time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02"} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return &parsed
		}
	}
	return nil
}

func inTimeRange(target time.Time, start, end *time.Time) bool {
	if target.IsZero() {
		return start == nil && end == nil
	}
	if start != nil && target.Before(*start) {
		return false
	}
	if end != nil && target.After(*end) {
		return false
	}
	return true
}
