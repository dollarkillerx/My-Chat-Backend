package relation

import (
	"context"
	"time"

	"github.com/my-chat/common/pkg/errors"
	"github.com/my-chat/seaking/internal/model"
	"github.com/my-chat/seaking/internal/storage"
	"gorm.io/gorm"
)

// Service 关系服务
type Service struct {
	storage *storage.Storage
}

// NewService 创建关系服务
func NewService(storage *storage.Storage) *Service {
	return &Service{storage: storage}
}

// SendFriendRequest 发送好友请求
func (s *Service) SendFriendRequest(ctx context.Context, fromUID, toUID, message string) error {
	// 检查是否已经是好友
	var friendship model.Friendship
	err := s.storage.DB().Where("user_id = ? AND friend_id = ?", fromUID, toUID).First(&friendship).Error
	if err == nil {
		return errors.ErrAlreadyFriend
	}

	// 检查是否有待处理的请求
	var existReq model.FriendRequest
	err = s.storage.DB().Where("from_uid = ? AND to_uid = ? AND status = ?",
		fromUID, toUID, model.FriendRequestPending).First(&existReq).Error
	if err == nil {
		return errors.New(errors.ErrCodeInvalidParam, "request already sent")
	}

	// 创建好友请求
	req := &model.FriendRequest{
		FromUID: fromUID,
		ToUID:   toUID,
		Message: message,
		Status:  model.FriendRequestPending,
	}

	return s.storage.DB().Create(req).Error
}

// AcceptFriendRequest 接受好友请求
func (s *Service) AcceptFriendRequest(ctx context.Context, requestID uint, uid string) error {
	var req model.FriendRequest
	if err := s.storage.DB().First(&req, requestID).Error; err != nil {
		return errors.ErrNotFound
	}

	if req.ToUID != uid {
		return errors.ErrForbidden
	}

	if req.Status != model.FriendRequestPending {
		return errors.New(errors.ErrCodeInvalidParam, "request already handled")
	}

	// 使用事务
	return s.storage.DB().Transaction(func(tx *gorm.DB) error {
		// 更新请求状态
		if err := tx.Model(&req).Update("status", model.FriendRequestAccepted).Error; err != nil {
			return err
		}

		// 创建双向好友关系
		now := time.Now()
		friendships := []model.Friendship{
			{UserID: req.FromUID, FriendID: req.ToUID, Status: model.FriendStatusNormal, CreatedAt: now},
			{UserID: req.ToUID, FriendID: req.FromUID, Status: model.FriendStatusNormal, CreatedAt: now},
		}

		return tx.Create(&friendships).Error
	})
}

// RejectFriendRequest 拒绝好友请求
func (s *Service) RejectFriendRequest(ctx context.Context, requestID uint, uid string) error {
	var req model.FriendRequest
	if err := s.storage.DB().First(&req, requestID).Error; err != nil {
		return errors.ErrNotFound
	}

	if req.ToUID != uid {
		return errors.ErrForbidden
	}

	return s.storage.DB().Model(&req).Update("status", model.FriendRequestRejected).Error
}

// DeleteFriend 删除好友
func (s *Service) DeleteFriend(ctx context.Context, uid, friendID string) error {
	return s.storage.DB().Transaction(func(tx *gorm.DB) error {
		// 删除双向好友关系
		if err := tx.Where("user_id = ? AND friend_id = ?", uid, friendID).Delete(&model.Friendship{}).Error; err != nil {
			return err
		}
		return tx.Where("user_id = ? AND friend_id = ?", friendID, uid).Delete(&model.Friendship{}).Error
	})
}

// BlockFriend 拉黑好友
func (s *Service) BlockFriend(ctx context.Context, uid, friendID string) error {
	return s.storage.DB().Model(&model.Friendship{}).
		Where("user_id = ? AND friend_id = ?", uid, friendID).
		Update("status", model.FriendStatusBlocked).Error
}

// UnblockFriend 取消拉黑
func (s *Service) UnblockFriend(ctx context.Context, uid, friendID string) error {
	return s.storage.DB().Model(&model.Friendship{}).
		Where("user_id = ? AND friend_id = ?", uid, friendID).
		Update("status", model.FriendStatusNormal).Error
}

// GetFriends 获取好友列表
func (s *Service) GetFriends(ctx context.Context, uid string) ([]model.Friendship, error) {
	var friends []model.Friendship
	err := s.storage.DB().Where("user_id = ? AND status = ?", uid, model.FriendStatusNormal).Find(&friends).Error
	return friends, err
}

// GetPendingRequests 获取待处理的好友请求
func (s *Service) GetPendingRequests(ctx context.Context, uid string) ([]model.FriendRequest, error) {
	var requests []model.FriendRequest
	err := s.storage.DB().Where("to_uid = ? AND status = ?", uid, model.FriendRequestPending).Find(&requests).Error
	return requests, err
}

// IsFriend 检查是否是好友
func (s *Service) IsFriend(ctx context.Context, uid1, uid2 string) bool {
	var friendship model.Friendship
	err := s.storage.DB().Where("user_id = ? AND friend_id = ? AND status = ?",
		uid1, uid2, model.FriendStatusNormal).First(&friendship).Error
	return err == nil
}

// IsBlocked 检查是否被拉黑
func (s *Service) IsBlocked(ctx context.Context, uid, targetUID string) bool {
	var friendship model.Friendship
	err := s.storage.DB().Where("user_id = ? AND friend_id = ? AND status = ?",
		targetUID, uid, model.FriendStatusBlocked).First(&friendship).Error
	return err == nil
}
