package utils

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// LocalTime 自定义时间类型，用于统一时间格式化
// 实现了 JSON 序列化/反序列化和数据库扫描/赋值接口
type LocalTime time.Time

const (
	// TimeFormat 标准时间格式
	TimeFormat = "2006-01-02 15:04:05"
	// DateFormat 日期格式
	DateFormat = "2006-01-02"
)

// MarshalJSON 实现 JSON 序列化
func (t LocalTime) MarshalJSON() ([]byte, error) {
	if time.Time(t).IsZero() {
		return []byte("null"), nil
	}
	formatted := fmt.Sprintf("\"%s\"", time.Time(t).Format(TimeFormat))
	return []byte(formatted), nil
}

// UnmarshalJSON 实现 JSON 反序列化
func (t *LocalTime) UnmarshalJSON(data []byte) error {
	if string(data) == "null" || string(data) == `""` {
		return nil
	}

	// 去除引号
	str := string(data)
	if len(str) > 2 && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
	}

	// 尝试多种时间格式
	formats := []string{
		TimeFormat,
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05",
		DateFormat,
	}

	var parsed time.Time
	var err error
	for _, format := range formats {
		parsed, err = time.ParseInLocation(format, str, time.Local)
		if err == nil {
			*t = LocalTime(parsed)
			return nil
		}
	}

	return fmt.Errorf("无法解析时间: %s", str)
}

// Value 实现 driver.Valuer 接口，用于数据库写入
func (t LocalTime) Value() (driver.Value, error) {
	if time.Time(t).IsZero() {
		return nil, nil
	}
	return time.Time(t), nil
}

// Scan 实现 sql.Scanner 接口，用于数据库读取
func (t *LocalTime) Scan(value interface{}) error {
	if value == nil {
		*t = LocalTime(time.Time{})
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		*t = LocalTime(v)
		return nil
	case []byte:
		parsed, err := time.ParseInLocation(TimeFormat, string(v), time.Local)
		if err != nil {
			return err
		}
		*t = LocalTime(parsed)
		return nil
	case string:
		parsed, err := time.ParseInLocation(TimeFormat, v, time.Local)
		if err != nil {
			return err
		}
		*t = LocalTime(parsed)
		return nil
	default:
		return fmt.Errorf("无法将 %T 转换为 LocalTime", value)
	}
}

// String 实现 Stringer 接口
func (t LocalTime) String() string {
	if time.Time(t).IsZero() {
		return ""
	}
	return time.Time(t).Format(TimeFormat)
}

// Time 转换为 time.Time
func (t LocalTime) Time() time.Time {
	return time.Time(t)
}

// IsZero 判断是否为零值
func (t LocalTime) IsZero() bool {
	return time.Time(t).IsZero()
}

// Now 返回当前时间
func Now() LocalTime {
	return LocalTime(time.Now())
}

// ParseLocalTime 解析字符串为 LocalTime
func ParseLocalTime(str string) (LocalTime, error) {
	parsed, err := time.ParseInLocation(TimeFormat, str, time.Local)
	if err != nil {
		return LocalTime{}, err
	}
	return LocalTime(parsed), nil
}
