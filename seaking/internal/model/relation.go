package model

import (
	"time"

	"gorm.io/gorm"
)

// Friendship 好友关系
type Friendship struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    string         `gorm:"index;size:32;not null" json:"user_id"`
	FriendID  string         `gorm:"index;size:32;not null" json:"friend_id"`
	Remark    string         `gorm:"size:64" json:"remark"`          // 好友备注
	Status    int            `gorm:"default:1" json:"status"`        // 1=正常, 2=拉黑
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (Friendship) TableName() string {
	return "friendships"
}

// FriendRequest 好友请求
type FriendRequest struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	FromUID   string         `gorm:"index;size:32;not null" json:"from_uid"`
	ToUID     string         `gorm:"index;size:32;not null" json:"to_uid"`
	Message   string         `gorm:"size:256" json:"message"`
	Status    int            `gorm:"default:0" json:"status"` // 0=待处理, 1=同意, 2=拒绝
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (FriendRequest) TableName() string {
	return "friend_requests"
}

// 好友关系状态
const (
	FriendStatusNormal  = 1
	FriendStatusBlocked = 2
)

// 好友请求状态
const (
	FriendRequestPending  = 0
	FriendRequestAccepted = 1
	FriendRequestRejected = 2
)
