package rpc

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/my-chat/common/pkg/auth"
	"github.com/my-chat/common/pkg/client"
	"github.com/my-chat/common/pkg/log"
)

// Handler Gateway RPC处理器
type Handler struct {
	jwtManager    *auth.JWTManager
	seakingClient *client.SeaKingClient
	relayClient   *client.RelayClient
	methods       map[string]MethodHandler
}

// MethodHandler 方法处理函数
type MethodHandler func(ctx *gin.Context, id any, params json.RawMessage) any

// RPCRequest JSON-RPC请求
type RPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	ID      any             `json:"id"`
}

// RPCResponse JSON-RPC响应
type RPCResponse struct {
	JSONRPC string    `json:"jsonrpc"`
	Result  any       `json:"result,omitempty"`
	Error   *RPCError `json:"error,omitempty"`
	ID      any       `json:"id"`
}

// RPCError JSON-RPC错误
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// NewHandler 创建Handler
func NewHandler(jwtManager *auth.JWTManager, seakingAddr, relayAddr string) *Handler {
	h := &Handler{
		jwtManager:    jwtManager,
		seakingClient: client.NewSeaKingClient(seakingAddr),
		relayClient:   client.NewRelayClient(relayAddr),
		methods:       make(map[string]MethodHandler),
	}

	// 注册方法
	h.registerMethods()

	return h
}

// registerMethods 注册所有RPC方法
func (h *Handler) registerMethods() {
	// 认证相关（无需token）
	h.methods["register"] = h.register
	h.methods["login"] = h.login

	// 用户相关（需要token）
	h.methods["getUserInfo"] = h.withAuth(h.getUserInfo)

	// 好友相关（需要token）
	h.methods["getFriends"] = h.withAuth(h.getFriends)
	h.methods["sendFriendRequest"] = h.withAuth(h.sendFriendRequest)
	h.methods["getPendingFriendRequests"] = h.withAuth(h.getPendingFriendRequests)
	h.methods["acceptFriendRequest"] = h.withAuth(h.acceptFriendRequest)
	h.methods["rejectFriendRequest"] = h.withAuth(h.rejectFriendRequest)
	h.methods["deleteFriend"] = h.withAuth(h.deleteFriend)

	// 会话相关（需要token）
	h.methods["getConversations"] = h.withAuth(h.getConversations)
	h.methods["createConversation"] = h.withAuth(h.createConversation)
	h.methods["getConversationMembers"] = h.withAuth(h.getConversationMembers)

	// 群组相关（需要token）
	h.methods["getGroups"] = h.withAuth(h.getGroups)
	h.methods["createGroup"] = h.withAuth(h.createGroup)
	h.methods["getGroupInfo"] = h.withAuth(h.getGroupInfo)
	h.methods["getGroupMembers"] = h.withAuth(h.getGroupMembers)
}

// Handle 处理RPC请求
func (h *Handler) Handle(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, RPCResponse{
			JSONRPC: "2.0",
			Error:   &RPCError{Code: -32700, Message: "Parse error"},
			ID:      nil,
		})
		return
	}

	var req RPCRequest
	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusOK, RPCResponse{
			JSONRPC: "2.0",
			Error:   &RPCError{Code: -32700, Message: "Parse error"},
			ID:      nil,
		})
		return
	}

	if req.JSONRPC != "2.0" {
		c.JSON(http.StatusOK, RPCResponse{
			JSONRPC: "2.0",
			Error:   &RPCError{Code: -32600, Message: "Invalid Request"},
			ID:      req.ID,
		})
		return
	}

	handler, ok := h.methods[req.Method]
	if !ok {
		c.JSON(http.StatusOK, RPCResponse{
			JSONRPC: "2.0",
			Error:   &RPCError{Code: -32601, Message: "Method not found"},
			ID:      req.ID,
		})
		return
	}

	result := handler(c, req.ID, req.Params)
	if rpcErr, ok := result.(*RPCError); ok {
		c.JSON(http.StatusOK, RPCResponse{
			JSONRPC: "2.0",
			Error:   rpcErr,
			ID:      req.ID,
		})
		return
	}

	c.JSON(http.StatusOK, RPCResponse{
		JSONRPC: "2.0",
		Result:  result,
		ID:      req.ID,
	})
}

