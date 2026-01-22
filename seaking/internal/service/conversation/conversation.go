package conversation

import (
	"context"
	"strings"
	"time"

	"github.com/my-chat/common/pkg/errors"
	"github.com/my-chat/seaking/internal/model"
	"github.com/my-chat/seaking/internal/storage"
	"github.com/rs/xid"
	"gorm.io/gorm"
)

// Service 会话服务
type Service struct {
	storage *storage.Storage
}

// NewService 创建会话服务
func NewService(storage *storage.Storage) *Service {
	return &Service{storage: storage}
}

// CreateDirectConversation 创建单聊会话
func (s *Service) CreateDirectConversation(ctx context.Context, uid1, uid2 string) (*model.Conversation, error) {
	cid := model.GenerateDirectCid(uid1, uid2)

	// 检查是否已存在
	var existing model.Conversation
	if err := s.storage.DB().Where("id = ?", cid).First(&existing).Error; err == nil {
		return &existing, nil
	}

	conv := &model.Conversation{
		ID:   cid,
		Type: model.ConversationTypeDirect,
	}

	err := s.storage.DB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(conv).Error; err != nil {
			return err
		}

		now := time.Now()
		members := []model.ConversationMember{
			{ConversationID: cid, UserID: uid1, JoinedAt: now},
			{ConversationID: cid, UserID: uid2, JoinedAt: now},
		}
		return tx.Create(&members).Error
	})

	if err != nil {
		return nil, err
	}

	return conv, nil
}

// CreateGroupConversation 创建群聊会话
func (s *Service) CreateGroupConversation(ctx context.Context, groupId string, memberIds []string) (*model.Conversation, error) {
	cid := model.GenerateGroupCid(groupId)

	// 获取群组信息
	var group model.Group
	if err := s.storage.DB().First(&group, "id = ?", groupId).Error; err != nil {
		return nil, errors.ErrGroupNotFound
	}

	conv := &model.Conversation{
		ID:     cid,
		Type:   model.ConversationTypeGroup,
		Name:   group.Name,
		Avatar: group.Avatar,
	}

	err := s.storage.DB().Transaction(func(tx *gorm.DB) error {
		// 使用 FirstOrCreate 避免重复创建
		if err := tx.FirstOrCreate(conv, model.Conversation{ID: cid}).Error; err != nil {
			return err
		}

		// 添加成员
		now := time.Now()
		for _, uid := range memberIds {
			member := model.ConversationMember{
				ConversationID: cid,
				UserID:         uid,
				JoinedAt:       now,
			}
			// 使用 FirstOrCreate 避免重复添加
			if err := tx.FirstOrCreate(&member, model.ConversationMember{
				ConversationID: cid,
				UserID:         uid,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return conv, nil
}

// GetConversation 获取会话信息
func (s *Service) GetConversation(ctx context.Context, cid string) (*model.Conversation, error) {
	var conv model.Conversation
	if err := s.storage.DB().First(&conv, "id = ?", cid).Error; err != nil {
		return nil, errors.ErrConversationNotFound
	}
	return &conv, nil
}

// GetConversationMembers 获取会话成员
func (s *Service) GetConversationMembers(ctx context.Context, cid string) ([]model.ConversationMember, error) {
	var members []model.ConversationMember
	err := s.storage.DB().Where("conversation_id = ?", cid).Find(&members).Error
	return members, err
}

// GetConversationMemberIds 获取会话成员ID列表
func (s *Service) GetConversationMemberIds(ctx context.Context, cid string) ([]string, error) {
	var members []model.ConversationMember
	err := s.storage.DB().Select("user_id").Where("conversation_id = ?", cid).Find(&members).Error
	if err != nil {
		return nil, err
	}

	ids := make([]string, len(members))
	for i, m := range members {
		ids[i] = m.UserID
	}
	return ids, nil
}

// GetUserConversations 获取用户的所有会话
func (s *Service) GetUserConversations(ctx context.Context, uid string) ([]model.Conversation, error) {
	var convs []model.Conversation
	err := s.storage.DB().
		Joins("JOIN conversation_members ON conversation_members.conversation_id = conversations.id").
		Where("conversation_members.user_id = ?", uid).
		Find(&convs).Error
	return convs, err
}

// CheckAccess 检查用户是否有权访问会话
func (s *Service) CheckAccess(ctx context.Context, uid, cid string) (bool, int, bool, error) {
	var member model.ConversationMember
	err := s.storage.DB().Where("conversation_id = ? AND user_id = ?", cid, uid).First(&member).Error
	if err != nil {
		return false, 0, false, nil
	}

	// 获取角色信息（如果是群聊）
	role := 0
	muted := false

	if strings.HasPrefix(cid, "g:") {
		// 群聊，获取群成员角色
		groupId := strings.TrimPrefix(cid, "g:")
		var groupMember model.GroupMember
		if err := s.storage.DB().Where("group_id = ? AND user_id = ?", groupId, uid).First(&groupMember).Error; err == nil {
			role = groupMember.Role
			muted = groupMember.Muted
		}
	}

	return true, role, muted, nil
}

// AddMember 添加会话成员
func (s *Service) AddMember(ctx context.Context, cid, uid string) error {
	member := &model.ConversationMember{
		ConversationID: cid,
		UserID:         uid,
		JoinedAt:       time.Now(),
	}
	return s.storage.DB().Create(member).Error
}

// RemoveMember 移除会话成员
func (s *Service) RemoveMember(ctx context.Context, cid, uid string) error {
	return s.storage.DB().Where("conversation_id = ? AND user_id = ?", cid, uid).
		Delete(&model.ConversationMember{}).Error
}

// UpdateLastReadMid 更新最后已读消息ID
func (s *Service) UpdateLastReadMid(ctx context.Context, cid, uid string, lastReadMid int64) error {
	return s.storage.DB().Model(&model.ConversationMember{}).
		Where("conversation_id = ? AND user_id = ?", cid, uid).
		Update("last_read_mid", lastReadMid).Error
}

// SetMuted 设置免打扰
func (s *Service) SetMuted(ctx context.Context, cid, uid string, muted bool) error {
	return s.storage.DB().Model(&model.ConversationMember{}).
		Where("conversation_id = ? AND user_id = ?", cid, uid).
		Update("muted", muted).Error
}

// SetPinned 设置置顶
func (s *Service) SetPinned(ctx context.Context, cid, uid string, pinned bool) error {
	return s.storage.DB().Model(&model.ConversationMember{}).
		Where("conversation_id = ? AND user_id = ?", cid, uid).
		Update("pinned", pinned).Error
}

// CreateConversationRequest 创建会话请求
type CreateConversationRequest struct {
	Type      int      `json:"type"`
	CreatorId string   `json:"creator_id"`
	MemberIds []string `json:"member_ids"`
	Name      string   `json:"name,omitempty"`
}

// CreateConversation 通用创建会话方法
func (s *Service) CreateConversation(ctx context.Context, req *CreateConversationRequest) (*model.Conversation, error) {
	if req.Type == model.ConversationTypeDirect {
		if len(req.MemberIds) != 2 {
			return nil, errors.ErrInvalidParam
		}
		return s.CreateDirectConversation(ctx, req.MemberIds[0], req.MemberIds[1])
	}

	// 群聊需要先创建群组或关联已有群组
	// 这里简化处理，生成一个临时群组ID
	groupId := xid.New().String()
	return s.CreateGroupConversation(ctx, groupId, req.MemberIds)
}
