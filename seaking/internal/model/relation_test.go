package model

import (
	"testing"
	"time"
)

func TestFriendship_TableName(t *testing.T) {
	f := Friendship{}
	if f.TableName() != "friendships" {
		t.Errorf("TableName() = %v, want %v", f.TableName(), "friendships")
	}
}

func TestFriendRequest_TableName(t *testing.T) {
	fr := FriendRequest{}
	if fr.TableName() != "friend_requests" {
		t.Errorf("TableName() = %v, want %v", fr.TableName(), "friend_requests")
	}
}

func TestFriendship_Fields(t *testing.T) {
	now := time.Now()
	f := Friendship{
		ID:        1,
		UserID:    "user1",
		FriendID:  "user2",
		Remark:    "Best Friend",
		Status:    FriendStatusNormal,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if f.ID != 1 {
		t.Errorf("ID = %v, want %v", f.ID, 1)
	}
	if f.UserID != "user1" {
		t.Errorf("UserID = %v, want %v", f.UserID, "user1")
	}
	if f.FriendID != "user2" {
		t.Errorf("FriendID = %v, want %v", f.FriendID, "user2")
	}
	if f.Remark != "Best Friend" {
		t.Errorf("Remark = %v, want %v", f.Remark, "Best Friend")
	}
	if f.Status != FriendStatusNormal {
		t.Errorf("Status = %v, want %v", f.Status, FriendStatusNormal)
	}
}

func TestFriendRequest_Fields(t *testing.T) {
	now := time.Now()
	fr := FriendRequest{
		ID:        1,
		FromUID:   "user1",
		ToUID:     "user2",
		Message:   "Hi, let's be friends!",
		Status:    FriendRequestPending,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if fr.ID != 1 {
		t.Errorf("ID = %v, want %v", fr.ID, 1)
	}
	if fr.FromUID != "user1" {
		t.Errorf("FromUID = %v, want %v", fr.FromUID, "user1")
	}
	if fr.ToUID != "user2" {
		t.Errorf("ToUID = %v, want %v", fr.ToUID, "user2")
	}
	if fr.Message != "Hi, let's be friends!" {
		t.Errorf("Message = %v, want %v", fr.Message, "Hi, let's be friends!")
	}
	if fr.Status != FriendRequestPending {
		t.Errorf("Status = %v, want %v", fr.Status, FriendRequestPending)
	}
}

func TestFriendStatus_Constants(t *testing.T) {
	if FriendStatusNormal != 1 {
		t.Errorf("FriendStatusNormal = %v, want %v", FriendStatusNormal, 1)
	}
	if FriendStatusBlocked != 2 {
		t.Errorf("FriendStatusBlocked = %v, want %v", FriendStatusBlocked, 2)
	}
}

func TestFriendRequestStatus_Constants(t *testing.T) {
	if FriendRequestPending != 0 {
		t.Errorf("FriendRequestPending = %v, want %v", FriendRequestPending, 0)
	}
	if FriendRequestAccepted != 1 {
		t.Errorf("FriendRequestAccepted = %v, want %v", FriendRequestAccepted, 1)
	}
	if FriendRequestRejected != 2 {
		t.Errorf("FriendRequestRejected = %v, want %v", FriendRequestRejected, 2)
	}
}

func TestFriendship_DefaultValues(t *testing.T) {
	f := Friendship{}

	// Status should default to 0 (will be set by DB to 1)
	if f.Status != 0 {
		t.Errorf("Default Status = %v, want %v", f.Status, 0)
	}
}

func TestFriendRequest_DefaultValues(t *testing.T) {
	fr := FriendRequest{}

	// Status should default to 0 (Pending)
	if fr.Status != 0 {
		t.Errorf("Default Status = %v, want %v", fr.Status, 0)
	}
}

func TestFriendship_BlockedStatus(t *testing.T) {
	f := Friendship{
		UserID:   "user1",
		FriendID: "user2",
		Status:   FriendStatusBlocked,
	}

	if f.Status != FriendStatusBlocked {
		t.Errorf("Status = %v, want %v", f.Status, FriendStatusBlocked)
	}
}

func TestFriendRequest_StatusTransitions(t *testing.T) {
	tests := []struct {
		name   string
		status int
		desc   string
	}{
		{"pending", FriendRequestPending, "waiting for response"},
		{"accepted", FriendRequestAccepted, "friendship established"},
		{"rejected", FriendRequestRejected, "request declined"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fr := FriendRequest{Status: tt.status}
			if fr.Status != tt.status {
				t.Errorf("Status = %v, want %v for %s", fr.Status, tt.status, tt.desc)
			}
		})
	}
}
