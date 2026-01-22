package group

import (
	"context"
	"time"

	"github.com/my-chat/common/pkg/errors"
	"github.com/my-chat/seaking/internal/model"
	"github.com/my-chat/seaking/internal/storage"
	"github.com/rs/xid"
	"gorm.io/gorm"
)

// Service 群组服务
type Service struct {
	storage *storage.Storage
}

// NewService 创建群组服务
func NewService(storage *storage.Storage) *Service {
	return &Service{storage: storage}
}

// CreateGroupRequest 创建群组请求
type CreateGroupRequest struct {
	Name        string   `json:"name" binding:"required,min=1,max=64"`
	Description string   `json:"description"`
	MemberIDs   []string `json:"member_ids"` // 初始成员
}

// CreateGroup 创建群组
func (s *Service) CreateGroup(ctx context.Context, ownerID string, req *CreateGroupRequest) (*model.Group, error) {
	group := &model.Group{
		ID:          xid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		OwnerID:     ownerID,
		MaxMembers:  500,
		Status:      model.GroupStatusNormal,
	}

	err := s.storage.DB().Transaction(func(tx *gorm.DB) error {
		// 创建群组
		if err := tx.Create(group).Error; err != nil {
			return err
		}

		// 添加群主为成员
		ownerMember := &model.GroupMember{
			GroupID:  group.ID,
			UserID:   ownerID,
			Role:     model.GroupRoleOwner,
			JoinedAt: time.Now(),
		}
		if err := tx.Create(ownerMember).Error; err != nil {
			return err
		}

		// 添加初始成员
		for _, memberID := range req.MemberIDs {
			if memberID == ownerID {
				continue
			}
			member := &model.GroupMember{
				GroupID:  group.ID,
				UserID:   memberID,
				Role:     model.GroupRoleMember,
				JoinedAt: time.Now(),
			}
			if err := tx.Create(member).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return group, nil
}

// GetGroup 获取群组信息
func (s *Service) GetGroup(ctx context.Context, groupID string) (*model.Group, error) {
	var group model.Group
	if err := s.storage.DB().First(&group, "id = ?", groupID).Error; err != nil {
		return nil, errors.ErrGroupNotFound
	}
	return &group, nil
}

// UpdateGroup 更新群组信息
func (s *Service) UpdateGroup(ctx context.Context, groupID, operatorID string, name, description, avatar string) error {
	// 检查权限
	if !s.HasPermission(ctx, groupID, operatorID, model.GroupRoleAdmin) {
		return errors.ErrNoPermission
	}

	updates := map[string]interface{}{}
	if name != "" {
		updates["name"] = name
	}
	if description != "" {
		updates["description"] = description
	}
	if avatar != "" {
		updates["avatar"] = avatar
	}

	return s.storage.DB().Model(&model.Group{}).Where("id = ?", groupID).Updates(updates).Error
}

// DismissGroup 解散群组
func (s *Service) DismissGroup(ctx context.Context, groupID, operatorID string) error {
	var group model.Group
	if err := s.storage.DB().First(&group, "id = ?", groupID).Error; err != nil {
		return errors.ErrGroupNotFound
	}

	if group.OwnerID != operatorID {
		return errors.ErrNoPermission
	}

	return s.storage.DB().Transaction(func(tx *gorm.DB) error {
		// 删除所有成员
		if err := tx.Where("group_id = ?", groupID).Delete(&model.GroupMember{}).Error; err != nil {
			return err
		}
		// 标记群组为解散
		return tx.Model(&group).Update("status", model.GroupStatusDissolved).Error
	})
}

// AddMember 添加成员
func (s *Service) AddMember(ctx context.Context, groupID, operatorID, userID string) error {
	// 检查权限
	if !s.HasPermission(ctx, groupID, operatorID, model.GroupRoleAdmin) {
		return errors.ErrNoPermission
	}

	// 检查群组是否存在
	var group model.Group
	if err := s.storage.DB().First(&group, "id = ?", groupID).Error; err != nil {
		return errors.ErrGroupNotFound
	}

	// 检查是否已是成员
	var existMember model.GroupMember
	if err := s.storage.DB().Where("group_id = ? AND user_id = ?", groupID, userID).First(&existMember).Error; err == nil {
		return errors.New(errors.ErrCodeInvalidParam, "user already in group")
	}

	// 检查人数限制
	var count int64
	s.storage.DB().Model(&model.GroupMember{}).Where("group_id = ?", groupID).Count(&count)
	if int(count) >= group.MaxMembers {
		return errors.ErrConversationFull
	}

	member := &model.GroupMember{
		GroupID:  groupID,
		UserID:   userID,
		Role:     model.GroupRoleMember,
		JoinedAt: time.Now(),
	}

	return s.storage.DB().Create(member).Error
}

// RemoveMember 移除成员
func (s *Service) RemoveMember(ctx context.Context, groupID, operatorID, userID string) error {
	// 检查权限
	if !s.HasPermission(ctx, groupID, operatorID, model.GroupRoleAdmin) {
		return errors.ErrNoPermission
	}

	// 不能移除群主
	var group model.Group
	if err := s.storage.DB().First(&group, "id = ?", groupID).Error; err != nil {
		return errors.ErrGroupNotFound
	}
	if group.OwnerID == userID {
		return errors.New(errors.ErrCodeInvalidParam, "cannot remove owner")
	}

	return s.storage.DB().Where("group_id = ? AND user_id = ?", groupID, userID).Delete(&model.GroupMember{}).Error
}

// LeaveGroup 退出群组
func (s *Service) LeaveGroup(ctx context.Context, groupID, userID string) error {
	var group model.Group
	if err := s.storage.DB().First(&group, "id = ?", groupID).Error; err != nil {
		return errors.ErrGroupNotFound
	}

	// 群主不能退出，只能转让或解散
	if group.OwnerID == userID {
		return errors.New(errors.ErrCodeInvalidParam, "owner cannot leave, transfer or dismiss instead")
	}

	return s.storage.DB().Where("group_id = ? AND user_id = ?", groupID, userID).Delete(&model.GroupMember{}).Error
}

// SetAdmin 设置/取消管理员
func (s *Service) SetAdmin(ctx context.Context, groupID, operatorID, userID string, isAdmin bool) error {
	var group model.Group
	if err := s.storage.DB().First(&group, "id = ?", groupID).Error; err != nil {
		return errors.ErrGroupNotFound
	}

	// 只有群主可以设置管理员
	if group.OwnerID != operatorID {
		return errors.ErrNoPermission
	}

	role := model.GroupRoleMember
	if isAdmin {
		role = model.GroupRoleAdmin
	}

	return s.storage.DB().Model(&model.GroupMember{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Update("role", role).Error
}

// TransferOwner 转让群主
func (s *Service) TransferOwner(ctx context.Context, groupID, ownerID, newOwnerID string) error {
	var group model.Group
	if err := s.storage.DB().First(&group, "id = ?", groupID).Error; err != nil {
		return errors.ErrGroupNotFound
	}

	if group.OwnerID != ownerID {
		return errors.ErrNoPermission
	}

	return s.storage.DB().Transaction(func(tx *gorm.DB) error {
		// 更新群组所有者
		if err := tx.Model(&group).Update("owner_id", newOwnerID).Error; err != nil {
			return err
		}

		// 更新原群主为普通成员
		if err := tx.Model(&model.GroupMember{}).
			Where("group_id = ? AND user_id = ?", groupID, ownerID).
			Update("role", model.GroupRoleMember).Error; err != nil {
			return err
		}

		// 更新新群主角色
		return tx.Model(&model.GroupMember{}).
			Where("group_id = ? AND user_id = ?", groupID, newOwnerID).
			Update("role", model.GroupRoleOwner).Error
	})
}

// GetMembers 获取群成员列表
func (s *Service) GetMembers(ctx context.Context, groupID string) ([]model.GroupMember, error) {
	var members []model.GroupMember
	err := s.storage.DB().Where("group_id = ?", groupID).Find(&members).Error
	return members, err
}

// GetUserGroups 获取用户所在的群组
func (s *Service) GetUserGroups(ctx context.Context, userID string) ([]model.Group, error) {
	var groups []model.Group
	err := s.storage.DB().
		Joins("JOIN group_members ON group_members.group_id = groups.id").
		Where("group_members.user_id = ? AND groups.status = ?", userID, model.GroupStatusNormal).
		Find(&groups).Error
	return groups, err
}

// IsMember 检查是否是群成员
func (s *Service) IsMember(ctx context.Context, groupID, userID string) bool {
	var member model.GroupMember
	err := s.storage.DB().Where("group_id = ? AND user_id = ?", groupID, userID).First(&member).Error
	return err == nil
}

// HasPermission 检查是否有权限（管理员或以上）
func (s *Service) HasPermission(ctx context.Context, groupID, userID string, minRole int) bool {
	var member model.GroupMember
	err := s.storage.DB().Where("group_id = ? AND user_id = ?", groupID, userID).First(&member).Error
	if err != nil {
		return false
	}
	return member.Role >= minRole
}

// MuteMember 禁言成员
func (s *Service) MuteMember(ctx context.Context, groupID, operatorID, userID string, duration time.Duration) error {
	if !s.HasPermission(ctx, groupID, operatorID, model.GroupRoleAdmin) {
		return errors.ErrNoPermission
	}

	mutedAt := time.Now()
	return s.storage.DB().Model(&model.GroupMember{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Updates(map[string]interface{}{
			"muted":    true,
			"muted_at": mutedAt,
		}).Error
}

// UnmuteMember 取消禁言
func (s *Service) UnmuteMember(ctx context.Context, groupID, operatorID, userID string) error {
	if !s.HasPermission(ctx, groupID, operatorID, model.GroupRoleAdmin) {
		return errors.ErrNoPermission
	}

	return s.storage.DB().Model(&model.GroupMember{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Updates(map[string]interface{}{
			"muted":    false,
			"muted_at": nil,
		}).Error
}