// withAuth 包装需要认证的方法
func (h *Handler) withAuth(fn func(ctx *gin.Context, uid string, id any, params json.RawMessage) any) MethodHandler {
	return func(ctx *gin.Context, id any, params json.RawMessage) any {
		token := ctx.GetHeader("Authorization")
		if token == "" {
			return &RPCError{Code: -32001, Message: "Authorization required"}
		}

		// 移除 "Bearer " 前缀
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		claims, err := h.jwtManager.ParseToken(token)
		if err != nil {
			return &RPCError{Code: -32002, Message: "Invalid token"}
		}

		return fn(ctx, claims.Uid, id, params)
	}
}

// ============== 认证相关 ==============

func (h *Handler) register(ctx *gin.Context, id any, params json.RawMessage) any {
	var req client.RegisterRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return &RPCError{Code: -32602, Message: "Invalid params"}
	}

	resp, err := h.seakingClient.Register(ctx.Request.Context(), &req)
	if err != nil {
		log.Error().Err(err).Msg("register failed")
		return &RPCError{Code: -32000, Message: err.Error()}
	}

	return resp
}

func (h *Handler) login(ctx *gin.Context, id any, params json.RawMessage) any {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		DeviceId string `json:"device_id"`
		Platform string `json:"platform"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return &RPCError{Code: -32602, Message: "Invalid params"}
	}

	resp, err := h.seakingClient.Login(ctx.Request.Context(), &client.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	}, req.DeviceId, req.Platform)
	if err != nil {
		log.Error().Err(err).Msg("login failed")
		return &RPCError{Code: -32000, Message: err.Error()}
	}

	return resp
}

// ============== 用户相关 ==============

func (h *Handler) getUserInfo(ctx *gin.Context, uid string, id any, params json.RawMessage) any {
	var req struct {
		Uid string `json:"uid"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return &RPCError{Code: -32602, Message: "Invalid params"}
	}

	targetUid := req.Uid
	if targetUid == "" {
		targetUid = uid
	}

	resp, err := h.seakingClient.GetUserInfo(ctx.Request.Context(), targetUid)
	if err != nil {
		log.Error().Err(err).Msg("getUserInfo failed")
		return &RPCError{Code: -32000, Message: err.Error()}
	}

	return resp
}

// ============== 好友相关 ==============

func (h *Handler) getFriends(ctx *gin.Context, uid string, id any, params json.RawMessage) any {
	friends, err := h.seakingClient.GetFriends(ctx.Request.Context(), uid)
	if err != nil {
		log.Error().Err(err).Msg("getFriends failed")
		return &RPCError{Code: -32000, Message: err.Error()}
	}

	return map[string]any{"friends": friends}
}

func (h *Handler) sendFriendRequest(ctx *gin.Context, uid string, id any, params json.RawMessage) any {
	var req struct {
		ToUid   string `json:"to_uid"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return &RPCError{Code: -32602, Message: "Invalid params"}
	}

	err := h.seakingClient.SendFriendRequest(ctx.Request.Context(), uid, req.ToUid, req.Message)
	if err != nil {
		log.Error().Err(err).Msg("sendFriendRequest failed")
		return &RPCError{Code: -32000, Message: err.Error()}
	}

	return map[string]any{"success": true}
}

func (h *Handler) getPendingFriendRequests(ctx *gin.Context, uid string, id any, params json.RawMessage) any {
	requests, err := h.seakingClient.GetPendingFriendRequests(ctx.Request.Context(), uid)
	if err != nil {
		log.Error().Err(err).Msg("getPendingFriendRequests failed")
		return &RPCError{Code: -32000, Message: err.Error()}
	}

	return map[string]any{"requests": requests}
}

func (h *Handler) acceptFriendRequest(ctx *gin.Context, uid string, id any, params json.RawMessage) any {
	var req struct {
		RequestId uint `json:"request_id"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return &RPCError{Code: -32602, Message: "Invalid params"}
	}

	err := h.seakingClient.AcceptFriendRequest(ctx.Request.Context(), req.RequestId, uid)
	if err != nil {
		log.Error().Err(err).Msg("acceptFriendRequest failed")
		return &RPCError{Code: -32000, Message: err.Error()}
	}

	return map[string]any{"success": true}
}

func (h *Handler) rejectFriendRequest(ctx *gin.Context, uid string, id any, params json.RawMessage) any {
	var req struct {
		RequestId uint `json:"request_id"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return &RPCError{Code: -32602, Message: "Invalid params"}
	}

	err := h.seakingClient.RejectFriendRequest(ctx.Request.Context(), req.RequestId, uid)
	if err != nil {
		log.Error().Err(err).Msg("rejectFriendRequest failed")
		return &RPCError{Code: -32000, Message: err.Error()}
	}

	return map[string]any{"success": true}
}

func (h *Handler) deleteFriend(ctx *gin.Context, uid string, id any, params json.RawMessage) any {
	var req struct {
		FriendId string `json:"friend_id"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return &RPCError{Code: -32602, Message: "Invalid params"}
	}

	err := h.seakingClient.DeleteFriend(ctx.Request.Context(), uid, req.FriendId)
	if err != nil {
		log.Error().Err(err).Msg("deleteFriend failed")
		return &RPCError{Code: -32000, Message: err.Error()}
	}

	return map[string]any{"success": true}
}

