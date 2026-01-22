package handler

import (
	"context"
	"encoding/json"
	"time"

	"github.com/my-chat/common/pkg/auth"
	"github.com/my-chat/common/pkg/client"
	"github.com/my-chat/common/pkg/errors"
	"github.com/my-chat/common/pkg/log"
	"github.com/my-chat/common/pkg/protocol"
	"github.com/my-chat/gateway/internal/ws"
)

// Handler 消息处理器
type Handler struct {
	hub           *ws.Hub
	jwtManager    *auth.JWTManager
	relayClient   *client.RelayClient
	seakingClient *client.SeaKingClient
}

// NewHandler 创建处理器
func NewHandler(hub *ws.Hub, jwtManager *auth.JWTManager, relayAddr, seakingAddr string) *Handler {
	return &Handler{
		hub:           hub,
		jwtManager:    jwtManager,
		relayClient:   client.NewRelayClient(relayAddr),
		seakingClient: client.NewSeaKingClient(seakingAddr),
	}
}

// HandleMessage 处理消息
func (h *Handler) HandleMessage(conn *ws.Conn, data []byte) {
	env, err := protocol.DecodeEnvelope(data)
	if err != nil {
		h.sendError(conn, 0, errors.ErrInvalidParam)
		return
	}

	switch env.Cmd {
	case protocol.CmdPing:
		h.handlePing(conn, env)

	case protocol.CmdEvent:
		h.handleEvent(conn, env)

	case protocol.CmdSubscribe:
		h.handleSubscribe(conn, env)

	case protocol.CmdUnsubscribe:
		h.handleUnsubscribe(conn, env)

	case protocol.CmdSync:
		h.handleSync(conn, env)

	// 好友相关命令
	case protocol.CmdGetFriends:
		h.handleGetFriends(conn, env)

	case protocol.CmdSendFriendRequest:
		h.handleSendFriendRequest(conn, env)

	case protocol.CmdGetFriendRequests:
		h.handleGetFriendRequests(conn, env)

	case protocol.CmdAcceptFriendRequest:
		h.handleAcceptFriendRequest(conn, env)

	case protocol.CmdRejectFriendRequest:
		h.handleRejectFriendRequest(conn, env)

	case protocol.CmdDeleteFriend:
		h.handleDeleteFriend(conn, env)

	// 会话相关命令
	case protocol.CmdGetConversations:
		h.handleGetConversations(conn, env)

	case protocol.CmdCreateConversation:
		h.handleCreateConversation(conn, env)

	case protocol.CmdGetConversationMembers:
		h.handleGetConversationMembers(conn, env)

	// 群组相关命令
	case protocol.CmdGetGroups:
		h.handleGetGroups(conn, env)

	case protocol.CmdCreateGroup:
		h.handleCreateGroup(conn, env)

	case protocol.CmdGetGroupInfo:
		h.handleGetGroupInfo(conn, env)

	case protocol.CmdGetGroupMembers:
		h.handleGetGroupMembers(conn, env)

	// 用户相关命令
	case protocol.CmdGetUserInfo:
		h.handleGetUserInfo(conn, env)

	default:
		h.sendError(conn, env.Seq, errors.New(errors.ErrCodeInvalidParam, "unknown command"))
	}
}

// handlePing 处理心跳
func (h *Handler) handlePing(conn *ws.Conn, env *protocol.Envelope) {
	pong := protocol.NewEnvelope(protocol.CmdPong, env.Seq, nil)
	conn.SendEnvelope(pong)
}

