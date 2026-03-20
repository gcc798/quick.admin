package utils

import (
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// TestNewBcryptHasher 测试创建 bcrypt 哈希器
func TestNewBcryptHasher(t *testing.T) {
	tests := []struct {
		name         string
		cost         int
		expectedCost int
	}{
		{
			name:         "使用默认成本",
			cost:         bcrypt.DefaultCost,
			expectedCost: bcrypt.DefaultCost,
		},
		{
			name:         "使用最小成本",
			cost:         bcrypt.MinCost,
			expectedCost: bcrypt.MinCost,
		},
		{
			name:         "使用最大成本",
			cost:         bcrypt.MaxCost,
			expectedCost: bcrypt.MaxCost,
		},
		{
			name:         "成本过低，使用默认值",
			cost:         bcrypt.MinCost - 1,
			expectedCost: bcrypt.DefaultCost,
		},
		{
			name:         "成本过高，使用默认值",
			cost:         bcrypt.MaxCost + 1,
			expectedCost: bcrypt.DefaultCost,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasher := NewBcryptHasher(tt.cost)
			if hasher.cost != tt.expectedCost {
				t.Errorf("expected cost %d, got %d", tt.expectedCost, hasher.cost)
			}
		})
	}
}

// TestNewDefaultBcryptHasher 测试创建默认哈希器
func TestNewDefaultBcryptHasher(t *testing.T) {
	hasher := NewDefaultBcryptHasher()
	if hasher.cost != bcrypt.DefaultCost {
		t.Errorf("expected default cost %d, got %d", bcrypt.DefaultCost, hasher.cost)
	}
}

