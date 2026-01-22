package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
	"time"
)

// RPCClient JSON-RPC 2.0 客户端
type RPCClient struct {
	baseURL    string
	httpClient *http.Client
	idCounter  uint64
}

// RPCRequest JSON-RPC 请求
type RPCRequest struct {
	JsonRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
	ID      string      `json:"id"`
}

// RPCResponse JSON-RPC 响应
type RPCResponse struct {
	JsonRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
	ID      string          `json:"id"`
}

// RPCError JSON-RPC 错误
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (e *RPCError) Error() string {
	return fmt.Sprintf("rpc error [%d]: %s", e.Code, e.Message)
}

// RPCClientOption 客户端选项
type RPCClientOption func(*RPCClient)

// WithRPCTimeout 设置超时时间
func WithRPCTimeout(timeout time.Duration) RPCClientOption {
	return func(c *RPCClient) {
		c.httpClient.Timeout = timeout
	}
}

// NewRPCClient 创建JSON-RPC客户端
func NewRPCClient(baseURL string, opts ...RPCClientOption) *RPCClient {
	c := &RPCClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Call 调用远程方法
func (c *RPCClient) Call(ctx context.Context, method string, params interface{}, result interface{}) error {
	// 生成请求ID
	id := atomic.AddUint64(&c.idCounter, 1)

	req := &RPCRequest{
		JsonRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      fmt.Sprintf("%d", id),
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	var rpcResp RPCResponse
	if err := json.Unmarshal(respBody, &rpcResp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if rpcResp.Error != nil {
		return rpcResp.Error
	}

	if result != nil && len(rpcResp.Result) > 0 {
		if err := json.Unmarshal(rpcResp.Result, result); err != nil {
			return fmt.Errorf("unmarshal result: %w", err)
		}
	}

	return nil
}

// Notify 发送通知（不需要响应）
func (c *RPCClient) Notify(ctx context.Context, method string, params interface{}) error {
	req := &RPCRequest{
		JsonRPC: "2.0",
		Method:  method,
		Params:  params,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	return nil
}
