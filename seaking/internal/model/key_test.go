package model

import (
	"testing"
	"time"
)

func TestUserKey_TableName(t *testing.T) {
	uk := UserKey{}
	if uk.TableName() != "user_keys" {
		t.Errorf("TableName() = %v, want %v", uk.TableName(), "user_keys")
	}
}

func TestChatKey_TableName(t *testing.T) {
	ck := ChatKey{}
	if ck.TableName() != "chat_keys" {
		t.Errorf("TableName() = %v, want %v", ck.TableName(), "chat_keys")
	}
}

func TestGroupKey_TableName(t *testing.T) {
	gk := GroupKey{}
	if gk.TableName() != "group_keys" {
		t.Errorf("TableName() = %v, want %v", gk.TableName(), "group_keys")
	}
}

func TestUserKey_Fields(t *testing.T) {
	now := time.Now()
	uk := UserKey{
		ID:                  1,
		UserID:              "user123",
		PublicKey:           "public-key-base64",
		EncryptedPrivateKey: "encrypted-private-key",
		KeySalt:             "salt-base64",
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	if uk.ID != 1 {
		t.Errorf("ID = %v, want %v", uk.ID, 1)
	}
	if uk.UserID != "user123" {
		t.Errorf("UserID = %v, want %v", uk.UserID, "user123")
	}
	if uk.PublicKey != "public-key-base64" {
		t.Errorf("PublicKey mismatch")
	}
	if uk.EncryptedPrivateKey != "encrypted-private-key" {
		t.Errorf("EncryptedPrivateKey mismatch")
	}
	if uk.KeySalt != "salt-base64" {
		t.Errorf("KeySalt mismatch")
	}
}

func TestChatKey_Fields(t *testing.T) {
	now := time.Now()
	ck := ChatKey{
		ID:             1,
		ConversationID: "d:user1:user2",
		UserID:         "user1",
		EncryptedKey:   "encrypted-key-base64",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if ck.ID != 1 {
		t.Errorf("ID = %v, want %v", ck.ID, 1)
	}
	if ck.ConversationID != "d:user1:user2" {
		t.Errorf("ConversationID = %v, want %v", ck.ConversationID, "d:user1:user2")
	}
	if ck.UserID != "user1" {
		t.Errorf("UserID = %v, want %v", ck.UserID, "user1")
	}
	if ck.EncryptedKey != "encrypted-key-base64" {
		t.Errorf("EncryptedKey mismatch")
	}
}

func TestGroupKey_Fields(t *testing.T) {
	now := time.Now()
	gk := GroupKey{
		ID:           1,
		GroupID:      "group123",
		UserID:       "user1",
		EncryptedKey: "encrypted-key-base64",
		Version:      2,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if gk.ID != 1 {
		t.Errorf("ID = %v, want %v", gk.ID, 1)
	}
	if gk.GroupID != "group123" {
		t.Errorf("GroupID = %v, want %v", gk.GroupID, "group123")
	}
	if gk.UserID != "user1" {
		t.Errorf("UserID = %v, want %v", gk.UserID, "user1")
	}
	if gk.Version != 2 {
		t.Errorf("Version = %v, want %v", gk.Version, 2)
	}
}

func TestUserKeyInfo_Fields(t *testing.T) {
	info := UserKeyInfo{
		UserID:    "user123",
		PublicKey: "public-key",
	}

	if info.UserID != "user123" {
		t.Errorf("UserID = %v, want %v", info.UserID, "user123")
	}
	if info.PublicKey != "public-key" {
		t.Errorf("PublicKey = %v, want %v", info.PublicKey, "public-key")
	}
}

func TestChatKeyInfo_Fields(t *testing.T) {
	info := ChatKeyInfo{
		ConversationID: "d:user1:user2",
		EncryptedKey:   "encrypted-key",
	}

	if info.ConversationID != "d:user1:user2" {
		t.Errorf("ConversationID = %v, want %v", info.ConversationID, "d:user1:user2")
	}
	if info.EncryptedKey != "encrypted-key" {
		t.Errorf("EncryptedKey = %v, want %v", info.EncryptedKey, "encrypted-key")
	}
}

func TestGroupKeyInfo_Fields(t *testing.T) {
	info := GroupKeyInfo{
		GroupID:      "group123",
		EncryptedKey: "encrypted-key",
		Version:      3,
	}

	if info.GroupID != "group123" {
		t.Errorf("GroupID = %v, want %v", info.GroupID, "group123")
	}
	if info.EncryptedKey != "encrypted-key" {
		t.Errorf("EncryptedKey = %v, want %v", info.EncryptedKey, "encrypted-key")
	}
	if info.Version != 3 {
		t.Errorf("Version = %v, want %v", info.Version, 3)
	}
}
