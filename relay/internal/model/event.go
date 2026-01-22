package model

import (
	"time"

	"gorm.io/gorm"
)

// Event 消息事件存储模型
type Event struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Mid       int64          `gorm:"uniqueIndex;not null" json:"mid"`              // 消息ID
	Cid       string         `gorm:"index;size:64;not null" json:"cid"`            // 会话ID
	Kind      int            `gorm:"index;not null" json:"kind"`                   // 消息类型
	Sender    string         `gorm:"index;size:32;not null" json:"sender"`         // 发送者UID
	Tags      string         `gorm:"type:jsonb" json:"tags"`                       // 标签JSON
	Data      string         `gorm:"type:jsonb" json:"data"`                       // 消息体JSON
	Flags     int            `gorm:"default:0" json:"flags"`                       // 标志位
	Sig       string         `gorm:"size:256" json:"sig"`                          // 签名
	Timestamp int64          `gorm:"index;not null" json:"timestamp"`              // 时间戳
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (Event) TableName() string {
	return "events"
}

// ReadReceipt 已读回执存储模型
type ReadReceipt struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Cid         string    `gorm:"index;size:64;not null" json:"cid"`     // 会话ID
	Uid         string    `gorm:"index;size:32;not null" json:"uid"`     // 用户ID
	LastReadMid int64     `gorm:"not null" json:"last_read_mid"`         // 最后已读消息ID
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 表名
func (ReadReceipt) TableName() string {
	return "read_receipts"
}

// Reaction 消息反应存储模型
type Reaction struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Mid       int64          `gorm:"index;not null" json:"mid"`           // 目标消息ID
	Cid       string         `gorm:"index;size:64;not null" json:"cid"`   // 会话ID
	Uid       string         `gorm:"index;size:32;not null" json:"uid"`   // 用户ID
	Emoji     string         `gorm:"size:32;not null" json:"emoji"`       // 表情
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (Reaction) TableName() string {
	return "reactions"
}
