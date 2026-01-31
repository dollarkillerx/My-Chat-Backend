package key

import (
	"context"

	"github.com/my-chat/common/pkg/errors"
	"github.com/my-chat/seaking/internal/model"
	"github.com/my-chat/seaking/internal/storage"
	"gorm.io/gorm"
)

// Service 密钥服务
type Service struct {
	storage *storage.Storage
}

// NewService 创建密钥服务
func NewService(storage *storage.Storage) *Service {
	return &Service{storage: storage}
}

// ===== 用户密钥 =====

// CreateUserKeyRequest 创建用户密钥请求
type CreateUserKeyRequest struct {
	UserID              string `json:"user_id"`
	PublicKey           string `json:"public_key"`
	EncryptedPrivateKey string `json:"encrypted_private_key"`
	KeySalt             string `json:"key_salt"`
}

// CreateUserKey 创建用户密钥
func (s *Service) CreateUserKey(ctx context.Context, req *CreateUserKeyRequest) error {
	userKey := &model.UserKey{
		UserID:              req.UserID,
		PublicKey:           req.PublicKey,
		EncryptedPrivateKey: req.EncryptedPrivateKey,
		KeySalt:             req.KeySalt,
	}

	return s.storage.DB().Create(userKey).Error
}

// GetUserKey 获取用户密钥
func (s *Service) GetUserKey(ctx context.Context, userID string) (*model.UserKey, error) {
	var userKey model.UserKey
	if err := s.storage.DB().Where("user_id = ?", userID).First(&userKey).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound
		}
		return nil, err
	}
	return &userKey, nil
}

// GetUserPublicKey 获取用户公钥
func (s *Service) GetUserPublicKey(ctx context.Context, userID string) (string, error) {
	userKey, err := s.GetUserKey(ctx, userID)
	if err != nil {
		return "", err
	}
	return userKey.PublicKey, nil
}

// GetUserPublicKeys 批量获取用户公钥
func (s *Service) GetUserPublicKeys(ctx context.Context, userIDs []string) (map[string]string, error) {
	var userKeys []model.UserKey
	if err := s.storage.DB().Where("user_id IN ?", userIDs).Find(&userKeys).Error; err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, uk := range userKeys {
		result[uk.UserID] = uk.PublicKey
	}
	return result, nil
}

// UpdateUserKey 更新用户密钥 (换密码时需要重新加密私钥)
func (s *Service) UpdateUserKey(ctx context.Context, userID string, encryptedPrivateKey string, keySalt string) error {
	return s.storage.DB().Model(&model.UserKey{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"encrypted_private_key": encryptedPrivateKey,
			"key_salt":              keySalt,
		}).Error
}

// ===== 私聊密钥 =====

// ChatKeyEntry 私聊密钥条目
type ChatKeyEntry struct {
	UserID       string `json:"user_id"`
	EncryptedKey string `json:"encrypted_key"`
}

// CreateChatKeysRequest 创建私聊密钥请求
type CreateChatKeysRequest struct {
	ConversationID string         `json:"conversation_id"`
	Keys           []ChatKeyEntry `json:"keys"`
}

// CreateChatKeys 创建私聊密钥
func (s *Service) CreateChatKeys(ctx context.Context, req *CreateChatKeysRequest) error {
	return s.storage.DB().Transaction(func(tx *gorm.DB) error {
		for _, entry := range req.Keys {
			chatKey := &model.ChatKey{
				ConversationID: req.ConversationID,
				UserID:         entry.UserID,
				EncryptedKey:   entry.EncryptedKey,
			}
			if err := tx.Create(chatKey).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetChatKey 获取私聊密钥
func (s *Service) GetChatKey(ctx context.Context, conversationID string, userID string) (*model.ChatKey, error) {
	var chatKey model.ChatKey
	if err := s.storage.DB().
		Where("conversation_id = ? AND user_id = ?", conversationID, userID).
		First(&chatKey).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound
		}
		return nil, err
	}
	return &chatKey, nil
}

// ChatKeyExists 检查私聊密钥是否存在
func (s *Service) ChatKeyExists(ctx context.Context, conversationID string) (bool, error) {
	var count int64
	if err := s.storage.DB().Model(&model.ChatKey{}).
		Where("conversation_id = ?", conversationID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ===== 群组密钥 =====

// GroupKeyEntry 群组密钥条目
type GroupKeyEntry struct {
	UserID       string `json:"user_id"`
	EncryptedKey string `json:"encrypted_key"`
}

// CreateGroupKeysRequest 创建群组密钥请求
type CreateGroupKeysRequest struct {
	GroupID string          `json:"group_id"`
	Keys    []GroupKeyEntry `json:"keys"`
	Version int             `json:"version"`
}

// CreateGroupKeys 创建群组密钥
func (s *Service) CreateGroupKeys(ctx context.Context, req *CreateGroupKeysRequest) error {
	return s.storage.DB().Transaction(func(tx *gorm.DB) error {
		for _, entry := range req.Keys {
			groupKey := &model.GroupKey{
				GroupID:      req.GroupID,
				UserID:       entry.UserID,
				EncryptedKey: entry.EncryptedKey,
				Version:      req.Version,
			}
			if err := tx.Create(groupKey).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetGroupKey 获取群组密钥
func (s *Service) GetGroupKey(ctx context.Context, groupID string, userID string, version int) (*model.GroupKey, error) {
	var groupKey model.GroupKey
	query := s.storage.DB().Where("group_id = ? AND user_id = ?", groupID, userID)

	if version > 0 {
		query = query.Where("version = ?", version)
	} else {
		// 获取最新版本
		query = query.Order("version DESC")
	}

	if err := query.First(&groupKey).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound
		}
		return nil, err
	}
	return &groupKey, nil
}

// GetLatestGroupKeyVersion 获取群组最新密钥版本
func (s *Service) GetLatestGroupKeyVersion(ctx context.Context, groupID string) (int, error) {
	var groupKey model.GroupKey
	if err := s.storage.DB().
		Where("group_id = ?", groupID).
		Order("version DESC").
		First(&groupKey).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, err
	}
	return groupKey.Version, nil
}

// AddMemberGroupKey 为新成员添加群组密钥
func (s *Service) AddMemberGroupKey(ctx context.Context, groupID string, userID string, encryptedKey string, version int) error {
	groupKey := &model.GroupKey{
		GroupID:      groupID,
		UserID:       userID,
		EncryptedKey: encryptedKey,
		Version:      version,
	}
	return s.storage.DB().Create(groupKey).Error
}

// RemoveMemberGroupKeys 移除成员的群组密钥
func (s *Service) RemoveMemberGroupKeys(ctx context.Context, groupID string, userID string) error {
	return s.storage.DB().
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Delete(&model.GroupKey{}).Error
}

// GroupKeyExists 检查群组密钥是否存在
func (s *Service) GroupKeyExists(ctx context.Context, groupID string) (bool, error) {
	var count int64
	if err := s.storage.DB().Model(&model.GroupKey{}).
		Where("group_id = ?", groupID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
