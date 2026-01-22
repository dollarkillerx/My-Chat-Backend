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
