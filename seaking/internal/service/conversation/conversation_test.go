package conversation

import (
	"testing"

	"github.com/my-chat/seaking/internal/model"
)

func TestConversationCidGeneration(t *testing.T) {
	// Test direct conversation cid generation
	tests := []struct {
		name     string
		uid1     string
		uid2     string
		expected string
	}{
		{
			name:     "uid1 < uid2",
			uid1:     "alice",
			uid2:     "bob",
			expected: "d:alice:bob",
		},
		{
			name:     "uid2 < uid1",
			uid1:     "bob",
			uid2:     "alice",
			expected: "d:alice:bob",
		},
		{
			name:     "same result regardless of order",
			uid1:     "user123",
			uid2:     "user456",
			expected: "d:user123:user456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cid := model.GenerateDirectCid(tt.uid1, tt.uid2)
			if cid != tt.expected {
				t.Errorf("GenerateDirectCid(%s, %s) = %v, want %v", tt.uid1, tt.uid2, cid, tt.expected)
			}
		})
	}
}

func TestGroupConversationCidGeneration(t *testing.T) {
	tests := []struct {
		groupId  string
		expected string
	}{
		{"group1", "g:group1"},
		{"abc123", "g:abc123"},
		{"test-group", "g:test-group"},
	}

	for _, tt := range tests {
		cid := model.GenerateGroupCid(tt.groupId)
		if cid != tt.expected {
			t.Errorf("GenerateGroupCid(%s) = %v, want %v", tt.groupId, cid, tt.expected)
		}
	}
}

func TestConversationCidSymmetry(t *testing.T) {
	// Direct conversation cid should be the same regardless of user order
	cid1 := model.GenerateDirectCid("userA", "userB")
	cid2 := model.GenerateDirectCid("userB", "userA")

	if cid1 != cid2 {
		t.Errorf("Direct cid should be symmetric: %s != %s", cid1, cid2)
	}
}

func TestConversationTypes(t *testing.T) {
	if model.ConversationTypeDirect != 1 {
		t.Errorf("ConversationTypeDirect = %d, want 1", model.ConversationTypeDirect)
	}
	if model.ConversationTypeGroup != 2 {
		t.Errorf("ConversationTypeGroup = %d, want 2", model.ConversationTypeGroup)
	}
}

func TestConversationModel(t *testing.T) {
	conv := model.Conversation{
		ID:   "d:user1:user2",
		Type: model.ConversationTypeDirect,
		Name: "Test Chat",
	}

	if conv.TableName() != "conversations" {
		t.Errorf("TableName() = %v, want conversations", conv.TableName())
	}

	if conv.ID != "d:user1:user2" {
		t.Errorf("ID = %v, want d:user1:user2", conv.ID)
	}
}

func TestConversationMemberModel(t *testing.T) {
	member := model.ConversationMember{
		ConversationID: "d:user1:user2",
		UserID:         "user1",
		LastReadMid:    100,
		Muted:          false,
		Pinned:         true,
	}

	if member.TableName() != "conversation_members" {
		t.Errorf("TableName() = %v, want conversation_members", member.TableName())
	}

	if member.ConversationID != "d:user1:user2" {
		t.Errorf("ConversationID = %v, want d:user1:user2", member.ConversationID)
	}

	if member.UserID != "user1" {
		t.Errorf("UserID = %v, want user1", member.UserID)
	}

	if member.LastReadMid != 100 {
		t.Errorf("LastReadMid = %v, want 100", member.LastReadMid)
	}

	if member.Muted {
		t.Error("Muted should be false")
	}

	if !member.Pinned {
		t.Error("Pinned should be true")
	}
}

func TestCidParsing(t *testing.T) {
	tests := []struct {
		cid      string
		isGroup  bool
		isDirect bool
	}{
		{"d:user1:user2", false, true},
		{"g:group123", true, false},
		{"d:abc:xyz", false, true},
		{"g:mygroup", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.cid, func(t *testing.T) {
			isGroup := len(tt.cid) > 2 && tt.cid[:2] == "g:"
			isDirect := len(tt.cid) > 2 && tt.cid[:2] == "d:"

			if isGroup != tt.isGroup {
				t.Errorf("cid %s: isGroup = %v, want %v", tt.cid, isGroup, tt.isGroup)
			}
			if isDirect != tt.isDirect {
				t.Errorf("cid %s: isDirect = %v, want %v", tt.cid, isDirect, tt.isDirect)
			}
		})
	}
}
