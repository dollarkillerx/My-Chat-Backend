package model

import (
	"testing"
	"time"
)

func TestConversation_TableName(t *testing.T) {
	c := Conversation{}
	if c.TableName() != "conversations" {
		t.Errorf("TableName() = %v, want %v", c.TableName(), "conversations")
	}
}

func TestConversationMember_TableName(t *testing.T) {
	cm := ConversationMember{}
	if cm.TableName() != "conversation_members" {
		t.Errorf("TableName() = %v, want %v", cm.TableName(), "conversation_members")
	}
}

func TestConversation_Fields(t *testing.T) {
	now := time.Now()
	c := Conversation{
		ID:        "d:user1:user2",
		Type:      ConversationTypeDirect,
		Name:      "Test Conversation",
		Avatar:    "https://example.com/avatar.jpg",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if c.ID != "d:user1:user2" {
		t.Errorf("ID = %v, want %v", c.ID, "d:user1:user2")
	}
	if c.Type != ConversationTypeDirect {
		t.Errorf("Type = %v, want %v", c.Type, ConversationTypeDirect)
	}
	if c.Name != "Test Conversation" {
		t.Errorf("Name = %v, want %v", c.Name, "Test Conversation")
	}
}

func TestConversationMember_Fields(t *testing.T) {
	now := time.Now()
	cm := ConversationMember{
		ID:             1,
		ConversationID: "d:user1:user2",
		UserID:         "user1",
		LastReadMid:    100,
		Muted:          true,
		Pinned:         false,
		JoinedAt:       now,
	}

	if cm.ID != 1 {
		t.Errorf("ID = %v, want %v", cm.ID, 1)
	}
	if cm.ConversationID != "d:user1:user2" {
		t.Errorf("ConversationID = %v, want %v", cm.ConversationID, "d:user1:user2")
	}
	if cm.UserID != "user1" {
		t.Errorf("UserID = %v, want %v", cm.UserID, "user1")
	}
	if cm.LastReadMid != 100 {
		t.Errorf("LastReadMid = %v, want %v", cm.LastReadMid, 100)
	}
	if !cm.Muted {
		t.Error("Muted should be true")
	}
	if cm.Pinned {
		t.Error("Pinned should be false")
	}
}

func TestConversationType_Constants(t *testing.T) {
	if ConversationTypeDirect != 1 {
		t.Errorf("ConversationTypeDirect = %v, want %v", ConversationTypeDirect, 1)
	}
	if ConversationTypeGroup != 2 {
		t.Errorf("ConversationTypeGroup = %v, want %v", ConversationTypeGroup, 2)
	}
}

func TestGenerateDirectCid(t *testing.T) {
	tests := []struct {
		name     string
		uid1     string
		uid2     string
		expected string
	}{
		{
			name:     "uid1 < uid2",
			uid1:     "aaa",
			uid2:     "bbb",
			expected: "d:aaa:bbb",
		},
		{
			name:     "uid2 < uid1",
			uid1:     "bbb",
			uid2:     "aaa",
			expected: "d:aaa:bbb",
		},
		{
			name:     "same order regardless of input order",
			uid1:     "user123",
			uid2:     "user456",
			expected: "d:user123:user456",
		},
		{
			name:     "reversed input",
			uid1:     "user456",
			uid2:     "user123",
			expected: "d:user123:user456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateDirectCid(tt.uid1, tt.uid2)
			if result != tt.expected {
				t.Errorf("GenerateDirectCid(%s, %s) = %v, want %v", tt.uid1, tt.uid2, result, tt.expected)
			}
		})
	}
}

func TestGenerateDirectCid_Symmetry(t *testing.T) {
	// Regardless of order, same cid should be generated
	cid1 := GenerateDirectCid("userA", "userB")
	cid2 := GenerateDirectCid("userB", "userA")

	if cid1 != cid2 {
		t.Errorf("GenerateDirectCid should be symmetric: %s != %s", cid1, cid2)
	}
}

func TestGenerateGroupCid(t *testing.T) {
	tests := []struct {
		groupId  string
		expected string
	}{
		{"group123", "g:group123"},
		{"abc", "g:abc"},
		{"xyz789", "g:xyz789"},
	}

	for _, tt := range tests {
		result := GenerateGroupCid(tt.groupId)
		if result != tt.expected {
			t.Errorf("GenerateGroupCid(%s) = %v, want %v", tt.groupId, result, tt.expected)
		}
	}
}

func TestCidPrefix(t *testing.T) {
	directCid := GenerateDirectCid("user1", "user2")
	groupCid := GenerateGroupCid("group1")

	// Direct cid should start with "d:"
	if directCid[:2] != "d:" {
		t.Errorf("Direct cid should start with 'd:', got %s", directCid)
	}

	// Group cid should start with "g:"
	if groupCid[:2] != "g:" {
		t.Errorf("Group cid should start with 'g:', got %s", groupCid)
	}
}
