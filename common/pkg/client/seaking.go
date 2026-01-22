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
