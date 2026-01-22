package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	commonerrors "github.com/my-chat/common/pkg/errors"
	"github.com/my-chat/common/pkg/protocol"
	"github.com/my-chat/relay/internal/service/event"
)

// API 接口层
type API struct {
	eventService *event.Service
}

// NewAPI 创建API
func NewAPI(eventService *event.Service) *API {
	return &API{
		eventService: eventService,
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

// StoreEvent 存储事件
func (a *API) StoreEvent(c *gin.Context) {
	var event protocol.Event
	if err := c.ShouldBindJSON(&event); err != nil {
		Error(c, commonerrors.ErrInvalidParam)
		return
	}

	stored, err := a.eventService.StoreEvent(c.Request.Context(), &event)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{
		"mid":       stored.Mid,
		"timestamp": stored.Timestamp,
	})
}

// GetEvent 获取事件
func (a *API) GetEvent(c *gin.Context) {
	midStr := c.Param("mid")
	mid, err := strconv.ParseInt(midStr, 10, 64)
	if err != nil {
		Error(c, commonerrors.ErrInvalidParam)
		return
	}

	e, err := a.eventService.GetEvent(c.Request.Context(), mid)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, e)
}

// QueryEvents 查询事件
func (a *API) QueryEvents(c *gin.Context) {
	var req event.QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, commonerrors.ErrInvalidParam)
		return
	}

	events, err := a.eventService.QueryEvents(c.Request.Context(), &req)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, events)
}

// SyncEvents 同步事件（获取最新消息）
func (a *API) SyncEvents(c *gin.Context) {
	cid := c.Query("cid")
	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)

	if cid == "" {
		Error(c, commonerrors.ErrInvalidParam)
		return
	}

	events, err := a.eventService.QueryEventsDesc(c.Request.Context(), cid, limit)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, events)
}

// UpdateReadReceipt 更新已读回执
func (a *API) UpdateReadReceipt(c *gin.Context) {
	var req struct {
		Cid         string `json:"cid" binding:"required"`
		Uid         string `json:"uid" binding:"required"`
		LastReadMid int64  `json:"last_read_mid" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, commonerrors.ErrInvalidParam)
		return
	}

	if err := a.eventService.UpdateReadReceipt(c.Request.Context(), req.Cid, req.Uid, req.LastReadMid); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// GetReadReceipts 获取会话已读回执
func (a *API) GetReadReceipts(c *gin.Context) {
	cid := c.Query("cid")
	if cid == "" {
		Error(c, commonerrors.ErrInvalidParam)
		return
	}

	receipts, err := a.eventService.GetConversationReadReceipts(c.Request.Context(), cid)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, receipts)
}

// AddReaction 添加反应
func (a *API) AddReaction(c *gin.Context) {
	var req struct {
		Mid   int64  `json:"mid" binding:"required"`
		Cid   string `json:"cid" binding:"required"`
		Uid   string `json:"uid" binding:"required"`
		Emoji string `json:"emoji" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, commonerrors.ErrInvalidParam)
		return
	}

	if err := a.eventService.AddReaction(c.Request.Context(), req.Mid, req.Cid, req.Uid, req.Emoji); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// RemoveReaction 移除反应
func (a *API) RemoveReaction(c *gin.Context) {
	var req struct {
		Mid   int64  `json:"mid" binding:"required"`
		Uid   string `json:"uid" binding:"required"`
		Emoji string `json:"emoji" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, commonerrors.ErrInvalidParam)
		return
	}

	if err := a.eventService.RemoveReaction(c.Request.Context(), req.Mid, req.Uid, req.Emoji); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// GetReactions 获取消息反应
func (a *API) GetReactions(c *gin.Context) {
	midStr := c.Param("mid")
	mid, err := strconv.ParseInt(midStr, 10, 64)
	if err != nil {
		Error(c, commonerrors.ErrInvalidParam)
		return
	}

	summary, err := a.eventService.GetReactionSummary(c.Request.Context(), mid)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, summary)
}
