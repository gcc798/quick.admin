// Package main 用户管理工具
// 用于新增用户和重置用户密码
// 使用方法：直接修改 main 函数中的变量，然后运行程序
//
// 示例：
//  1. 新增用户：设置 operation = "create" 并填写用户信息
//  2. 重置密码：设置 operation = "reset" 并填写用户名和新密码
//
// 运行：go run cmd/usermgr/main.go
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/force-c/nai-tizi/internal/config"
	"github.com/force-c/nai-tizi/internal/domain/model"
	"github.com/force-c/nai-tizi/internal/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// ==================== 配置区域：开发者根据需求修改以下变量 ====================

	// 操作类型："create" 表示新增用户，"reset" 表示重置密码
	operation := "create"

	// -------------------- 新增用户配置（当 operation = "create" 时有效） --------------------
	newUser := &model.User{
		UserName:    "admin", // 用户名（必填，唯一）
		NickName:    "管理员",   // 昵称（显示名称）
		Password:    "",      // 密码（留空则使用默认密码 admin123，会被自动加密）
		UserType:    0,       // 用户类型：0系统用户 1微信用户 2APP用户
		Email:       "",      // 邮箱（可选）
		Phonenumber: "",      // 手机号（可选）
		Sex:         0,       // 性别：0男 1女 2未知
		Avatar:      "",      // 头像URL（可选）
		Status:      0,       // 状态：0正常 1停用
		Remark:      "",      // 备注（可选）
		CreateBy:    0,       // 创建人ID
		UpdateBy:    0,       // 更新人ID
	}
	// 新增用户时的默认密码（当 newUser.Password 为空时使用）
	defaultPassword := "admin123"

	// -------------------- 重置密码配置（当 operation = "reset" 时有效） --------------------
	resetUsername := "admin"  // 要重置密码的用户名
	newPassword := "admin123" // 新密码

	// ============================================================================

	// 获取可执行文件所在目录
	execPath, err := os.Executable()
	if err != nil {
		fmt.Printf("获取可执行文件路径失败: %v\n", err)
		os.Exit(1)
	}
	execDir := filepath.Dir(execPath)

	// 加载配置
	cfg, _, err := config.Load(execDir)
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化数据库连接
	db, err := initDB(cfg.Database.DSN)
	if err != nil {
		fmt.Printf("初始化数据库失败: %v\n", err)
		os.Exit(1)
	}

	// 根据操作类型执行相应逻辑
	switch operation {
	case "create":
		if err := createUser(db, newUser, defaultPassword); err != nil {
			fmt.Printf("创建用户失败: %v\n", err)
			os.Exit(1)
		}
	case "reset":
		if err := resetPassword(db, resetUsername, newPassword); err != nil {
			fmt.Printf("重置密码失败: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("未知的操作类型: %s，请使用 \"create\" 或 \"reset\"\n", operation)
		os.Exit(1)
	}
}

// initDB 初始化数据库连接
func initDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取 sql.DB 失败: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(10)

	return db, nil
}

// createUser 创建新用户
func createUser(db *gorm.DB, user *model.User, defaultPassword string) error {
	// 检查用户名是否已存在
	existingUser, err := (&model.User{}).FindByUsername(db, user.UserName)
	if err == nil && existingUser != nil {
		return fmt.Errorf("用户名 \"%s\" 已存在", user.UserName)
	}

	// 检查手机号是否已存在（如果提供了手机号）
	if user.Phonenumber != "" {
		existingUser, err = (&model.User{}).FindByPhonenumber(db, user.Phonenumber)
		if err == nil && existingUser != nil {
			return fmt.Errorf("手机号 \"%s\" 已被使用", user.Phonenumber)
		}
	}

	// 检查邮箱是否已存在（如果提供了邮箱）
	if user.Email != "" {
		existingUser, err = (&model.User{}).FindByEmail(db, user.Email)
		if err == nil && existingUser != nil {
			return fmt.Errorf("邮箱 \"%s\" 已被使用", user.Email)
		}
	}

	// 处理密码
	password := user.Password
	if password == "" {
		password = defaultPassword
	}

	// 加密密码
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}
	user.Password = hashedPassword

	// 创建用户
	if err := user.Create(db, user); err != nil {
		return fmt.Errorf("创建用户失败: %w", err)
	}

	fmt.Printf("✅ 用户创建成功！\n")
	fmt.Printf("   用户ID: %d\n", user.ID)
	fmt.Printf("   用户名: %s\n", user.UserName)
	fmt.Printf("   昵称: %s\n", user.NickName)
	fmt.Printf("   密码: %s\n", password)
	return nil
}

// resetPassword 重置用户密码
func resetPassword(db *gorm.DB, username, newPassword string) error {
	// 查找用户
	user, err := (&model.User{}).FindByUsername(db, username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("用户名 \"%s\" 不存在", username)
		}
		return fmt.Errorf("查询用户失败: %w", err)
	}

	// 加密新密码
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	// 更新密码
	if err := user.UpdatePassword(db, user.ID, hashedPassword); err != nil {
		return fmt.Errorf("更新密码失败: %w", err)
	}

	fmt.Printf("✅ 密码重置成功！\n")
	fmt.Printf("   用户ID: %d\n", user.ID)
	fmt.Printf("   用户名: %s\n", user.UserName)
	fmt.Printf("   新密码: %s\n", newPassword)
	return nil
}

// GetUserByUsername 根据用户名查询用户（便于开发者扩展）
func GetUserByUsername(db *gorm.DB, username string) (*model.User, error) {
	return (&model.User{}).FindByUsername(db, username)
}

// ListUsers 列出所有用户（便于开发者扩展）
func ListUsers(db *gorm.DB, limit int) ([]model.User, error) {
	var users []model.User
	err := db.Limit(limit).Find(&users).Error
	return users, err
}

// DeleteUser 删除用户（便于开发者扩展）
func DeleteUser(ctx context.Context, db *gorm.DB, userId int64) error {
	return (&model.User{}).Delete(db, userId)
}
