package utils

import (
	"encoding/json"
	"fmt"
)

func Marshal[T any](data T) (string, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal: %w", err)
	}
	return string(bytes), nil
}

func MarshalIndent[T any](data T, prefix, indent string) (string, error) {
	bytes, err := json.MarshalIndent(data, prefix, indent)
	if err != nil {
		return "", fmt.Errorf("failed to marshal with indent: %w", err)
	}
	return string(bytes), nil
}

func ToBytes[T any](data T) ([]byte, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal to bytes: %w", err)
	}
	return bytes, nil
}

func Unmarshal[T any](jsonStr string) (T, error) {
	var result T
	if jsonStr == "" {
		return result, fmt.Errorf("cannot unmarshal empty string")
	}
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return result, fmt.Errorf("failed to unmarshal: %w", err)
	}
	return result, nil
}

func FromBytes[T any](data []byte) (T, error) {
	var result T
	if len(data) == 0 {
		return result, fmt.Errorf("cannot unmarshal empty bytes")
	}
	err := json.Unmarshal(data, &result)
	if err != nil {
		return result, fmt.Errorf("failed to unmarshal from bytes: %w", err)
	}
	return result, nil
}