// handleEvent 处理事件消息
func (h *Handler) handleEvent(conn *ws.Conn, env *protocol.Envelope) {
	event, err := protocol.DecodeEventFromBody(env.Body)
	if err != nil {
		h.sendError(conn, env.Seq, errors.ErrInvalidParam)
		return
	}

	// 设置发送者
	event.Sender = conn.UID()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 检查用户是否有权发送消息到该会话
	accessResp, err := h.seakingClient.CheckAccess(ctx, conn.UID(), event.Cid)
	if err != nil {
		log.Error().Err(err).Msg("failed to check access")
		h.sendError(conn, env.Seq, errors.ErrInternal)
		return
	}

	if !accessResp.HasAccess {
		h.sendError(conn, env.Seq, errors.ErrNotInConversation)
		return
	}

	// 检查是否被禁言
	if accessResp.Muted && event.Kind != protocol.KindReadReceipt {
		h.sendError(conn, env.Seq, errors.New(errors.ErrCodeForbidden, "you are muted"))
		return
	}

	// 根据消息类型处理
	switch event.Kind {
	case protocol.KindTyping:
		// Typing消息不持久化，直接转发
		h.broadcastEvent(event)
		h.sendAck(conn, env.Seq, 0)

	case protocol.KindRevoke:
		// 撤销消息需要验证权限
		h.handleRevokeEvent(ctx, conn, env, event, accessResp.Role >= 1)

	case protocol.KindEdit:
		// 编辑消息需要验证权限
		h.handleEditEvent(ctx, conn, env, event)

	case protocol.KindReadReceipt:
		// 已读回执直接更新
		h.handleReadReceiptEvent(ctx, conn, env, event)

	default:
		// 其他消息需要持久化
		h.handlePersistentEvent(ctx, conn, env, event)
	}

	log.Debug().
		Str("uid", conn.UID()).
		Int("kind", event.Kind).
		Str("cid", event.Cid).
		Msg("event processed")
}

// handlePersistentEvent 处理需要持久化的事件
func (h *Handler) handlePersistentEvent(ctx context.Context, conn *ws.Conn, env *protocol.Envelope, event *protocol.Event) {
	// 存储到Relay
	resp, err := h.relayClient.StoreEvent(ctx, event)
	if err != nil {
		log.Error().Err(err).Msg("failed to store event")
		h.sendError(conn, env.Seq, errors.ErrInternal)
		return
	}

	// 更新事件的mid和时间戳
	event.Mid = resp.Mid
	event.Timestamp = resp.Timestamp

	// 广播给会话中的其他用户
	h.broadcastEvent(event)

	// 发送确认
	h.sendAck(conn, env.Seq, resp.Mid)
}

// handleRevokeEvent 处理撤销事件
func (h *Handler) handleRevokeEvent(ctx context.Context, conn *ws.Conn, env *protocol.Envelope, event *protocol.Event, isAdmin bool) {
	// 获取目标消息ID
	targetMid, ok := protocol.GetTargetMid(event.Tags)
	if !ok {
		h.sendError(conn, env.Seq, errors.ErrInvalidParam)
		return
	}

	// 验证撤销权限
	validateResp, err := h.relayClient.ValidateRevoke(ctx, event.Cid, conn.UID(), targetMid, isAdmin)
	if err != nil {
		log.Error().Err(err).Msg("failed to validate revoke")
		h.sendError(conn, env.Seq, errors.ErrInternal)
		return
	}

	if !validateResp.Valid {
		h.sendError(conn, env.Seq, errors.New(errors.ErrCodeCannotRevoke, validateResp.Reason))
		return
	}

	// 存储撤销事件
	h.handlePersistentEvent(ctx, conn, env, event)
}

// handleEditEvent 处理编辑事件
func (h *Handler) handleEditEvent(ctx context.Context, conn *ws.Conn, env *protocol.Envelope, event *protocol.Event) {
	// 获取目标消息ID
	targetMid, ok := protocol.GetTargetMid(event.Tags)
	if !ok {
		h.sendError(conn, env.Seq, errors.ErrInvalidParam)
		return
	}

	// 验证编辑权限
	validateResp, err := h.relayClient.ValidateEdit(ctx, event.Cid, conn.UID(), targetMid)
	if err != nil {
		log.Error().Err(err).Msg("failed to validate edit")
		h.sendError(conn, env.Seq, errors.ErrInternal)
		return
	}

	if !validateResp.Valid {
		h.sendError(conn, env.Seq, errors.New(errors.ErrCodeCannotEdit, validateResp.Reason))
		return
	}

	// 存储编辑事件
	h.handlePersistentEvent(ctx, conn, env, event)
}

