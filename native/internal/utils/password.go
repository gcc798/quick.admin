package utils

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrPasswordEmpty 密码为空错误
	ErrPasswordEmpty = errors.New("password cannot be empty")
	// ErrHashEmpty 哈希值为空错误
	ErrHashEmpty = errors.New("hash cannot be empty")
	// ErrPasswordMismatch 密码不匹配错误
	ErrPasswordMismatch = errors.New("password mismatch")
)

// PasswordHasher 密码哈希器接口
type PasswordHasher interface {
	// HashPassword 生成密码哈希
	HashPassword(password string) (string, error)
	// VerifyPassword 验证密码
	VerifyPassword(hashedPassword, password string) error
}

// BcryptHasher bcrypt 密码哈希器
type BcryptHasher struct {
	cost int
}

// NewBcryptHasher 创建 bcrypt 密码哈希器
// cost: bcrypt 成本因子，默认为 bcrypt.DefaultCost (10)
// 成本越高，计算越慢，安全性越高，建议范围 10-14
func NewBcryptHasher(cost int) *BcryptHasher {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}
	return &BcryptHasher{cost: cost}
}

// NewDefaultBcryptHasher 创建默认的 bcrypt 密码哈希器
func NewDefaultBcryptHasher() *BcryptHasher {
	return NewBcryptHasher(bcrypt.DefaultCost)
}

// HashPassword 生成密码哈希
// 参数:
//   - password: 明文密码
//
// 返回:
//   - string: bcrypt 哈希值（60个字符）
//   - error: 错误信息
func (h *BcryptHasher) HashPassword(password string) (string, error) {
	if password == "" {
		return "", ErrPasswordEmpty
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hashedBytes), nil
}

// VerifyPassword 验证密码
// 参数:
//   - hashedPassword: bcrypt 哈希值
//   - password: 明文密码
//
// 返回:
//   - error: 如果密码匹配返回 nil，否则返回错误
func (h *BcryptHasher) VerifyPassword(hashedPassword, password string) error {
	if hashedPassword == "" {
		return ErrHashEmpty
	}
	if password == "" {
		return ErrPasswordEmpty
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrPasswordMismatch
		}
		return fmt.Errorf("failed to verify password: %w", err)
	}

	return nil
}

// MustHashPassword 生成密码哈希，如果失败则 panic
// 仅用于测试或初始化场景
func (h *BcryptHasher) MustHashPassword(password string) string {
	hash, err := h.HashPassword(password)
	if err != nil {
		panic(fmt.Sprintf("failed to hash password: %v", err))
	}
	return hash
}

// 全局默认哈希器实例
var defaultHasher = NewDefaultBcryptHasher()

// HashPassword 使用默认哈希器生成密码哈希
func HashPassword(password string) (string, error) {
	return defaultHasher.HashPassword(password)
}

// VerifyPassword 使用默认哈希器验证密码
func VerifyPassword(hashedPassword, password string) error {
	return defaultHasher.VerifyPassword(hashedPassword, password)
}

// MustHashPassword 使用默认哈希器生成密码哈希，失败则 panic
func MustHashPassword(password string) string {
	return defaultHasher.MustHashPassword(password)
}
