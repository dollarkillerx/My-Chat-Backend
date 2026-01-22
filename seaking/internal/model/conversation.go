package model

import (
	"time"

	"gorm.io/gorm"
)

// Conversation 会话
type Conversation struct {
	ID        string         `gorm:"primaryKey;size:64" json:"id"` // 会话ID (cid)
	Type      int            `gorm:"not null" json:"type"`         // 1=单聊, 2=群聊
	Name      string         `gorm:"size:64" json:"name"`          // 会话名称（群聊时为群名）
	Avatar    string         `gorm:"size:256" json:"avatar"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (Conversation) TableName() string {
	return "conversations"
}

// ConversationMember 会话成员
type ConversationMember struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	ConversationID string         `gorm:"index;size:64;not null" json:"conversation_id"`
	UserID         string         `gorm:"index;size:32;not null" json:"user_id"`
	LastReadMid    int64          `gorm:"default:0" json:"last_read_mid"` // 最后已读消息ID
	Muted          bool           `gorm:"default:false" json:"muted"`     // 是否免打扰
	Pinned         bool           `gorm:"default:false" json:"pinned"`    // 是否置顶
	JoinedAt       time.Time      `json:"joined_at"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (ConversationMember) TableName() string {
	return "conversation_members"
}

// 会话类型
const (
	ConversationTypeDirect = 1 // 单聊
	ConversationTypeGroup  = 2 // 群聊
)

// GenerateDirectCid 生成单聊会话ID
func GenerateDirectCid(uid1, uid2 string) string {
	if uid1 < uid2 {
		return "d:" + uid1 + ":" + uid2
	}
	return "d:" + uid2 + ":" + uid1
}

// GenerateGroupCid 生成群聊会话ID
func GenerateGroupCid(groupId string) string {
	return "g:" + groupId
}
