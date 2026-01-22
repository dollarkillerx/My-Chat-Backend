package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/my-chat/common/pkg/auth"
	commonerrors "github.com/my-chat/common/pkg/errors"
	"github.com/my-chat/seaking/internal/service/group"
	"github.com/my-chat/seaking/internal/service/relation"
	"github.com/my-chat/seaking/internal/service/user"
)

// API 接口层
type API struct {
	userService     *user.Service
	relationService *relation.Service
	groupService    *group.Service
	jwtManager      *auth.JWTManager
}

// NewAPI 创建API
func NewAPI(
	userService *user.Service,
	relationService *relation.Service,
	groupService *group.Service,
	jwtManager *auth.JWTManager,
) *API {
	return &API{
		userService:     userService,
		relationService: relationService,
		groupService:    groupService,
		jwtManager:      jwtManager,
	}
}

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// Error 错误响应
func Error(c *gin.Context, err error) {
	if e, ok := err.(*commonerrors.Error); ok {
		c.JSON(http.StatusOK, Response{
			Code:    e.Code,
			Message: e.Message,
		})
		return
	}
	c.JSON(http.StatusOK, Response{
		Code:    commonerrors.ErrCodeUnknown,
		Message: err.Error(),
	})
}

// Register 注册
func (a *API) Register(c *gin.Context) {
	var req user.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, commonerrors.ErrInvalidParam)
		return
	}

	u, err := a.userService.Register(c.Request.Context(), &req)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{
		"id":       u.ID,
		"username": u.Username,
		"nickname": u.Nickname,
	})
}

// Login 登录
func (a *API) Login(c *gin.Context) {
	var req user.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, commonerrors.ErrInvalidParam)
		return
	}

	deviceId := c.GetHeader("X-Device-ID")
	platform := c.GetHeader("X-Platform")

	u, err := a.userService.Login(c.Request.Context(), &req)
	if err != nil {
		Error(c, err)
		return
	}

	token, err := a.jwtManager.GenerateToken(u.ID, deviceId, platform)
	if err != nil {
		Error(c, commonerrors.ErrInternal)
		return
	}

	Success(c, gin.H{
		"token": token,
		"user": gin.H{
			"id":       u.ID,
			"username": u.Username,
			"nickname": u.Nickname,
			"avatar":   u.Avatar,
		},
	})
}

// GetProfile 获取用户资料
func (a *API) GetProfile(c *gin.Context) {
	uid := c.GetString("uid")

	u, err := a.userService.GetByID(c.Request.Context(), uid)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{
		"id":       u.ID,
		"username": u.Username,
		"nickname": u.Nickname,
		"avatar":   u.Avatar,
		"phone":    u.Phone,
		"email":    u.Email,
	})
}

// UpdateProfile 更新资料
func (a *API) UpdateProfile(c *gin.Context) {
	uid := c.GetString("uid")

	var req struct {
		Nickname string `json:"nickname"`
		Avatar   string `json:"avatar"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, commonerrors.ErrInvalidParam)
		return
	}

	if err := a.userService.UpdateProfile(c.Request.Context(), uid, req.Nickname, req.Avatar); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// GetFriends 获取好友列表
func (a *API) GetFriends(c *gin.Context) {
	uid := c.GetString("uid")

	friends, err := a.relationService.GetFriends(c.Request.Context(), uid)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, friends)
}

// SendFriendRequest 发送好友请求
func (a *API) SendFriendRequest(c *gin.Context) {
	uid := c.GetString("uid")

	var req struct {
		ToUID   string `json:"to_uid" binding:"required"`
		Message string `json:"message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, commonerrors.ErrInvalidParam)
		return
	}

	if err := a.relationService.SendFriendRequest(c.Request.Context(), uid, req.ToUID, req.Message); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// CreateGroup 创建群组
func (a *API) CreateGroup(c *gin.Context) {
	uid := c.GetString("uid")

	var req group.CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, commonerrors.ErrInvalidParam)
		return
	}

	g, err := a.groupService.CreateGroup(c.Request.Context(), uid, &req)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{
		"id":   g.ID,
		"name": g.Name,
	})
}

// GetGroupMembers 获取群成员
func (a *API) GetGroupMembers(c *gin.Context) {
	groupID := c.Param("group_id")

	members, err := a.groupService.GetMembers(c.Request.Context(), groupID)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, members)
}

// ========== 好友相关 API ==========

