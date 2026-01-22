package ws

import (
	"testing"

	"github.com/my-chat/gateway/internal/conf"
)

func TestNewHub(t *testing.T) {
	config := conf.GatewayConfiguration{
		MaxConnPerUser:   5,
		HeartbeatTimeout: 30,
	}

	hub := NewHub(config)

	if hub == nil {
		t.Fatal("NewHub returned nil")
	}

	if hub.GetOnlineUsers() != 0 {
		t.Errorf("initial online users = %d, want 0", hub.GetOnlineUsers())
	}

	if hub.GetTotalConns() != 0 {
		t.Errorf("initial total conns = %d, want 0", hub.GetTotalConns())
	}
}

func TestHubConfig(t *testing.T) {
	config := conf.GatewayConfiguration{
		MaxConnPerUser:   10,
		HeartbeatTimeout: 60,
		WriteTimeout:     10,
		ReadTimeout:      10,
	}

	hub := NewHub(config)

	if hub == nil {
		t.Fatal("NewHub returned nil")
	}

	// Hub should be properly configured
	// We can't directly access config, but hub should not be nil
}
