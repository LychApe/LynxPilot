package user

import "time"

// User 用户模型
type User struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Username  string    `gorm:"uniqueIndex;size:32;not null" json:"username"`
	Password  string    `gorm:"size:255;not null" json:"-"` // bcrypt 哈希（对前端传来的 MD5 密码做 bcrypt）
	Email     string    `gorm:"size:128" json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
