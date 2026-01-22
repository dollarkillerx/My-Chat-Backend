package client

import (
	"context"
	"time"
)

// SeaKingClient SeaKing服务RPC客户端
type SeaKingClient struct {
	rpc *RPCClient
}

// NewSeaKingClient 创建SeaKing客户端
func NewSeaKingClient(addr string) *SeaKingClient {
	return &SeaKingClient{
		rpc: NewRPCClient(addr+"/api/rpc", WithRPCTimeout(5*time.Second)),
	}
}

// CheckAccessRequest 检查访问权限请求
type CheckAccessRequest struct {
	Uid string `json:"uid"`
	Cid string `json:"cid"`
}

// CheckAccessResponse 检查访问权限响应
type CheckAccessResponse struct {
	HasAccess bool   `json:"has_access"`
	Role      int    `json:"role"`      // 0=普通成员, 1=管理员, 2=群主
	Muted     bool   `json:"muted"`     // 是否被禁言
	Reason    string `json:"reason,omitempty"`
}

// CheckAccess 检查用户是否有权访问会话
func (c *SeaKingClient) CheckAccess(ctx context.Context, uid, cid string) (*CheckAccessResponse, error) {
	var resp CheckAccessResponse
	err := c.rpc.Call(ctx, "seaking.checkAccess", &CheckAccessRequest{Uid: uid, Cid: cid}, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetConversationRequest 获取会话信息请求
type GetConversationRequest struct {
	Cid string `json:"cid"`
}

// ConversationInfo 会话信息
type ConversationInfo struct {
	Cid       string   `json:"cid"`
	Type      int      `json:"type"`       // 1=单聊, 2=群聊
	Name      string   `json:"name"`
	Avatar    string   `json:"avatar"`
	MemberIds []string `json:"member_ids"`
}

// GetConversation 获取会话信息
func (c *SeaKingClient) GetConversation(ctx context.Context, cid string) (*ConversationInfo, error) {
	var resp ConversationInfo
	err := c.rpc.Call(ctx, "seaking.getConversation", &GetConversationRequest{Cid: cid}, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetConversationMembersRequest 获取会话成员请求
type GetConversationMembersRequest struct {
	Cid string `json:"cid"`
}

// MemberInfo 成员信息
type MemberInfo struct {
	Uid      string `json:"uid"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Role     int    `json:"role"`
}

// GetConversationMembersResponse 获取会话成员响应
type GetConversationMembersResponse struct {
	Members []MemberInfo `json:"members"`
}

// GetConversationMembers 获取会话成员列表
func (c *SeaKingClient) GetConversationMembers(ctx context.Context, cid string) (*GetConversationMembersResponse, error) {
	var resp GetConversationMembersResponse
	err := c.rpc.Call(ctx, "seaking.getConversationMembers", &GetConversationMembersRequest{Cid: cid}, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateConversationRequest 创建会话请求
type CreateConversationRequest struct {
	Type      int      `json:"type"`       // 1=单聊, 2=群聊
	CreatorId string   `json:"creator_id"`
	MemberIds []string `json:"member_ids"`
	Name      string   `json:"name,omitempty"`
}

// CreateConversationResponse 创建会话响应
type CreateConversationResponse struct {
	Cid string `json:"cid"`
}

// CreateConversation 创建会话
func (c *SeaKingClient) CreateConversation(ctx context.Context, convType int, creatorId string, memberIds []string, name string) (*CreateConversationResponse, error) {
	var resp CreateConversationResponse
	err := c.rpc.Call(ctx, "seaking.createConversation", &CreateConversationRequest{
		Type:      convType,
		CreatorId: creatorId,
		MemberIds: memberIds,
		Name:      name,
	}, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetUserConversationsRequest 获取用户会话列表请求
type GetUserConversationsRequest struct {
	Uid string `json:"uid"`
}

// GetUserConversationsResponse 获取用户会话列表响应
type GetUserConversationsResponse struct {
	Conversations []ConversationInfo `json:"conversations"`
}

// GetUserConversations 获取用户的所有会话
func (c *SeaKingClient) GetUserConversations(ctx context.Context, uid string) (*GetUserConversationsResponse, error) {
	var resp GetUserConversationsResponse
	err := c.rpc.Call(ctx, "seaking.getUserConversations", &GetUserConversationsRequest{Uid: uid}, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ValidateTokenRequest 验证Token请求
type ValidateTokenRequest struct {
	Token string `json:"token"`
}

// ValidateTokenResponse 验证Token响应
type ValidateTokenResponse struct {
	Valid    bool   `json:"valid"`
	Uid      string `json:"uid"`
	DeviceId string `json:"device_id"`
	Platform string `json:"platform"`
}

// ValidateToken 验证Token
func (c *SeaKingClient) ValidateToken(ctx context.Context, token string) (*ValidateTokenResponse, error) {
	var resp ValidateTokenResponse
	err := c.rpc.Call(ctx, "seaking.validateToken", &ValidateTokenRequest{Token: token}, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetUserInfoRequest 获取用户信息请求
type GetUserInfoRequest struct {
	Uid string `json:"uid"`
}

// UserInfo 用户信息
type UserInfo struct {
	Uid      string `json:"uid"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Status   int    `json:"status"`
}

// GetUserInfo 获取用户信息
func (c *SeaKingClient) GetUserInfo(ctx context.Context, uid string) (*UserInfo, error) {
	var resp UserInfo
	err := c.rpc.Call(ctx, "seaking.getUserInfo", &GetUserInfoRequest{Uid: uid}, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
	Phone    string `json:"phone,omitempty"`
	Email    string `json:"email,omitempty"`
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	Uid      string `json:"uid"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
}

// Register 用户注册
func (c *SeaKingClient) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	var resp RegisterResponse
	err := c.rpc.Call(ctx, "seaking.register", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token string    `json:"token"`
	User  *UserInfo `json:"user"`
}

// Login 用户登录
func (c *SeaKingClient) Login(ctx context.Context, req *LoginRequest, deviceId, platform string) (*LoginResponse, error) {
	var resp LoginResponse
	err := c.rpc.Call(ctx, "seaking.login", map[string]interface{}{
		"username":  req.Username,
		"password":  req.Password,
		"device_id": deviceId,
		"platform":  platform,
	}, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// FriendInfo 好友信息
type FriendInfo struct {
	Uid      string `json:"uid"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Remark   string `json:"remark"`
}

// GetFriends 获取好友列表
func (c *SeaKingClient) GetFriends(ctx context.Context, uid string) ([]FriendInfo, error) {
	var resp struct {
		Friends []FriendInfo `json:"friends"`
	}
	err := c.rpc.Call(ctx, "seaking.getFriends", map[string]string{"uid": uid}, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Friends, nil
}

// SendFriendRequest 发送好友请求
func (c *SeaKingClient) SendFriendRequest(ctx context.Context, fromUid, toUid, message string) error {
	var resp struct{}
	return c.rpc.Call(ctx, "seaking.sendFriendRequest", map[string]string{
		"from_uid": fromUid,
		"to_uid":   toUid,
		"message":  message,
	}, &resp)
}

// FriendRequestInfo 好友请求信息
type FriendRequestInfo struct {
	RequestId uint   `json:"request_id"`
	FromUid   string `json:"from_uid"`
	ToUid     string `json:"to_uid"`
	Message   string `json:"message"`
	Status    int    `json:"status"`
}

// GetPendingFriendRequests 获取待处理的好友请求
func (c *SeaKingClient) GetPendingFriendRequests(ctx context.Context, uid string) ([]FriendRequestInfo, error) {
	var resp struct {
		Requests []FriendRequestInfo `json:"requests"`
	}
	err := c.rpc.Call(ctx, "seaking.getPendingFriendRequests", map[string]string{"uid": uid}, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Requests, nil
}

// AcceptFriendRequest 接受好友请求
func (c *SeaKingClient) AcceptFriendRequest(ctx context.Context, requestId uint, uid string) error {
	var resp struct{}
	return c.rpc.Call(ctx, "seaking.acceptFriendRequest", map[string]interface{}{
		"request_id": requestId,
		"uid":        uid,
	}, &resp)
}

// RejectFriendRequest 拒绝好友请求
func (c *SeaKingClient) RejectFriendRequest(ctx context.Context, requestId uint, uid string) error {
	var resp struct{}
	return c.rpc.Call(ctx, "seaking.rejectFriendRequest", map[string]interface{}{
		"request_id": requestId,
		"uid":        uid,
	}, &resp)
}

// DeleteFriend 删除好友
func (c *SeaKingClient) DeleteFriend(ctx context.Context, uid, friendId string) error {
	var resp struct{}
	return c.rpc.Call(ctx, "seaking.deleteFriend", map[string]string{
		"uid":       uid,
		"friend_id": friendId,
	}, &resp)
}

// GroupInfo 群组信息
type GroupInfo struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Avatar      string `json:"avatar"`
	OwnerId     string `json:"owner_id"`
	MaxMembers  int    `json:"max_members"`
}

// GetUserGroups 获取用户的群组列表
func (c *SeaKingClient) GetUserGroups(ctx context.Context, uid string) ([]GroupInfo, error) {
	var resp struct {
		Groups []GroupInfo `json:"groups"`
	}
	err := c.rpc.Call(ctx, "seaking.getUserGroups", map[string]string{"uid": uid}, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Groups, nil
}

// CreateGroup 创建群组
func (c *SeaKingClient) CreateGroup(ctx context.Context, ownerId, name, description string, memberIds []string) (*GroupInfo, error) {
	var resp GroupInfo
	err := c.rpc.Call(ctx, "seaking.createGroup", map[string]interface{}{
		"owner_id":    ownerId,
		"name":        name,
		"description": description,
		"member_ids":  memberIds,
	}, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetGroupInfo 获取群组信息
func (c *SeaKingClient) GetGroupInfo(ctx context.Context, groupId string) (*GroupInfo, error) {
	var resp GroupInfo
	err := c.rpc.Call(ctx, "seaking.getGroupInfo", map[string]string{"group_id": groupId}, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetGroupMembers 获取群组成员
func (c *SeaKingClient) GetGroupMembers(ctx context.Context, groupId string) ([]MemberInfo, error) {
	var resp struct {
		Members []MemberInfo `json:"members"`
	}
	err := c.rpc.Call(ctx, "seaking.getGroupMembers", map[string]string{"group_id": groupId}, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Members, nil
}
