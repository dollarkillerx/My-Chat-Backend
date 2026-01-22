package user

import (
	"context"

	"github.com/my-chat/common/pkg/crypto"
	"github.com/my-chat/common/pkg/errors"
	"github.com/my-chat/seaking/internal/model"
	"github.com/my-chat/seaking/internal/storage"
	"github.com/rs/xid"
)

// Service 用户服务
type Service struct {
	storage *storage.Storage
}

// NewService 创建用户服务
func NewService(storage *storage.Storage) *Service {
	return &Service{storage: storage}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=6,max=32"`
	Nickname string `json:"nickname"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
}

// Register 用户注册
func (s *Service) Register(ctx context.Context, req *RegisterRequest) (*model.User, error) {
	// 检查用户名是否已存在
	var existUser model.User
	if err := s.storage.DB().Where("username = ?", req.Username).First(&existUser).Error; err == nil {
		return nil, errors.ErrUserExists
	}

	// 密码加密
	hashedPassword, err := crypto.HashPassword(req.Password)
	if err != nil {
		return nil, errors.ErrInternal
	}

	// 创建用户
	user := &model.User{
		ID:       xid.New().String(),
		Username: req.Username,
		Nickname: req.Nickname,
		Password: hashedPassword,
		Phone:    req.Phone,
		Email:    req.Email,
		Status:   model.UserStatusNormal,
	}

	if user.Nickname == "" {
		user.Nickname = user.Username
	}

	if err := s.storage.DB().Create(user).Error; err != nil {
		return nil, errors.ErrInternal
	}

	return user, nil
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login 用户登录
func (s *Service) Login(ctx context.Context, req *LoginRequest) (*model.User, error) {
	var user model.User
	if err := s.storage.DB().Where("username = ?", req.Username).First(&user).Error; err != nil {
		return nil, errors.ErrUserNotFound
	}

	if user.Status == model.UserStatusDisabled {
		return nil, errors.ErrUserDisabled
	}

	if !crypto.CheckPassword(req.Password, user.Password) {
		return nil, errors.ErrPasswordWrong
	}

	return &user, nil
}

// GetByID 根据ID获取用户
func (s *Service) GetByID(ctx context.Context, id string) (*model.User, error) {
	var user model.User
	if err := s.storage.DB().First(&user, "id = ?", id).Error; err != nil {
		return nil, errors.ErrUserNotFound
	}
	return &user, nil
}

// GetByUsername 根据用户名获取用户
func (s *Service) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	if err := s.storage.DB().Where("username = ?", username).First(&user).Error; err != nil {
		return nil, errors.ErrUserNotFound
	}
	return &user, nil
}

// UpdateProfile 更新用户资料
func (s *Service) UpdateProfile(ctx context.Context, id string, nickname, avatar string) error {
	updates := map[string]interface{}{}
	if nickname != "" {
		updates["nickname"] = nickname
	}
	if avatar != "" {
		updates["avatar"] = avatar
	}

	if len(updates) == 0 {
		return nil
	}

	return s.storage.DB().Model(&model.User{}).Where("id = ?", id).Updates(updates).Error
}

// ChangePassword 修改密码
func (s *Service) ChangePassword(ctx context.Context, id, oldPassword, newPassword string) error {
	var user model.User
	if err := s.storage.DB().First(&user, "id = ?", id).Error; err != nil {
		return errors.ErrUserNotFound
	}

	if !crypto.CheckPassword(oldPassword, user.Password) {
		return errors.ErrPasswordWrong
	}

	hashedPassword, err := crypto.HashPassword(newPassword)
	if err != nil {
		return errors.ErrInternal
	}

	return s.storage.DB().Model(&user).Update("password", hashedPassword).Error
}
