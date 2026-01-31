package model

import (
	"time"
)

// UserKey 用户密钥表
type UserKey struct {
	ID                  uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID              string    `gorm:"uniqueIndex;size:32;not null" json:"user_id"`
	PublicKey           string    `gorm:"type:text;not null" json:"public_key"`                // 公钥 (明文, Base64 PEM)
	EncryptedPrivateKey string    `gorm:"type:text;not null" json:"encrypted_private_key"`     // 私钥 (密码加密, Base64)
	KeySalt             string    `gorm:"size:64;not null" json:"key_salt"`                    // 密钥派生盐值
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// TableName 表名
func (UserKey) TableName() string {
	return "user_keys"
}

// ChatKey 私聊加密密钥表
type ChatKey struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ConversationID string    `gorm:"index;size:64;not null" json:"conversation_id"` // 会话ID (d:uid1:uid2)
	UserID         string    `gorm:"size:32;not null" json:"user_id"`               // 用户ID
	EncryptedKey   string    `gorm:"type:text;not null" json:"encrypted_key"`       // 对称密钥 (用该用户公钥加密)
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// TableName 表名
func (ChatKey) TableName() string {
	return "chat_keys"
}

// GroupKey 群组加密密钥表
type GroupKey struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	GroupID      string    `gorm:"index;size:32;not null" json:"group_id"`    // 群组ID
	UserID       string    `gorm:"size:32;not null" json:"user_id"`           // 用户ID
	EncryptedKey string    `gorm:"type:text;not null" json:"encrypted_key"`   // 群密钥 (用该用户公钥加密)
	Version      int       `gorm:"default:1" json:"version"`                  // 密钥版本
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName 表名
func (GroupKey) TableName() string {
	return "group_keys"
}

// UserKeyInfo 用户密钥信息 (用于返回给客户端)
type UserKeyInfo struct {
	UserID    string `json:"user_id"`
	PublicKey string `json:"public_key"`
}

// ChatKeyInfo 私聊密钥信息
type ChatKeyInfo struct {
	ConversationID string `json:"conversation_id"`
	EncryptedKey   string `json:"encrypted_key"`
}

// GroupKeyInfo 群组密钥信息
type GroupKeyInfo struct {
	GroupID      string `json:"group_id"`
	EncryptedKey string `json:"encrypted_key"`
	Version      int    `json:"version"`
}
