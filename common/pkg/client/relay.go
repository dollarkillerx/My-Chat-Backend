package client

import (
	"context"
	"time"

	"github.com/my-chat/common/pkg/protocol"
)

// RelayClient Relay服务RPC客户端
type RelayClient struct {
	rpc *RPCClient
}

// NewRelayClient 创建Relay客户端
func NewRelayClient(addr string) *RelayClient {
	return &RelayClient{
		rpc: NewRPCClient(addr+"/api/rpc", WithRPCTimeout(5*time.Second)),
	}
}

// StoreEventRequest 存储事件请求
type StoreEventRequest struct {
	Event *protocol.Event `json:"event"`
}

// StoreEventResponse 存储事件响应
type StoreEventResponse struct {
	Mid       int64 `json:"mid"`
	Timestamp int64 `json:"timestamp"`
}

// StoreEvent 存储事件
func (c *RelayClient) StoreEvent(ctx context.Context, event *protocol.Event) (*StoreEventResponse, error) {
	var resp StoreEventResponse
	err := c.rpc.Call(ctx, "relay.storeEvent", &StoreEventRequest{Event: event}, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// QueryEventsRequest 查询事件请求
type QueryEventsRequest struct {
	Cid     string `json:"cid"`
	LastMid int64  `json:"last_mid,omitempty"`
	Before  int64  `json:"before,omitempty"`
	After   int64  `json:"after,omitempty"`
	Kinds   []int  `json:"kinds,omitempty"`
	Limit   int    `json:"limit,omitempty"`
}

// EventData 事件数据
type EventData struct {
	Mid       int64  `json:"mid"`
	Cid       string `json:"cid"`
	Kind      int    `json:"kind"`
	Sender    string `json:"sender"`
	Tags      string `json:"tags"`
	Data      string `json:"data"`
	Timestamp int64  `json:"timestamp"`
}

// QueryEventsResponse 查询事件响应
type QueryEventsResponse struct {
	Events []EventData `json:"events"`
}

// QueryEvents 查询事件
func (c *RelayClient) QueryEvents(ctx context.Context, req *QueryEventsRequest) (*QueryEventsResponse, error) {
	var resp QueryEventsResponse
	err := c.rpc.Call(ctx, "relay.queryEvents", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// SyncEventsRequest 同步事件请求
type SyncEventsRequest struct {
	Cid   string `json:"cid"`
	Limit int    `json:"limit,omitempty"`
}

// SyncEvents 同步最新事件
func (c *RelayClient) SyncEvents(ctx context.Context, cid string, limit int) (*QueryEventsResponse, error) {
	var resp QueryEventsResponse
	err := c.rpc.Call(ctx, "relay.syncEvents", &SyncEventsRequest{Cid: cid, Limit: limit}, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetEventRequest 获取单条事件请求
type GetEventRequest struct {
	Mid int64 `json:"mid"`
}

// GetEvent 获取单条事件
func (c *RelayClient) GetEvent(ctx context.Context, mid int64) (*EventData, error) {
	var resp EventData
	err := c.rpc.Call(ctx, "relay.getEvent", &GetEventRequest{Mid: mid}, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateReadReceiptRequest 更新已读回执请求
type UpdateReadReceiptRequest struct {
	Cid         string `json:"cid"`
	Uid         string `json:"uid"`
	LastReadMid int64  `json:"last_read_mid"`
}

// UpdateReadReceipt 更新已读回执
func (c *RelayClient) UpdateReadReceipt(ctx context.Context, cid, uid string, lastReadMid int64) error {
	return c.rpc.Call(ctx, "relay.updateReadReceipt", &UpdateReadReceiptRequest{
		Cid:         cid,
		Uid:         uid,
		LastReadMid: lastReadMid,
	}, nil)
}

// ValidateRevokeRequest 验证撤销请求
type ValidateRevokeRequest struct {
	Cid       string `json:"cid"`
	Uid       string `json:"uid"`
	TargetMid int64  `json:"target_mid"`
	IsAdmin   bool   `json:"is_admin"`
}

// ValidateRevokeResponse 验证撤销响应
type ValidateRevokeResponse struct {
	Valid  bool   `json:"valid"`
	Reason string `json:"reason,omitempty"`
}

// ValidateRevoke 验证撤销权限
func (c *RelayClient) ValidateRevoke(ctx context.Context, cid, uid string, targetMid int64, isAdmin bool) (*ValidateRevokeResponse, error) {
	var resp ValidateRevokeResponse
	err := c.rpc.Call(ctx, "relay.validateRevoke", &ValidateRevokeRequest{
		Cid:       cid,
		Uid:       uid,
		TargetMid: targetMid,
		IsAdmin:   isAdmin,
	}, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ValidateEditRequest 验证编辑请求
type ValidateEditRequest struct {
	Cid       string `json:"cid"`
	Uid       string `json:"uid"`
	TargetMid int64  `json:"target_mid"`
}

// ValidateEdit 验证编辑权限
func (c *RelayClient) ValidateEdit(ctx context.Context, cid, uid string, targetMid int64) (*ValidateRevokeResponse, error) {
	var resp ValidateRevokeResponse
	err := c.rpc.Call(ctx, "relay.validateEdit", &ValidateEditRequest{
		Cid:       cid,
		Uid:       uid,
		TargetMid: targetMid,
	}, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
