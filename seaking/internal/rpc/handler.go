package rpc

import (
	"context"
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/my-chat/common/pkg/auth"
	"github.com/my-chat/seaking/internal/service/conversation"
	"github.com/my-chat/seaking/internal/service/group"
	"github.com/my-chat/seaking/internal/service/key"
	"github.com/my-chat/seaking/internal/service/relation"
	"github.com/my-chat/seaking/internal/service/user"
)

// Handler RPC处理器
type Handler struct {
	userService     *user.Service
	convService     *conversation.Service
	relationService *relation.Service
	groupService    *group.Service
	keyService      *key.Service
	jwtManager      *auth.JWTManager
	methods         map[string]MethodHandler
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
func NewHandler(userService *user.Service, convService *conversation.Service, relationService *relation.Service, groupService *group.Service, keyService *key.Service, jwtManager *auth.JWTManager) *Handler {
	h := &Handler{
		userService:     userService,
		convService:     convService,
		relationService: relationService,
		groupService:    groupService,
		keyService:      keyService,
		jwtManager:      jwtManager,
		methods:         make(map[string]MethodHandler),
	}
	h.registerMethods()
	return h
}

// registerMethods 注册所有方法
func (h *Handler) registerMethods() {
	// 用户相关
	h.methods["seaking.register"] = h.register
	h.methods["seaking.login"] = h.login
	h.methods["seaking.validateToken"] = h.validateToken
	h.methods["seaking.getUserInfo"] = h.getUserInfo

	// 会话相关
	h.methods["seaking.checkAccess"] = h.checkAccess
	h.methods["seaking.getConversation"] = h.getConversation
	h.methods["seaking.getConversationMembers"] = h.getConversationMembers
	h.methods["seaking.createConversation"] = h.createConversation
	h.methods["seaking.getUserConversations"] = h.getUserConversations

	// 好友相关
	h.methods["seaking.getFriends"] = h.getFriends
	h.methods["seaking.sendFriendRequest"] = h.sendFriendRequest
	h.methods["seaking.getPendingFriendRequests"] = h.getPendingFriendRequests
	h.methods["seaking.acceptFriendRequest"] = h.acceptFriendRequest
	h.methods["seaking.rejectFriendRequest"] = h.rejectFriendRequest
	h.methods["seaking.deleteFriend"] = h.deleteFriend

	// 群组相关
	h.methods["seaking.getUserGroups"] = h.getUserGroups
	h.methods["seaking.createGroup"] = h.createGroup
	h.methods["seaking.getGroupInfo"] = h.getGroupInfo
	h.methods["seaking.getGroupMembers"] = h.getGroupMembers

	// 加密密钥相关
	h.methods["seaking.getUserPublicKey"] = h.getUserPublicKey
	h.methods["seaking.getMemberPublicKeys"] = h.getMemberPublicKeys
	h.methods["seaking.getChatKey"] = h.getChatKey
	h.methods["seaking.createChatKey"] = h.createChatKey
	h.methods["seaking.getGroupKey"] = h.getGroupKey
	h.methods["seaking.createGroupKey"] = h.createGroupKey
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

// register 用户注册
func (h *Handler) register(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Username            string `json:"username"`
		Password            string `json:"password"`
		Nickname            string `json:"nickname"`
		Phone               string `json:"phone"`
		Email               string `json:"email"`
		PublicKey           string `json:"public_key"`
		EncryptedPrivateKey string `json:"encrypted_private_key"`
		KeySalt             string `json:"key_salt"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	u, err := h.userService.Register(ctx, &user.RegisterRequest{
		Username: req.Username,
		Password: req.Password,
		Nickname: req.Nickname,
		Phone:    req.Phone,
		Email:    req.Email,
	})
	if err != nil {
		return nil, err
	}

	// 如果提供了密钥，则存储用户密钥
	if req.PublicKey != "" && req.EncryptedPrivateKey != "" && req.KeySalt != "" {
		if err := h.keyService.CreateUserKey(ctx, &key.CreateUserKeyRequest{
			UserID:              u.ID,
			PublicKey:           req.PublicKey,
			EncryptedPrivateKey: req.EncryptedPrivateKey,
			KeySalt:             req.KeySalt,
		}); err != nil {
			// 密钥存储失败，但用户已创建，记录错误但不影响注册
			// 用户可以后续重新上传密钥
		}
	}

	return map[string]interface{}{
		"uid":      u.ID,
		"username": u.Username,
		"nickname": u.Nickname,
	}, nil
}

// login 用户登录
func (h *Handler) login(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		DeviceId string `json:"device_id"`
		Platform string `json:"platform"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	u, err := h.userService.Login(ctx, &user.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}

	token, err := h.jwtManager.GenerateToken(u.ID, req.DeviceId, req.Platform)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"uid":      u.ID,
			"username": u.Username,
			"nickname": u.Nickname,
			"avatar":   u.Avatar,
			"status":   u.Status,
		},
	}

	// 获取用户密钥信息（用于新设备解密私钥）
	userKey, err := h.keyService.GetUserKey(ctx, u.ID)
	if err == nil && userKey != nil {
		result["encryption"] = map[string]interface{}{
			"public_key":            userKey.PublicKey,
			"encrypted_private_key": userKey.EncryptedPrivateKey,
			"key_salt":              userKey.KeySalt,
		}
	}

	return result, nil
}

