package rpc

import (
	"context"
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/my-chat/common/pkg/auth"
	"github.com/my-chat/seaking/internal/service/conversation"
	"github.com/my-chat/seaking/internal/service/user"
)

// Handler RPC处理器
type Handler struct {
	userService *user.Service
	convService *conversation.Service
	jwtManager  *auth.JWTManager
	methods     map[string]MethodHandler
}

// MethodHandler 方法处理函数
type MethodHandler func(ctx context.Context, params json.RawMessage) (interface{}, error)

// Request JSON-RPC请求
type Request struct {
	JsonRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	ID      string          `json:"id"`
}

// Response JSON-RPC响应
type Response struct {
	JsonRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *Error      `json:"error,omitempty"`
	ID      string      `json:"id"`
}

// Error JSON-RPC错误
type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// NewHandler 创建RPC处理器
func NewHandler(userService *user.Service, convService *conversation.Service, jwtManager *auth.JWTManager) *Handler {
	h := &Handler{
		userService: userService,
		convService: convService,
		jwtManager:  jwtManager,
		methods:     make(map[string]MethodHandler),
	}
	h.registerMethods()
	return h
}

// registerMethods 注册所有方法
func (h *Handler) registerMethods() {
	h.methods["seaking.checkAccess"] = h.checkAccess
	h.methods["seaking.getConversation"] = h.getConversation
	h.methods["seaking.getConversationMembers"] = h.getConversationMembers
	h.methods["seaking.createConversation"] = h.createConversation
	h.methods["seaking.getUserConversations"] = h.getUserConversations
	h.methods["seaking.validateToken"] = h.validateToken
	h.methods["seaking.getUserInfo"] = h.getUserInfo
}

// Handle 处理RPC请求
func (h *Handler) Handle(c *gin.Context) {
	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, Response{
			JsonRPC: "2.0",
			Error:   &Error{Code: -32700, Message: "Parse error"},
			ID:      req.ID,
		})
		return
	}

	if req.JsonRPC != "2.0" {
		c.JSON(200, Response{
			JsonRPC: "2.0",
			Error:   &Error{Code: -32600, Message: "Invalid Request"},
			ID:      req.ID,
		})
		return
	}

	method, ok := h.methods[req.Method]
	if !ok {
		c.JSON(200, Response{
			JsonRPC: "2.0",
			Error:   &Error{Code: -32601, Message: "Method not found"},
			ID:      req.ID,
		})
		return
	}

	result, err := method(c.Request.Context(), req.Params)
	if err != nil {
		c.JSON(200, Response{
			JsonRPC: "2.0",
			Error:   &Error{Code: -32000, Message: err.Error()},
			ID:      req.ID,
		})
		return
	}

	c.JSON(200, Response{
		JsonRPC: "2.0",
		Result:  result,
		ID:      req.ID,
	})
}

// checkAccess 检查访问权限
func (h *Handler) checkAccess(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Uid string `json:"uid"`
		Cid string `json:"cid"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	hasAccess, role, muted, err := h.convService.CheckAccess(ctx, req.Uid, req.Cid)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"has_access": hasAccess,
		"role":       role,
		"muted":      muted,
	}, nil
}

// getConversation 获取会话信息
func (h *Handler) getConversation(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Cid string `json:"cid"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	conv, err := h.convService.GetConversation(ctx, req.Cid)
	if err != nil {
		return nil, err
	}

	memberIds, _ := h.convService.GetConversationMemberIds(ctx, req.Cid)

	return map[string]interface{}{
		"cid":        conv.ID,
		"type":       conv.Type,
		"name":       conv.Name,
		"avatar":     conv.Avatar,
		"member_ids": memberIds,
	}, nil
}

// getConversationMembers 获取会话成员
func (h *Handler) getConversationMembers(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Cid string `json:"cid"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	members, err := h.convService.GetConversationMembers(ctx, req.Cid)
	if err != nil {
		return nil, err
	}

	// 获取用户详细信息
	var memberInfos []map[string]interface{}
	for _, m := range members {
		u, err := h.userService.GetByID(ctx, m.UserID)
		if err != nil {
			continue
		}
		memberInfos = append(memberInfos, map[string]interface{}{
			"uid":      u.ID,
			"nickname": u.Nickname,
			"avatar":   u.Avatar,
		})
	}

	return map[string]interface{}{
		"members": memberInfos,
	}, nil
}

// createConversation 创建会话
func (h *Handler) createConversation(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req conversation.CreateConversationRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	conv, err := h.convService.CreateConversation(ctx, &req)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"cid": conv.ID,
	}, nil
}

// getUserConversations 获取用户会话列表
func (h *Handler) getUserConversations(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Uid string `json:"uid"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	convs, err := h.convService.GetUserConversations(ctx, req.Uid)
	if err != nil {
		return nil, err
	}

	var convInfos []map[string]interface{}
	for _, c := range convs {
		memberIds, _ := h.convService.GetConversationMemberIds(ctx, c.ID)
		convInfos = append(convInfos, map[string]interface{}{
			"cid":        c.ID,
			"type":       c.Type,
			"name":       c.Name,
			"avatar":     c.Avatar,
			"member_ids": memberIds,
		})
	}

	return map[string]interface{}{
		"conversations": convInfos,
	}, nil
}

// validateToken 验证Token
func (h *Handler) validateToken(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	claims, err := h.jwtManager.ParseToken(req.Token)
	if err != nil {
		return map[string]interface{}{
			"valid": false,
		}, nil
	}

	return map[string]interface{}{
		"valid":     true,
		"uid":       claims.Uid,
		"device_id": claims.DeviceId,
		"platform":  claims.Platform,
	}, nil
}

// getUserInfo 获取用户信息
func (h *Handler) getUserInfo(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Uid string `json:"uid"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	u, err := h.userService.GetByID(ctx, req.Uid)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"uid":      u.ID,
		"username": u.Username,
		"nickname": u.Nickname,
		"avatar":   u.Avatar,
		"status":   u.Status,
	}, nil
}
