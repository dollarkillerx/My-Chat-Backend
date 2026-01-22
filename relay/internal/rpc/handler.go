package rpc

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/my-chat/common/pkg/protocol"
	"github.com/my-chat/relay/internal/conf"
	"github.com/my-chat/relay/internal/service/event"
)

// Handler RPC处理器
type Handler struct {
	eventService *event.Service
	config       conf.RelayConfiguration
	methods      map[string]MethodHandler
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
func NewHandler(eventService *event.Service, config conf.RelayConfiguration) *Handler {
	h := &Handler{
		eventService: eventService,
		config:       config,
		methods:      make(map[string]MethodHandler),
	}
	h.registerMethods()
	return h
}

// registerMethods 注册所有方法
func (h *Handler) registerMethods() {
	h.methods["relay.storeEvent"] = h.storeEvent
	h.methods["relay.getEvent"] = h.getEvent
	h.methods["relay.queryEvents"] = h.queryEvents
	h.methods["relay.syncEvents"] = h.syncEvents
	h.methods["relay.updateReadReceipt"] = h.updateReadReceipt
	h.methods["relay.validateRevoke"] = h.validateRevoke
	h.methods["relay.validateEdit"] = h.validateEdit
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

// storeEvent 存储事件
func (h *Handler) storeEvent(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Event *protocol.Event `json:"event"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	stored, err := h.eventService.StoreEvent(ctx, req.Event)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"mid":       stored.Mid,
		"timestamp": stored.Timestamp,
	}, nil
}

// getEvent 获取事件
func (h *Handler) getEvent(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Mid int64 `json:"mid"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	e, err := h.eventService.GetEvent(ctx, req.Mid)
	if err != nil {
		return nil, err
	}

	return e, nil
}

// queryEvents 查询事件
func (h *Handler) queryEvents(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req event.QueryRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	events, err := h.eventService.QueryEvents(ctx, &req)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"events": events,
	}, nil
}

// syncEvents 同步最新事件
func (h *Handler) syncEvents(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Cid   string `json:"cid"`
		Limit int    `json:"limit"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	if req.Limit <= 0 {
		req.Limit = 50
	}

	events, err := h.eventService.QueryEventsDesc(ctx, req.Cid, req.Limit)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"events": events,
	}, nil
}

// updateReadReceipt 更新已读回执
func (h *Handler) updateReadReceipt(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Cid         string `json:"cid"`
		Uid         string `json:"uid"`
		LastReadMid int64  `json:"last_read_mid"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	err := h.eventService.UpdateReadReceipt(ctx, req.Cid, req.Uid, req.LastReadMid)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"success": true,
	}, nil
}

// validateRevoke 验证撤销权限
func (h *Handler) validateRevoke(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Cid       string `json:"cid"`
		Uid       string `json:"uid"`
		TargetMid int64  `json:"target_mid"`
		IsAdmin   bool   `json:"is_admin"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	// 获取目标消息
	targetEvent, err := h.eventService.GetEvent(ctx, req.TargetMid)
	if err != nil {
		return map[string]interface{}{
			"valid":  false,
			"reason": "message not found",
		}, nil
	}

	// 检查是否属于同一会话
	if targetEvent.Cid != req.Cid {
		return map[string]interface{}{
			"valid":  false,
			"reason": "message not in this conversation",
		}, nil
	}

	// 检查是否是自己发送的消息，或者是管理员
	if targetEvent.Sender != req.Uid && !req.IsAdmin {
		return map[string]interface{}{
			"valid":  false,
			"reason": "no permission to revoke",
		}, nil
	}

	// 检查时间窗口（2分钟内可撤销）
	revokeWindow := int64(2 * 60) // 2分钟
	if time.Now().Unix()-targetEvent.Timestamp > revokeWindow && !req.IsAdmin {
		return map[string]interface{}{
			"valid":  false,
			"reason": "revoke time exceeded",
		}, nil
	}

	return map[string]interface{}{
		"valid": true,
	}, nil
}

// validateEdit 验证编辑权限
func (h *Handler) validateEdit(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Cid       string `json:"cid"`
		Uid       string `json:"uid"`
		TargetMid int64  `json:"target_mid"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	// 获取目标消息
	targetEvent, err := h.eventService.GetEvent(ctx, req.TargetMid)
	if err != nil {
		return map[string]interface{}{
			"valid":  false,
			"reason": "message not found",
		}, nil
	}

	// 检查是否属于同一会话
	if targetEvent.Cid != req.Cid {
		return map[string]interface{}{
			"valid":  false,
			"reason": "message not in this conversation",
		}, nil
	}

	// 只能编辑自己发送的消息
	if targetEvent.Sender != req.Uid {
		return map[string]interface{}{
			"valid":  false,
			"reason": "can only edit your own message",
		}, nil
	}

	// 只能编辑文本消息
	if targetEvent.Kind != protocol.KindText {
		return map[string]interface{}{
			"valid":  false,
			"reason": "can only edit text message",
		}, nil
	}

	// 检查时间窗口（24小时内可编辑）
	editWindow := int64(24 * 60 * 60) // 24小时
	if time.Now().Unix()-targetEvent.Timestamp > editWindow {
		return map[string]interface{}{
			"valid":  false,
			"reason": "edit time exceeded",
		}, nil
	}

	return map[string]interface{}{
		"valid": true,
	}, nil
}