// getFriends 获取好友列表
func (h *Handler) getFriends(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Uid string `json:"uid"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	friends, err := h.relationService.GetFriends(ctx, req.Uid)
	if err != nil {
		return nil, err
	}

	var friendInfos []map[string]interface{}
	for _, f := range friends {
		u, err := h.userService.GetByID(ctx, f.FriendID)
		if err != nil {
			continue
		}
		friendInfos = append(friendInfos, map[string]interface{}{
			"uid":      u.ID,
			"username": u.Username,
			"nickname": u.Nickname,
			"avatar":   u.Avatar,
			"remark":   f.Remark,
		})
	}

	return map[string]interface{}{
		"friends": friendInfos,
	}, nil
}

// sendFriendRequest 发送好友请求
func (h *Handler) sendFriendRequest(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		FromUid string `json:"from_uid"`
		ToUid   string `json:"to_uid"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	if err := h.relationService.SendFriendRequest(ctx, req.FromUid, req.ToUid, req.Message); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"success": true,
	}, nil
}

// getPendingFriendRequests 获取待处理的好友请求
func (h *Handler) getPendingFriendRequests(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Uid string `json:"uid"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	requests, err := h.relationService.GetPendingRequests(ctx, req.Uid)
	if err != nil {
		return nil, err
	}

	var requestInfos []map[string]interface{}
	for _, r := range requests {
		u, _ := h.userService.GetByID(ctx, r.FromUID)
		requestInfos = append(requestInfos, map[string]interface{}{
			"request_id": r.ID,
			"from_uid":   r.FromUID,
			"from_name":  u.Nickname,
			"message":    r.Message,
			"status":     r.Status,
		})
	}

	return map[string]interface{}{
		"requests": requestInfos,
	}, nil
}

// acceptFriendRequest 接受好友请求
func (h *Handler) acceptFriendRequest(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		RequestId uint   `json:"request_id"`
		Uid       string `json:"uid"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	if err := h.relationService.AcceptFriendRequest(ctx, req.RequestId, req.Uid); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"success": true,
	}, nil
}

// rejectFriendRequest 拒绝好友请求
func (h *Handler) rejectFriendRequest(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		RequestId uint   `json:"request_id"`
		Uid       string `json:"uid"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	if err := h.relationService.RejectFriendRequest(ctx, req.RequestId, req.Uid); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"success": true,
	}, nil
}

// deleteFriend 删除好友
func (h *Handler) deleteFriend(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Uid      string `json:"uid"`
		FriendId string `json:"friend_id"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	if err := h.relationService.DeleteFriend(ctx, req.Uid, req.FriendId); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"success": true,
	}, nil
}

// getUserGroups 获取用户群组列表
func (h *Handler) getUserGroups(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Uid string `json:"uid"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	groups, err := h.groupService.GetUserGroups(ctx, req.Uid)
	if err != nil {
		return nil, err
	}

	var groupInfos []map[string]interface{}
	for _, g := range groups {
		groupInfos = append(groupInfos, map[string]interface{}{
			"id":          g.ID,
			"name":        g.Name,
			"description": g.Description,
			"avatar":      g.Avatar,
			"owner_id":    g.OwnerID,
			"max_members": g.MaxMembers,
		})
	}

	return map[string]interface{}{
		"groups": groupInfos,
	}, nil
}

// createGroup 创建群组
func (h *Handler) createGroup(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		OwnerId     string   `json:"owner_id"`
		Name        string   `json:"name"`
		Description string   `json:"description"`
		MemberIds   []string `json:"member_ids"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	g, err := h.groupService.CreateGroup(ctx, req.OwnerId, &group.CreateGroupRequest{
		Name:        req.Name,
		Description: req.Description,
		MemberIDs:   req.MemberIds,
	})
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":          g.ID,
		"name":        g.Name,
		"description": g.Description,
		"avatar":      g.Avatar,
		"owner_id":    g.OwnerID,
		"max_members": g.MaxMembers,
	}, nil
}