// ============== 会话相关 ==============

func (h *Handler) getConversations(ctx *gin.Context, uid string, id any, params json.RawMessage) any {
	resp, err := h.seakingClient.GetUserConversations(ctx.Request.Context(), uid)
	if err != nil {
		log.Error().Err(err).Msg("getConversations failed")
		return &RPCError{Code: -32000, Message: err.Error()}
	}

	return map[string]any{"conversations": resp.Conversations}
}

func (h *Handler) createConversation(ctx *gin.Context, uid string, id any, params json.RawMessage) any {
	var req struct {
		Type      int      `json:"type"`
		MemberIds []string `json:"member_ids"`
		Name      string   `json:"name"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return &RPCError{Code: -32602, Message: "Invalid params"}
	}

	resp, err := h.seakingClient.CreateConversation(ctx.Request.Context(), req.Type, uid, req.MemberIds, req.Name)
	if err != nil {
		log.Error().Err(err).Msg("createConversation failed")
		return &RPCError{Code: -32000, Message: err.Error()}
	}

	return map[string]any{"cid": resp.Cid}
}

func (h *Handler) getConversationMembers(ctx *gin.Context, uid string, id any, params json.RawMessage) any {
	var req struct {
		Cid string `json:"cid"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return &RPCError{Code: -32602, Message: "Invalid params"}
	}

	// 检查权限
	accessResp, err := h.seakingClient.CheckAccess(ctx.Request.Context(), uid, req.Cid)
	if err != nil {
		log.Error().Err(err).Msg("checkAccess failed")
		return &RPCError{Code: -32000, Message: err.Error()}
	}
	if !accessResp.HasAccess {
		return &RPCError{Code: -32003, Message: "Access denied"}
	}

	resp, err := h.seakingClient.GetConversationMembers(ctx.Request.Context(), req.Cid)
	if err != nil {
		log.Error().Err(err).Msg("getConversationMembers failed")
		return &RPCError{Code: -32000, Message: err.Error()}
	}

	return map[string]any{"members": resp.Members}
}

// ============== 群组相关 ==============

func (h *Handler) getGroups(ctx *gin.Context, uid string, id any, params json.RawMessage) any {
	groups, err := h.seakingClient.GetUserGroups(ctx.Request.Context(), uid)
	if err != nil {
		log.Error().Err(err).Msg("getGroups failed")
		return &RPCError{Code: -32000, Message: err.Error()}
	}

	return map[string]any{"groups": groups}
}

func (h *Handler) createGroup(ctx *gin.Context, uid string, id any, params json.RawMessage) any {
	var req struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		MemberIds   []string `json:"member_ids"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return &RPCError{Code: -32602, Message: "Invalid params"}
	}

	group, err := h.seakingClient.CreateGroup(ctx.Request.Context(), uid, req.Name, req.Description, req.MemberIds)
	if err != nil {
		log.Error().Err(err).Msg("createGroup failed")
		return &RPCError{Code: -32000, Message: err.Error()}
	}

	return map[string]any{"group": group}
}

func (h *Handler) getGroupInfo(ctx *gin.Context, uid string, id any, params json.RawMessage) any {
	var req struct {
		GroupId string `json:"group_id"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return &RPCError{Code: -32602, Message: "Invalid params"}
	}

	group, err := h.seakingClient.GetGroupInfo(ctx.Request.Context(), req.GroupId)
	if err != nil {
		log.Error().Err(err).Msg("getGroupInfo failed")
		return &RPCError{Code: -32000, Message: err.Error()}
	}

	return map[string]any{"group": group}
}

func (h *Handler) getGroupMembers(ctx *gin.Context, uid string, id any, params json.RawMessage) any {
	var req struct {
		GroupId string `json:"group_id"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return &RPCError{Code: -32602, Message: "Invalid params"}
	}

	members, err := h.seakingClient.GetGroupMembers(ctx.Request.Context(), req.GroupId)
	if err != nil {
		log.Error().Err(err).Msg("getGroupMembers failed")
		return &RPCError{Code: -32000, Message: err.Error()}
	}

	return map[string]any{"members": members}
}
