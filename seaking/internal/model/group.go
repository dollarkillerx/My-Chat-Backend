package model

import (
	"time"

	"gorm.io/gorm"
)

// Group 群组
type Group struct {
	ID          string         `gorm:"primaryKey;size:32" json:"id"`
	Name        string         `gorm:"size:64;not null" json:"name"`
	Avatar      string         `gorm:"size:256" json:"avatar"`
	Description string         `gorm:"size:512" json:"description"`
	OwnerID     string         `gorm:"index;size:32;not null" json:"owner_id"`
	MaxMembers  int            `gorm:"default:500" json:"max_members"`
	Status      int            `gorm:"default:1" json:"status"` // 1=正常, 0=解散
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (Group) TableName() string {
	return "groups"
}

// GroupMember 群成员
type GroupMember struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	GroupID   string         `gorm:"index;size:32;not null" json:"group_id"`
	UserID    string         `gorm:"index;size:32;not null" json:"user_id"`
	Role      int            `gorm:"default:0" json:"role"`     // 0=普通成员, 1=管理员, 2=群主
	Nickname  string         `gorm:"size:64" json:"nickname"`   // 群昵称
	Muted     bool           `gorm:"default:false" json:"muted"` // 是否被禁言
	MutedAt   *time.Time     `json:"muted_at"`
	JoinedAt  time.Time      `json:"joined_at"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (GroupMember) TableName() string {
	return "group_members"
}

// 群组状态
const (
	GroupStatusDissolved = 0
	GroupStatusNormal    = 1
)

// 群成员角色
const (
	GroupRoleMember = 0
	GroupRoleAdmin  = 1
	GroupRoleOwner  = 2
)