// getGroupInfo 获取群组信息
func (h *Handler) getGroupInfo(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		GroupId string `json:"group_id"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	g, err := h.groupService.GetGroup(ctx, req.GroupId)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":          g.ID,
		"name":        g.Name,
		"description": g.Description,
		"avatar":      g.Avatar,
		"owner_id":    g.OwnerID,
		"max_members": g.MaxMembers,
	}, nil
}

// getGroupMembers 获取群组成员
func (h *Handler) getGroupMembers(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		GroupId string `json:"group_id"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	members, err := h.groupService.GetMembers(ctx, req.GroupId)
	if err != nil {
		return nil, err
	}

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
			"role":     m.Role,
		})
	}

	return map[string]interface{}{
		"members": memberInfos,
	}, nil
}

// ==================== 加密密钥相关 ====================

// getUserPublicKey 获取用户公钥
func (h *Handler) getUserPublicKey(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Uid string `json:"uid"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	publicKey, err := h.keyService.GetUserPublicKey(ctx, req.Uid)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"uid":        req.Uid,
		"public_key": publicKey,
	}, nil
}

// getMemberPublicKeys 批量获取成员公钥
func (h *Handler) getMemberPublicKeys(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Uids []string `json:"uids"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	publicKeys, err := h.keyService.GetUserPublicKeys(ctx, req.Uids)
	if err != nil {
		return nil, err
	}

	var keys []map[string]string
	for uid, pk := range publicKeys {
		keys = append(keys, map[string]string{
			"uid":        uid,
			"public_key": pk,
		})
	}

	return map[string]interface{}{
		"keys": keys,
	}, nil
}

// getChatKey 获取私聊会话密钥
func (h *Handler) getChatKey(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Cid string `json:"cid"`
		Uid string `json:"uid"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	chatKey, err := h.keyService.GetChatKey(ctx, req.Cid, req.Uid)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"cid":           chatKey.ConversationID,
		"encrypted_key": chatKey.EncryptedKey,
	}, nil
}

// createChatKey 创建私聊会话密钥
func (h *Handler) createChatKey(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Cid  string `json:"cid"`
		Keys []struct {
			Uid          string `json:"uid"`
			EncryptedKey string `json:"encrypted_key"`
		} `json:"keys"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	// 检查密钥是否已存在
	exists, err := h.keyService.ChatKeyExists(ctx, req.Cid)
	if err != nil {
		return nil, err
	}
	if exists {
		return map[string]interface{}{
			"success": true,
			"message": "chat key already exists",
		}, nil
	}

	var keyEntries []key.ChatKeyEntry
	for _, k := range req.Keys {
		keyEntries = append(keyEntries, key.ChatKeyEntry{
			UserID:       k.Uid,
			EncryptedKey: k.EncryptedKey,
		})
	}

	if err := h.keyService.CreateChatKeys(ctx, &key.CreateChatKeysRequest{
		ConversationID: req.Cid,
		Keys:           keyEntries,
	}); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"success": true,
	}, nil
}

// getGroupKey 获取群组密钥
func (h *Handler) getGroupKey(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		GroupId string `json:"group_id"`
		Uid     string `json:"uid"`
		Version int    `json:"version"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	groupKey, err := h.keyService.GetGroupKey(ctx, req.GroupId, req.Uid, req.Version)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"group_id":      groupKey.GroupID,
		"encrypted_key": groupKey.EncryptedKey,
		"version":       groupKey.Version,
	}, nil
}

// createGroupKey 创建群组密钥
func (h *Handler) createGroupKey(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		GroupId string `json:"group_id"`
		Keys    []struct {
			Uid          string `json:"uid"`
			EncryptedKey string `json:"encrypted_key"`
		} `json:"keys"`
		Version int `json:"version"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	// 如果没有指定版本，获取最新版本并+1
	version := req.Version
	if version == 0 {
		latestVersion, err := h.keyService.GetLatestGroupKeyVersion(ctx, req.GroupId)
		if err != nil {
			return nil, err
		}
		version = latestVersion + 1
	}

	var keyEntries []key.GroupKeyEntry
	for _, k := range req.Keys {
		keyEntries = append(keyEntries, key.GroupKeyEntry{
			UserID:       k.Uid,
			EncryptedKey: k.EncryptedKey,
		})
	}

	if err := h.keyService.CreateGroupKeys(ctx, &key.CreateGroupKeysRequest{
		GroupID: req.GroupId,
		Keys:    keyEntries,
		Version: version,
	}); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"success": true,
		"version": version,
	}, nil
}