// TestHashPassword 测试密码哈希生成
func TestHashPassword(t *testing.T) {
	hasher := NewDefaultBcryptHasher()

	tests := []struct {
		name        string
		password    string
		expectError bool
		errorType   error
	}{
		{
			name:        "正常密码",
			password:    "admin123",
			expectError: false,
		},
		{
			name:        "短密码",
			password:    "123",
			expectError: false,
		},
		{
			name:        "长密码",
			password:    strings.Repeat("a", 72), // bcrypt 最大支持 72 字节
			expectError: false,
		},
		{
			name:        "包含特殊字符",
			password:    "P@ssw0rd!#$%",
			expectError: false,
		},
		{
			name:        "包含中文",
			password:    "密码123",
			expectError: false,
		},
		{
			name:        "空密码",
			password:    "",
			expectError: true,
			errorType:   ErrPasswordEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := hasher.HashPassword(tt.password)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if tt.errorType != nil && err != tt.errorType {
					t.Errorf("expected error %v, got %v", tt.errorType, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// 验证哈希值长度（bcrypt 固定 60 字符）
			if len(hash) != 60 {
				t.Errorf("expected hash length 60, got %d", len(hash))
			}

			// 验证哈希值格式（应该以 $2a$ 或 $2b$ 开头）
			if !strings.HasPrefix(hash, "$2a$") && !strings.HasPrefix(hash, "$2b$") {
				t.Errorf("invalid hash format: %s", hash)
			}

			// 验证生成的哈希可以验证原密码
			if err := hasher.VerifyPassword(hash, tt.password); err != nil {
				t.Errorf("failed to verify generated hash: %v", err)
			}
		})
	}
}

// TestHashPasswordDeterminism 测试密码哈希的非确定性
// bcrypt 每次生成的哈希都应该不同（因为包含随机盐值）
func TestHashPasswordDeterminism(t *testing.T) {
	hasher := NewDefaultBcryptHasher()
	password := "admin123"

	hash1, err := hasher.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	hash2, err := hasher.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	// 两次哈希应该不同
	if hash1 == hash2 {
		t.Error("expected different hashes for same password, got identical")
	}

	// 但都应该能验证原密码
	if err := hasher.VerifyPassword(hash1, password); err != nil {
		t.Errorf("hash1 failed to verify: %v", err)
	}
	if err := hasher.VerifyPassword(hash2, password); err != nil {
		t.Errorf("hash2 failed to verify: %v", err)
	}
}

// TestVerifyPassword 测试密码验证
func TestVerifyPassword(t *testing.T) {
	hasher := NewDefaultBcryptHasher()

	// 预先生成一些测试哈希
	validPassword := "admin123"
	validHash, _ := hasher.HashPassword(validPassword)

	tests := []struct {
		name           string
		hashedPassword string
		password       string
		expectError    bool
		errorType      error
	}{
		{
			name:           "正确的密码",
			hashedPassword: validHash,
			password:       validPassword,
			expectError:    false,
		},
		{
			name:           "错误的密码",
			hashedPassword: validHash,
			password:       "wrongpassword",
			expectError:    true,
			errorType:      ErrPasswordMismatch,
		},
		{
			name:           "空哈希值",
			hashedPassword: "",
			password:       "admin123",
			expectError:    true,
			errorType:      ErrHashEmpty,
		},
		{
			name:           "空密码",
			hashedPassword: validHash,
			password:       "",
			expectError:    true,
			errorType:      ErrPasswordEmpty,
		},
		{
			name:           "无效的哈希格式",
			hashedPassword: "invalid_hash",
			password:       "admin123",
			expectError:    true,
		},
		{
			name:           "已知的正确哈希（admin123）",
			hashedPassword: "$2a$10$Q55.ONb4ACprCH5Wl9NqouI9uWyvV.wGT4BSRRnCWQXdfJiWgOHzK",
			password:       "admin123",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := hasher.VerifyPassword(tt.hashedPassword, tt.password)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if tt.errorType != nil && err != tt.errorType {
					t.Errorf("expected error %v, got %v", tt.errorType, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestVerifyPasswordCaseSensitive 测试密码验证的大小写敏感性
func TestVerifyPasswordCaseSensitive(t *testing.T) {
	hasher := NewDefaultBcryptHasher()
	password := "Admin123"
	hash, _ := hasher.HashPassword(password)

	tests := []struct {
		name        string
		password    string
		expectError bool
	}{
		{
			name:        "完全匹配",
			password:    "Admin123",
			expectError: false,
		},
		{
			name:        "全小写",
			password:    "admin123",
			expectError: true,
		},
		{
			name:        "全大写",
			password:    "ADMIN123",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := hasher.VerifyPassword(hash, tt.password)
			if tt.expectError && err == nil {
				t.Error("expected error for case mismatch, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestMustHashPassword 测试 MustHashPassword
func TestMustHashPassword(t *testing.T) {
	hasher := NewDefaultBcryptHasher()

	t.Run("正常密码", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("unexpected panic: %v", r)
			}
		}()

		hash := hasher.MustHashPassword("admin123")
		if len(hash) != 60 {
			t.Errorf("expected hash length 60, got %d", len(hash))
		}
		t.Log(hash)
	})

	t.Run("空密码应该 panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for empty password, got none")
			}
		}()

		hasher.MustHashPassword("")
	})
}

// TestGlobalFunctions 测试全局函数
func TestGlobalFunctions(t *testing.T) {
	password := "testpassword"

	t.Run("HashPassword", func(t *testing.T) {
		hash, err := HashPassword(password)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(hash) != 60 {
			t.Errorf("expected hash length 60, got %d", len(hash))
		}
	})

	t.Run("VerifyPassword", func(t *testing.T) {
		hash, _ := HashPassword(password)
		err := VerifyPassword(hash, password)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("MustHashPassword", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("unexpected panic: %v", r)
			}
		}()

		hash := MustHashPassword(password)
		if len(hash) != 60 {
			t.Errorf("expected hash length 60, got %d", len(hash))
		}
	})
}

// TestDifferentCosts 测试不同的成本因子
func TestDifferentCosts(t *testing.T) {
	password := "admin123"

	costs := []int{bcrypt.MinCost, bcrypt.DefaultCost, 12}

	for _, cost := range costs {
		t.Run(string(rune(cost)), func(t *testing.T) {
			hasher := NewBcryptHasher(cost)
			hash, err := hasher.HashPassword(password)
			if err != nil {
				t.Errorf("failed to hash with cost %d: %v", cost, err)
				return
			}

			// 验证密码
			if err := hasher.VerifyPassword(hash, password); err != nil {
				t.Errorf("failed to verify with cost %d: %v", cost, err)
			}
		})
	}
}

// BenchmarkHashPassword 基准测试：密码哈希生成
func BenchmarkHashPassword(b *testing.B) {
	hasher := NewDefaultBcryptHasher()
	password := "admin123"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = hasher.HashPassword(password)
	}
}

// BenchmarkVerifyPassword 基准测试：密码验证
func BenchmarkVerifyPassword(b *testing.B) {
	hasher := NewDefaultBcryptHasher()
	password := "admin123"
	hash, _ := hasher.HashPassword(password)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = hasher.VerifyPassword(hash, password)
	}
}

// BenchmarkHashPasswordDifferentCosts 基准测试：不同成本因子的性能
func BenchmarkHashPasswordDifferentCosts(b *testing.B) {
	password := "admin123"
	costs := []int{bcrypt.MinCost, bcrypt.DefaultCost, 12, 14}

	for _, cost := range costs {
		b.Run(string(rune(cost)), func(b *testing.B) {
			hasher := NewBcryptHasher(cost)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = hasher.HashPassword(password)
			}
		})
	}
}