// AcceptFriendRequest 接受好友请求
func (a *API) AcceptFriendRequest(c *gin.Context) {
	uid := c.GetString("uid")

	var req struct {
		RequestID uint `json:"request_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, commonerrors.ErrInvalidParam)
		return
	}

	if err := a.relationService.AcceptFriendRequest(c.Request.Context(), req.RequestID, uid); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// RejectFriendRequest 拒绝好友请求
func (a *API) RejectFriendRequest(c *gin.Context) {
	uid := c.GetString("uid")

	var req struct {
		RequestID uint `json:"request_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, commonerrors.ErrInvalidParam)
		return
	}

	if err := a.relationService.RejectFriendRequest(c.Request.Context(), req.RequestID, uid); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// DeleteFriend 删除好友
func (a *API) DeleteFriend(c *gin.Context) {
	uid := c.GetString("uid")
	friendID := c.Param("uid")

	if err := a.relationService.DeleteFriend(c.Request.Context(), uid, friendID); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// BlockFriend 拉黑好友
func (a *API) BlockFriend(c *gin.Context) {
	uid := c.GetString("uid")

	var req struct {
		FriendID string `json:"friend_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, commonerrors.ErrInvalidParam)
		return
	}

	if err := a.relationService.BlockFriend(c.Request.Context(), uid, req.FriendID); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// UnblockFriend 取消拉黑
func (a *API) UnblockFriend(c *gin.Context) {
	uid := c.GetString("uid")

	var req struct {
		FriendID string `json:"friend_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, commonerrors.ErrInvalidParam)
		return
	}

	if err := a.relationService.UnblockFriend(c.Request.Context(), uid, req.FriendID); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// GetPendingRequests 获取待处理的好友请求
func (a *API) GetPendingRequests(c *gin.Context) {
	uid := c.GetString("uid")

	requests, err := a.relationService.GetPendingRequests(c.Request.Context(), uid)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, requests)
}

// ========== 群组相关 API ==========

// GetUserGroups 获取用户所在的群组列表
func (a *API) GetUserGroups(c *gin.Context) {
	uid := c.GetString("uid")

	groups, err := a.groupService.GetUserGroups(c.Request.Context(), uid)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, groups)
}

// GetGroup 获取群组详情
func (a *API) GetGroup(c *gin.Context) {
	groupID := c.Param("group_id")

	g, err := a.groupService.GetGroup(c.Request.Context(), groupID)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{
		"id":          g.ID,
		"name":        g.Name,
		"description": g.Description,
		"avatar":      g.Avatar,
		"owner_id":    g.OwnerID,
		"max_members": g.MaxMembers,
		"status":      g.Status,
		"created_at":  g.CreatedAt,
	})
}

// UpdateGroup 更新群组信息
func (a *API) UpdateGroup(c *gin.Context) {
	uid := c.GetString("uid")
	groupID := c.Param("group_id")

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Avatar      string `json:"avatar"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, commonerrors.ErrInvalidParam)
		return
	}

	if err := a.groupService.UpdateGroup(c.Request.Context(), groupID, uid, req.Name, req.Description, req.Avatar); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// DismissGroup 解散群组
func (a *API) DismissGroup(c *gin.Context) {
	uid := c.GetString("uid")
	groupID := c.Param("group_id")

	if err := a.groupService.DismissGroup(c.Request.Context(), groupID, uid); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// AddGroupMember 添加群成员
func (a *API) AddGroupMember(c *gin.Context) {
	uid := c.GetString("uid")
	groupID := c.Param("group_id")

	var req struct {
		UserID string `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, commonerrors.ErrInvalidParam)
		return
	}

	if err := a.groupService.AddMember(c.Request.Context(), groupID, uid, req.UserID); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// RemoveGroupMember 移除群成员
func (a *API) RemoveGroupMember(c *gin.Context) {
	uid := c.GetString("uid")
	groupID := c.Param("group_id")
	memberID := c.Param("member_id")

	if err := a.groupService.RemoveMember(c.Request.Context(), groupID, uid, memberID); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// LeaveGroup 退出群组
func (a *API) LeaveGroup(c *gin.Context) {
	uid := c.GetString("uid")
	groupID := c.Param("group_id")

	if err := a.groupService.LeaveGroup(c.Request.Context(), groupID, uid); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// TransferGroupOwner 转让群主
func (a *API) TransferGroupOwner(c *gin.Context) {
	uid := c.GetString("uid")
	groupID := c.Param("group_id")

	var req struct {
		NewOwnerID string `json:"new_owner_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, commonerrors.ErrInvalidParam)
		return
	}

	if err := a.groupService.TransferOwner(c.Request.Context(), groupID, uid, req.NewOwnerID); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// SetGroupAdmin 设置/取消管理员
func (a *API) SetGroupAdmin(c *gin.Context) {
	uid := c.GetString("uid")
	groupID := c.Param("group_id")

	var req struct {
		UserID  string `json:"user_id" binding:"required"`
		IsAdmin bool   `json:"is_admin"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, commonerrors.ErrInvalidParam)
		return
	}

	if err := a.groupService.SetAdmin(c.Request.Context(), groupID, uid, req.UserID, req.IsAdmin); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// ========== 用户相关 API ==========

// ChangePassword 修改密码
func (a *API) ChangePassword(c *gin.Context) {
	uid := c.GetString("uid")

	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6,max=32"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, commonerrors.ErrInvalidParam)
		return
	}

	if err := a.userService.ChangePassword(c.Request.Context(), uid, req.OldPassword, req.NewPassword); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}
