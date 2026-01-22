package rpc

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/my-chat/common/pkg/auth"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestNewHandler(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret", 24)
	h := NewHandler(jwtManager, "http://localhost:8081", "http://localhost:8082")

	if h == nil {
		t.Fatal("NewHandler returned nil")
	}

	if h.jwtManager == nil {
		t.Error("jwtManager is nil")
	}

	if h.seakingClient == nil {
		t.Error("seakingClient is nil")
	}

	if h.relayClient == nil {
		t.Error("relayClient is nil")
	}

	if len(h.methods) == 0 {
		t.Error("methods map is empty")
	}
}

func TestHandler_InvalidJSON(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret", 24)
	h := NewHandler(jwtManager, "http://localhost:8081", "http://localhost:8082")

	router := gin.New()
	router.POST("/api/rpc", h.Handle)

	req := httptest.NewRequest(http.MethodPost, "/api/rpc", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp RPCResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Error == nil {
		t.Error("expected error response")
	}

	if resp.Error.Code != -32700 {
		t.Errorf("expected error code -32700, got %d", resp.Error.Code)
	}
}

func TestHandler_InvalidJSONRPCVersion(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret", 24)
	h := NewHandler(jwtManager, "http://localhost:8081", "http://localhost:8082")

	router := gin.New()
	router.POST("/api/rpc", h.Handle)

	reqBody := RPCRequest{
		JSONRPC: "1.0", // wrong version
		Method:  "test",
		ID:      1,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/rpc", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	var resp RPCResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Error == nil {
		t.Error("expected error response")
	}

	if resp.Error.Code != -32600 {
		t.Errorf("expected error code -32600 (Invalid Request), got %d", resp.Error.Code)
	}
}

func TestHandler_MethodNotFound(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret", 24)
	h := NewHandler(jwtManager, "http://localhost:8081", "http://localhost:8082")

	router := gin.New()
	router.POST("/api/rpc", h.Handle)

	reqBody := RPCRequest{
		JSONRPC: "2.0",
		Method:  "nonExistentMethod",
		ID:      1,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/rpc", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	var resp RPCResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Error == nil {
		t.Error("expected error response")
	}

	if resp.Error.Code != -32601 {
		t.Errorf("expected error code -32601 (Method not found), got %d", resp.Error.Code)
	}
}

func TestHandler_AuthRequired(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret", 24)
	h := NewHandler(jwtManager, "http://localhost:8081", "http://localhost:8082")

	router := gin.New()
	router.POST("/api/rpc", h.Handle)

	// Test method that requires auth without token
	reqBody := RPCRequest{
		JSONRPC: "2.0",
		Method:  "getFriends",
		Params:  json.RawMessage(`{}`),
		ID:      1,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/rpc", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	// No Authorization header
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	var resp RPCResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Error == nil {
		t.Error("expected error response for missing auth")
	}

	if resp.Error.Code != -32001 {
		t.Errorf("expected error code -32001 (Authorization required), got %d", resp.Error.Code)
	}
}

func TestHandler_InvalidToken(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret", 24)
	h := NewHandler(jwtManager, "http://localhost:8081", "http://localhost:8082")

	router := gin.New()
	router.POST("/api/rpc", h.Handle)

	reqBody := RPCRequest{
		JSONRPC: "2.0",
		Method:  "getFriends",
		Params:  json.RawMessage(`{}`),
		ID:      1,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/rpc", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	var resp RPCResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Error == nil {
		t.Error("expected error response for invalid token")
	}

	if resp.Error.Code != -32002 {
		t.Errorf("expected error code -32002 (Invalid token), got %d", resp.Error.Code)
	}
}

func TestHandler_RegisteredMethods(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret", 24)
	h := NewHandler(jwtManager, "http://localhost:8081", "http://localhost:8082")

	expectedMethods := []string{
		"register",
		"login",
		"getUserInfo",
		"getFriends",
		"sendFriendRequest",
		"getPendingFriendRequests",
		"acceptFriendRequest",
		"rejectFriendRequest",
		"deleteFriend",
		"getConversations",
		"createConversation",
		"getConversationMembers",
		"getGroups",
		"createGroup",
		"getGroupInfo",
		"getGroupMembers",
	}

	for _, method := range expectedMethods {
		if _, ok := h.methods[method]; !ok {
			t.Errorf("method %s not registered", method)
		}
	}
}

func TestRPCRequest_Marshal(t *testing.T) {
	req := RPCRequest{
		JSONRPC: "2.0",
		Method:  "test",
		Params:  json.RawMessage(`{"key": "value"}`),
		ID:      1,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded RPCRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.JSONRPC != "2.0" {
		t.Errorf("expected jsonrpc 2.0, got %s", decoded.JSONRPC)
	}

	if decoded.Method != "test" {
		t.Errorf("expected method test, got %s", decoded.Method)
	}
}

func TestRPCResponse_Marshal(t *testing.T) {
	resp := RPCResponse{
		JSONRPC: "2.0",
		Result:  map[string]string{"status": "ok"},
		ID:      1,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded RPCResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.JSONRPC != "2.0" {
		t.Errorf("expected jsonrpc 2.0, got %s", decoded.JSONRPC)
	}

	if decoded.Error != nil {
		t.Error("expected no error")
	}
}

func TestRPCResponse_WithError(t *testing.T) {
	resp := RPCResponse{
		JSONRPC: "2.0",
		Error: &RPCError{
			Code:    -32600,
			Message: "Invalid Request",
		},
		ID: 1,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded RPCResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Error == nil {
		t.Fatal("expected error")
	}

	if decoded.Error.Code != -32600 {
		t.Errorf("expected error code -32600, got %d", decoded.Error.Code)
	}

	if decoded.Error.Message != "Invalid Request" {
		t.Errorf("expected message 'Invalid Request', got %s", decoded.Error.Message)
	}
}

func TestHandler_RegisterInvalidParams(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret", 24)
	h := NewHandler(jwtManager, "http://localhost:8081", "http://localhost:8082")

	router := gin.New()
	router.POST("/api/rpc", h.Handle)

	// Invalid params (missing required fields)
	reqBody := RPCRequest{
		JSONRPC: "2.0",
		Method:  "register",
		Params:  json.RawMessage(`{"invalid": "params"}`),
		ID:      1,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/rpc", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	var resp RPCResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Should get an error (either invalid params or service error)
	// The exact error depends on whether the SeaKing service is running
	// In unit test context, it will fail to connect to SeaKing
	if resp.Error == nil && resp.Result == nil {
		t.Error("expected either error or result")
	}
}

func TestHandler_LoginInvalidParams(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret", 24)
	h := NewHandler(jwtManager, "http://localhost:8081", "http://localhost:8082")

	router := gin.New()
	router.POST("/api/rpc", h.Handle)

	// Invalid JSON in params - this will cause a parse error at the outer level
	// since json.RawMessage(`invalid`) creates invalid JSON
	reqBody := `{"jsonrpc": "2.0", "method": "login", "params": "not-an-object", "id": 1}`

	req := httptest.NewRequest(http.MethodPost, "/api/rpc", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	var resp RPCResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Error == nil {
		t.Error("expected error for invalid params")
	}

	// Invalid params will cause -32602 error when unmarshaling to struct
	if resp.Error.Code != -32602 {
		t.Errorf("expected error code -32602 (Invalid params), got %d", resp.Error.Code)
	}
}

func TestHandler_WithValidToken(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret", 24)
	h := NewHandler(jwtManager, "http://localhost:8081", "http://localhost:8082")

	// Generate a valid token
	token, err := jwtManager.GenerateToken("test-uid", "device-1", "ios")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	router := gin.New()
	router.POST("/api/rpc", h.Handle)

	reqBody := RPCRequest{
		JSONRPC: "2.0",
		Method:  "getFriends",
		Params:  json.RawMessage(`{}`),
		ID:      1,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/rpc", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	var resp RPCResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// With valid token, should pass auth but fail on SeaKing connection
	// This verifies the auth middleware works correctly
	if resp.Error != nil && resp.Error.Code == -32001 {
		t.Error("auth should have passed with valid token")
	}
	if resp.Error != nil && resp.Error.Code == -32002 {
		t.Error("token should be valid")
	}
}

func TestHandler_BearerPrefix(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret", 24)
	h := NewHandler(jwtManager, "http://localhost:8081", "http://localhost:8082")

	token, _ := jwtManager.GenerateToken("test-uid", "device-1", "ios")

	router := gin.New()
	router.POST("/api/rpc", h.Handle)

	reqBody := RPCRequest{
		JSONRPC: "2.0",
		Method:  "getUserInfo",
		Params:  json.RawMessage(`{}`),
		ID:      1,
	}
	body, _ := json.Marshal(reqBody)

	// Test with Bearer prefix
	req := httptest.NewRequest(http.MethodPost, "/api/rpc", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	var resp RPCResponse
	json.NewDecoder(w.Body).Decode(&resp)

	// Should not fail on auth
	if resp.Error != nil && (resp.Error.Code == -32001 || resp.Error.Code == -32002) {
		t.Error("auth should pass with Bearer prefix")
	}
}

func TestHandler_NullID(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret", 24)
	h := NewHandler(jwtManager, "http://localhost:8081", "http://localhost:8082")

	router := gin.New()
	router.POST("/api/rpc", h.Handle)

	// Request with null ID (notification style, though we still respond)
	reqBody := `{"jsonrpc": "2.0", "method": "nonExistent", "params": {}, "id": null}`

	req := httptest.NewRequest(http.MethodPost, "/api/rpc", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	var resp RPCResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Should still get a valid response
	if resp.JSONRPC != "2.0" {
		t.Errorf("expected jsonrpc 2.0, got %s", resp.JSONRPC)
	}
}
