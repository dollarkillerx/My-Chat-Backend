package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID        string         `gorm:"primaryKey;size:32" json:"id"`
	Username  string         `gorm:"uniqueIndex;size:64;not null" json:"username"`
	Nickname  string         `gorm:"size:64" json:"nickname"`
	Avatar    string         `gorm:"size:256" json:"avatar"`
	Password  string         `gorm:"size:128;not null" json:"-"`
	Phone     string         `gorm:"index;size:20" json:"phone"`
	Email     string         `gorm:"index;size:128" json:"email"`
	Status    int            `gorm:"default:1" json:"status"` // 1=正常, 0=禁用
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (User) TableName() string {
	return "users"
}

// UserStatus 用户状态
const (
	UserStatusDisabled = 0
	UserStatusNormal   = 1
)
