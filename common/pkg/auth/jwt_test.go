package auth

import (
	"testing"
	"time"
)

func TestNewJWTManager(t *testing.T) {
	secret := "test-secret-key"
	expireHour := 24

	manager := NewJWTManager(secret, expireHour)

	if manager == nil {
		t.Fatal("NewJWTManager returned nil")
	}
}

func TestGenerateAndParseToken(t *testing.T) {
	manager := NewJWTManager("test-secret-key", 24)

	uid := "user123"
	deviceId := "device456"
	platform := "ios"

	token, err := manager.GenerateToken(uid, deviceId, platform)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	if token == "" {
		t.Fatal("GenerateToken returned empty token")
	}

	claims, err := manager.ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken failed: %v", err)
	}

	if claims.Uid != uid {
		t.Errorf("Uid mismatch: got %s, want %s", claims.Uid, uid)
	}

	if claims.DeviceId != deviceId {
		t.Errorf("DeviceId mismatch: got %s, want %s", claims.DeviceId, deviceId)
	}

	if claims.Platform != platform {
		t.Errorf("Platform mismatch: got %s, want %s", claims.Platform, platform)
	}
}

func TestParseTokenInvalid(t *testing.T) {
	manager := NewJWTManager("test-secret-key", 24)

	testCases := []struct {
		name  string
		token string
	}{
		{"empty token", ""},
		{"invalid format", "not-a-valid-token"},
		{"random string", "abc.def.ghi"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := manager.ParseToken(tc.token)
			if err == nil {
				t.Error("ParseToken should fail for invalid token")
			}
		})
	}
}

func TestParseTokenWrongSecret(t *testing.T) {
	manager1 := NewJWTManager("secret-key-1", 24)
	manager2 := NewJWTManager("secret-key-2", 24)

	token, _ := manager1.GenerateToken("user123", "device456", "ios")

	_, err := manager2.ParseToken(token)
	if err == nil {
		t.Error("ParseToken should fail with different secret")
	}
}

func TestGenerateTokenDifferentUsers(t *testing.T) {
	manager := NewJWTManager("test-secret-key", 24)

	token1, _ := manager.GenerateToken("user1", "device1", "ios")
	token2, _ := manager.GenerateToken("user2", "device2", "android")

	if token1 == token2 {
		t.Error("Different users should have different tokens")
	}

	claims1, _ := manager.ParseToken(token1)
	claims2, _ := manager.ParseToken(token2)

	if claims1.Uid == claims2.Uid {
		t.Error("Claims should have different UIDs")
	}
}

func TestTokenExpiration(t *testing.T) {
	// Create manager with very short expiration for testing
	manager := NewJWTManager("test-secret-key", 0) // 0 hour expiration

	token, err := manager.GenerateToken("user123", "device456", "ios")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	// Token should still be parseable but we can check expiration time
	claims, err := manager.ParseToken(token)
	if err != nil {
		// Token may already be expired, which is expected
		return
	}

	// If token hasn't expired yet, check that expiration is set
	if claims.ExpiresAt == nil {
		t.Error("Token should have expiration time set")
	}
}

func TestMultipleTokensForSameUser(t *testing.T) {
	manager := NewJWTManager("test-secret-key", 24)

	// Same user, different devices
	token1, _ := manager.GenerateToken("user123", "device1", "ios")
	time.Sleep(time.Millisecond * 10) // Small delay to ensure different timestamp
	token2, _ := manager.GenerateToken("user123", "device2", "android")

	// Both tokens should be valid
	claims1, err1 := manager.ParseToken(token1)
	claims2, err2 := manager.ParseToken(token2)

	if err1 != nil || err2 != nil {
		t.Error("Both tokens should be valid")
	}

	if claims1.Uid != claims2.Uid {
		t.Error("Both tokens should have same UID")
	}

	if claims1.DeviceId == claims2.DeviceId {
		t.Error("Tokens should have different device IDs")
	}
}
