package userService

import (
	"errors"

	userModel "github.com/LychApe/LynxPilot/internal/model/user"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Login 根据用户名查询用户，并用 bcrypt 校验密码（前端传来的是 MD5 值）
func Login(db *gorm.DB, username, password string) (*userModel.User, error) {
	var user userModel.User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, err
	}

	return &user, nil
}

// CreateUser 创建用户，password 为前端传来的 MD5 值，内部做 bcrypt 加密。
func CreateUser(db *gorm.DB, username, password, email string) (*userModel.User, error) {
	var existing userModel.User
	if err := db.First(&existing).Error; err == nil {
		return nil, errors.New("用户已存在，单用户系统不允许创建多个用户")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &userModel.User{
		Username: username,
		Password: string(hash),
		Email:    email,
	}

	if err := db.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// IsInstalled 检查是否已完成安装向导（数据库中是否存在用户）
func IsInstalled(db *gorm.DB) bool {
	var count int64
	db.Model(&userModel.User{}).Count(&count)
	return count > 0
}

// UpdateUser 编辑用户信息，仅更新非空字段
func UpdateUser(db *gorm.DB, userID uint, username, password, email string) (*userModel.User, error) {
	var user userModel.User
	if err := db.First(&user, userID).Error; err != nil {
		return nil, err
	}

	updates := map[string]any{}
	if username != "" {
		updates["username"] = username
	}
	if password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		updates["password"] = string(hash)
	}
	if email != "" {
		updates["email"] = email
	}

	if len(updates) > 0 {
		if err := db.Model(&user).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	return &user, nil
}