// handleReadReceiptEvent 处理已读回执
func (h *Handler) handleReadReceiptEvent(ctx context.Context, conn *ws.Conn, env *protocol.Envelope, event *protocol.Event) {
	// 获取已读消息ID
	lastReadMid, ok := event.Data[0].(int64)
	if !ok {
		// 尝试float64转换（JSON解析可能产生float64）
		if f, ok := event.Data[0].(float64); ok {
			lastReadMid = int64(f)
		} else {
			h.sendError(conn, env.Seq, errors.ErrInvalidParam)
			return
		}
	}

	// 更新已读回执
	err := h.relayClient.UpdateReadReceipt(ctx, event.Cid, conn.UID(), lastReadMid)
	if err != nil {
		log.Error().Err(err).Msg("failed to update read receipt")
		h.sendError(conn, env.Seq, errors.ErrInternal)
		return
	}

	// 广播已读回执给其他用户
	h.broadcastEvent(event)
	h.sendAck(conn, env.Seq, 0)
}

// handleSubscribe 处理订阅
func (h *Handler) handleSubscribe(conn *ws.Conn, env *protocol.Envelope) {
	cid, ok := env.Body.(string)
	if !ok {
		h.sendError(conn, env.Seq, errors.ErrInvalidParam)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 检查用户是否有权限订阅该会话
	accessResp, err := h.seakingClient.CheckAccess(ctx, conn.UID(), cid)
	if err != nil {
		log.Error().Err(err).Msg("failed to check access")
		h.sendError(conn, env.Seq, errors.ErrInternal)
		return
	}

	if !accessResp.HasAccess {
		h.sendError(conn, env.Seq, errors.ErrNotInConversation)
		return
	}

	h.hub.Subscribe(conn, cid)
	h.sendAck(conn, env.Seq, 0)
}

// handleUnsubscribe 处理取消订阅
func (h *Handler) handleUnsubscribe(conn *ws.Conn, env *protocol.Envelope) {
	cid, ok := env.Body.(string)
	if !ok {
		h.sendError(conn, env.Seq, errors.ErrInvalidParam)
		return
	}

	h.hub.Unsubscribe(conn, cid)
	h.sendAck(conn, env.Seq, 0)
}

// handleSync 处理同步请求
func (h *Handler) handleSync(conn *ws.Conn, env *protocol.Envelope) {
	var syncBody protocol.SyncBody
	bodyData, _ := json.Marshal(env.Body)
	if err := json.Unmarshal(bodyData, &syncBody); err != nil {
		h.sendError(conn, env.Seq, errors.ErrInvalidParam)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 检查用户是否有权限访问该会话
	accessResp, err := h.seakingClient.CheckAccess(ctx, conn.UID(), syncBody.Cid)
	if err != nil {
		log.Error().Err(err).Msg("failed to check access")
		h.sendError(conn, env.Seq, errors.ErrInternal)
		return
	}

	if !accessResp.HasAccess {
		h.sendError(conn, env.Seq, errors.ErrNotInConversation)
		return
	}

	// 从Relay获取历史消息
	limit := syncBody.Limit
	if limit <= 0 {
		limit = 50
	}

	var events *client.QueryEventsResponse
	if syncBody.LastMid > 0 {
		// 增量同步
		events, err = h.relayClient.QueryEvents(ctx, &client.QueryEventsRequest{
			Cid:     syncBody.Cid,
			LastMid: syncBody.LastMid,
			Before:  syncBody.Before,
			After:   syncBody.After,
			Limit:   limit,
		})
	} else {
		// 全量同步（获取最新消息）
		events, err = h.relayClient.SyncEvents(ctx, syncBody.Cid, limit)
	}

	if err != nil {
		log.Error().Err(err).Msg("failed to query events")
		h.sendError(conn, env.Seq, errors.ErrInternal)
		return
	}

	// 发送同步结果
	syncResult := protocol.NewEnvelope(protocol.CmdEvent, env.Seq, map[string]interface{}{
		"cid":    syncBody.Cid,
		"events": events.Events,
	})
	conn.SendEnvelope(syncResult)
}

// broadcastEvent 广播事件
func (h *Handler) broadcastEvent(event *protocol.Event) {
	data, err := protocol.Encode(protocol.NewEnvelope(protocol.CmdEvent, 0, event))
	if err != nil {
		log.Error().Err(err).Msg("failed to encode event")
		return
	}

	h.hub.Broadcast(event.Cid, data)
}

// sendAck 发送确认
func (h *Handler) sendAck(conn *ws.Conn, seq int64, mid int64) {
	ack := protocol.NewEnvelope(protocol.CmdAck, seq, &protocol.AckBody{
		Seq: seq,
		Mid: mid,
	})
	conn.SendEnvelope(ack)
}

// sendError 发送错误
func (h *Handler) sendError(conn *ws.Conn, seq int64, err *errors.Error) {
	errEnv := protocol.NewEnvelope(protocol.CmdError, seq, &protocol.ErrorBody{
		Code:    err.Code,
		Message: err.Message,
		Seq:     seq,
	})
	conn.SendEnvelope(errEnv)
}

// sendResult 发送结果
func (h *Handler) sendResult(conn *ws.Conn, seq int64, data interface{}) {
	result := protocol.NewEnvelope(protocol.CmdResult, seq, data)
	conn.SendEnvelope(result)
}

// ============== 好友相关处理 ==============

// handleGetFriends 获取好友列表
func (h *Handler) handleGetFriends(conn *ws.Conn, env *protocol.Envelope) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	friends, err := h.seakingClient.GetFriends(ctx, conn.UID())
	if err != nil {
		log.Error().Err(err).Msg("failed to get friends")
		h.sendError(conn, env.Seq, errors.ErrInternal)
		return
	}

	h.sendResult(conn, env.Seq, map[string]interface{}{
		"friends": friends,
	})
}

// handleSendFriendRequest 发送好友请求
func (h *Handler) handleSendFriendRequest(conn *ws.Conn, env *protocol.Envelope) {
	var body struct {
		ToUid   string `json:"to_uid"`
		Message string `json:"message"`
	}
	bodyData, _ := json.Marshal(env.Body)
	if err := json.Unmarshal(bodyData, &body); err != nil {
		h.sendError(conn, env.Seq, errors.ErrInvalidParam)
		return
	}

	if body.ToUid == "" {
		h.sendError(conn, env.Seq, errors.ErrInvalidParam)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := h.seakingClient.SendFriendRequest(ctx, conn.UID(), body.ToUid, body.Message)
	if err != nil {
		log.Error().Err(err).Msg("failed to send friend request")
		h.sendError(conn, env.Seq, errors.ErrInternal)
		return
	}

	h.sendAck(conn, env.Seq, 0)
}

// handleGetFriendRequests 获取待处理好友请求
func (h *Handler) handleGetFriendRequests(conn *ws.Conn, env *protocol.Envelope) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	requests, err := h.seakingClient.GetPendingFriendRequests(ctx, conn.UID())
	if err != nil {
		log.Error().Err(err).Msg("failed to get friend requests")
		h.sendError(conn, env.Seq, errors.ErrInternal)
		return
	}

	h.sendResult(conn, env.Seq, map[string]interface{}{
		"requests": requests,
	})
}

// handleAcceptFriendRequest 接受好友请求
func (h *Handler) handleAcceptFriendRequest(conn *ws.Conn, env *protocol.Envelope) {
	var body struct {
		RequestId uint `json:"request_id"`
	}
	bodyData, _ := json.Marshal(env.Body)
	if err := json.Unmarshal(bodyData, &body); err != nil {
		h.sendError(conn, env.Seq, errors.ErrInvalidParam)
		return
	}

	if body.RequestId == 0 {
		h.sendError(conn, env.Seq, errors.ErrInvalidParam)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := h.seakingClient.AcceptFriendRequest(ctx, body.RequestId, conn.UID())
	if err != nil {
		log.Error().Err(err).Msg("failed to accept friend request")
		h.sendError(conn, env.Seq, errors.ErrInternal)
		return
	}

	h.sendAck(conn, env.Seq, 0)
}

// handleRejectFriendRequest 拒绝好友请求
func (h *Handler) handleRejectFriendRequest(conn *ws.Conn, env *protocol.Envelope) {
	var body struct {
		RequestId uint `json:"request_id"`
	}
	bodyData, _ := json.Marshal(env.Body)
	if err := json.Unmarshal(bodyData, &body); err != nil {
		h.sendError(conn, env.Seq, errors.ErrInvalidParam)
		return
	}

	if body.RequestId == 0 {
		h.sendError(conn, env.Seq, errors.ErrInvalidParam)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := h.seakingClient.RejectFriendRequest(ctx, body.RequestId, conn.UID())
	if err != nil {
		log.Error().Err(err).Msg("failed to reject friend request")
		h.sendError(conn, env.Seq, errors.ErrInternal)
		return
	}

	h.sendAck(conn, env.Seq, 0)
}

// handleDeleteFriend 删除好友
func (h *Handler) handleDeleteFriend(conn *ws.Conn, env *protocol.Envelope) {
	var body struct {
		FriendId string `json:"friend_id"`
	}
	bodyData, _ := json.Marshal(env.Body)
	if err := json.Unmarshal(bodyData, &body); err != nil {
		h.sendError(conn, env.Seq, errors.ErrInvalidParam)
		return
	}

	if body.FriendId == "" {
		h.sendError(conn, env.Seq, errors.ErrInvalidParam)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := h.seakingClient.DeleteFriend(ctx, conn.UID(), body.FriendId)
	if err != nil {
		log.Error().Err(err).Msg("failed to delete friend")
		h.sendError(conn, env.Seq, errors.ErrInternal)
		return
	}

	h.sendAck(conn, env.Seq, 0)
}

// ============== 会话相关处理 ==============

// handleGetConversations 获取会话列表
func (h *Handler) handleGetConversations(conn *ws.Conn, env *protocol.Envelope) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.seakingClient.GetUserConversations(ctx, conn.UID())
	if err != nil {
		log.Error().Err(err).Msg("failed to get conversations")
		h.sendError(conn, env.Seq, errors.ErrInternal)
		return
	}

	h.sendResult(conn, env.Seq, map[string]interface{}{
		"conversations": resp.Conversations,
	})
}

// handleCreateConversation 创建会话
func (h *Handler) handleCreateConversation(conn *ws.Conn, env *protocol.Envelope) {
	var body struct {
		Type      int      `json:"type"`       // 1=单聊, 2=群聊
		MemberIds []string `json:"member_ids"`
		Name      string   `json:"name,omitempty"`
	}
	bodyData, _ := json.Marshal(env.Body)
	if err := json.Unmarshal(bodyData, &body); err != nil {
		h.sendError(conn, env.Seq, errors.ErrInvalidParam)
		return
	}

	if body.Type != 1 && body.Type != 2 {
		h.sendError(conn, env.Seq, errors.ErrInvalidParam)
		return
	}
	if len(body.MemberIds) == 0 {
		h.sendError(conn, env.Seq, errors.ErrInvalidParam)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.seakingClient.CreateConversation(ctx, body.Type, conn.UID(), body.MemberIds, body.Name)
	if err != nil {
		log.Error().Err(err).Msg("failed to create conversation")
		h.sendError(conn, env.Seq, errors.ErrInternal)
		return
	}

	h.sendResult(conn, env.Seq, map[string]interface{}{
		"cid": resp.Cid,
	})
}

// handleGetConversationMembers 获取会话成员
func (h *Handler) handleGetConversationMembers(conn *ws.Conn, env *protocol.Envelope) {
	cid, ok := env.Body.(string)
	if !ok {
		h.sendError(conn, env.Seq, errors.ErrInvalidParam)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 检查权限
	accessResp, err := h.seakingClient.CheckAccess(ctx, conn.UID(), cid)
	if err != nil {
		log.Error().Err(err).Msg("failed to check access")
		h.sendError(conn, env.Seq, errors.ErrInternal)
		return
	}
	if !accessResp.HasAccess {
		h.sendError(conn, env.Seq, errors.ErrNotInConversation)
		return
	}

	resp, err := h.seakingClient.GetConversationMembers(ctx, cid)
	if err != nil {
		log.Error().Err(err).Msg("failed to get conversation members")
		h.sendError(conn, env.Seq, errors.ErrInternal)
		return
	}

	h.sendResult(conn, env.Seq, map[string]interface{}{
		"members": resp.Members,
	})
}

// ============== 群组相关处理 ==============

// handleGetGroups 获取群组列表
func (h *Handler) handleGetGroups(conn *ws.Conn, env *protocol.Envelope) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	groups, err := h.seakingClient.GetUserGroups(ctx, conn.UID())
	if err != nil {
		log.Error().Err(err).Msg("failed to get groups")
		h.sendError(conn, env.Seq, errors.ErrInternal)
		return
	}

	h.sendResult(conn, env.Seq, map[string]interface{}{
		"groups": groups,
	})
}

// handleCreateGroup 创建群组
func (h *Handler) handleCreateGroup(conn *ws.Conn, env *protocol.Envelope) {
	var body struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		MemberIds   []string `json:"member_ids"`
	}
	bodyData, _ := json.Marshal(env.Body)
	if err := json.Unmarshal(bodyData, &body); err != nil {
		h.sendError(conn, env.Seq, errors.ErrInvalidParam)
		return
	}

	if body.Name == "" {
		h.sendError(conn, env.Seq, errors.ErrInvalidParam)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	group, err := h.seakingClient.CreateGroup(ctx, conn.UID(), body.Name, body.Description, body.MemberIds)
	if err != nil {
		log.Error().Err(err).Msg("failed to create group")
		h.sendError(conn, env.Seq, errors.ErrInternal)
		return
	}

	h.sendResult(conn, env.Seq, map[string]interface{}{
		"group": group,
	})
}

// handleGetGroupInfo 获取群组信息
func (h *Handler) handleGetGroupInfo(conn *ws.Conn, env *protocol.Envelope) {
	groupId, ok := env.Body.(string)
	if !ok {
		h.sendError(conn, env.Seq, errors.ErrInvalidParam)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	group, err := h.seakingClient.GetGroupInfo(ctx, groupId)
	if err != nil {
		log.Error().Err(err).Msg("failed to get group info")
		h.sendError(conn, env.Seq, errors.ErrInternal)
		return
	}

	h.sendResult(conn, env.Seq, map[string]interface{}{
		"group": group,
	})
}

// handleGetGroupMembers 获取群组成员
func (h *Handler) handleGetGroupMembers(conn *ws.Conn, env *protocol.Envelope) {
	groupId, ok := env.Body.(string)
	if !ok {
		h.sendError(conn, env.Seq, errors.ErrInvalidParam)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	members, err := h.seakingClient.GetGroupMembers(ctx, groupId)
	if err != nil {
		log.Error().Err(err).Msg("failed to get group members")
		h.sendError(conn, env.Seq, errors.ErrInternal)
		return
	}

	h.sendResult(conn, env.Seq, map[string]interface{}{
		"members": members,
	})
}

// ============== 用户相关处理 ==============

// handleGetUserInfo 获取用户信息
func (h *Handler) handleGetUserInfo(conn *ws.Conn, env *protocol.Envelope) {
	uid, ok := env.Body.(string)
	if !ok {
		h.sendError(conn, env.Seq, errors.ErrInvalidParam)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user, err := h.seakingClient.GetUserInfo(ctx, uid)
	if err != nil {
		log.Error().Err(err).Msg("failed to get user info")
		h.sendError(conn, env.Seq, errors.ErrInternal)
		return
	}

	h.sendResult(conn, env.Seq, map[string]interface{}{
		"user": user,
	})
}