// TestRealWorldScenario 测试真实场景
func TestRealWorldScenario(t *testing.T) {
	hasher := NewDefaultBcryptHasher()

	// 场景1：用户注册
	t.Run("用户注册", func(t *testing.T) {
		username := "testuser"
		password := "MySecureP@ssw0rd"

		// 生成密码哈希存储到数据库
		hashedPassword, err := hasher.HashPassword(password)
		if err != nil {
			t.Fatalf("failed to hash password: %v", err)
		}

		t.Logf("User %s registered with hash: %s", username, hashedPassword)

		// 验证哈希格式
		if len(hashedPassword) != 60 {
			t.Errorf("invalid hash length: %d", len(hashedPassword))
		}
	})

	// 场景2：用户登录
	t.Run("用户登录", func(t *testing.T) {
		// 模拟从数据库读取的哈希值
		storedHash := "$2a$10$Q55.ONb4ACprCH5Wl9NqouI9uWyvV.wGT4BSRRnCWQXdfJiWgOHzK"
		inputPassword := "admin123"

		// 验证密码
		err := hasher.VerifyPassword(storedHash, inputPassword)
		if err != nil {
			t.Errorf("login failed: %v", err)
		} else {
			t.Log("Login successful")
		}
	})

	// 场景3：密码错误
	t.Run("密码错误", func(t *testing.T) {
		storedHash := "$2a$10$Q55.ONb4ACprCH5Wl9NqouI9uWyvV.wGT4BSRRnCWQXdfJiWgOHzK"
		wrongPassword := "wrongpassword"

		err := hasher.VerifyPassword(storedHash, wrongPassword)
		if err == nil {
			t.Error("expected error for wrong password, got nil")
		}
		if err != ErrPasswordMismatch {
			t.Errorf("expected ErrPasswordMismatch, got %v", err)
		}
		t.Logf("Login failed as expected: %v", err)
	})

	// 场景4：修改密码
	t.Run("修改密码", func(t *testing.T) {
		oldPassword := "OldPassword123"
		newPassword := "NewPassword456"

		// 生成旧密码哈希
		oldHash, _ := hasher.HashPassword(oldPassword)

		// 用户输入旧密码验证
		if err := hasher.VerifyPassword(oldHash, oldPassword); err != nil {
			t.Errorf("old password verification failed: %v", err)
			return
		}

		// 生成新密码哈希
		newHash, err := hasher.HashPassword(newPassword)
		if err != nil {
			t.Errorf("failed to hash new password: %v", err)
			return
		}

		// 验证新密码
		if err := hasher.VerifyPassword(newHash, newPassword); err != nil {
			t.Errorf("new password verification failed: %v", err)
		}

		// 确保旧密码不能用新哈希验证
		if err := hasher.VerifyPassword(newHash, oldPassword); err == nil {
			t.Error("old password should not work with new hash")
		}

		t.Log("Password changed successfully")
	})
}

func Test2(t *testing.T) {
	password, err := HashPassword("admin123")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(password)
}
