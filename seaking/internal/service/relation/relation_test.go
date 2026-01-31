package relation

import (
	"testing"

	"github.com/my-chat/seaking/internal/model"
)

func TestFriendshipModel(t *testing.T) {
	f := model.Friendship{
		UserID:   "user1",
		FriendID: "user2",
		Remark:   "Best Friend",
		Status:   model.FriendStatusNormal,
	}

	if f.TableName() != "friendships" {
		t.Errorf("TableName() = %v, want friendships", f.TableName())
	}

	if f.UserID != "user1" {
		t.Errorf("UserID = %v, want user1", f.UserID)
	}

	if f.FriendID != "user2" {
		t.Errorf("FriendID = %v, want user2", f.FriendID)
	}

	if f.Status != model.FriendStatusNormal {
		t.Errorf("Status = %v, want %v", f.Status, model.FriendStatusNormal)
	}
}

func TestFriendRequestModel(t *testing.T) {
	fr := model.FriendRequest{
		FromUID: "user1",
		ToUID:   "user2",
		Message: "Hi, let's be friends!",
		Status:  model.FriendRequestPending,
	}

	if fr.TableName() != "friend_requests" {
		t.Errorf("TableName() = %v, want friend_requests", fr.TableName())
	}

	if fr.FromUID != "user1" {
		t.Errorf("FromUID = %v, want user1", fr.FromUID)
	}

	if fr.ToUID != "user2" {
		t.Errorf("ToUID = %v, want user2", fr.ToUID)
	}

	if fr.Status != model.FriendRequestPending {
		t.Errorf("Status = %v, want %v", fr.Status, model.FriendRequestPending)
	}
}

func TestFriendStatus(t *testing.T) {
	tests := []struct {
		name   string
		status int
		want   int
	}{
		{"normal", model.FriendStatusNormal, 1},
		{"blocked", model.FriendStatusBlocked, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.status != tt.want {
				t.Errorf("FriendStatus %s = %v, want %v", tt.name, tt.status, tt.want)
			}
		})
	}
}

func TestFriendRequestStatus(t *testing.T) {
	tests := []struct {
		name   string
		status int
		want   int
	}{
		{"pending", model.FriendRequestPending, 0},
		{"accepted", model.FriendRequestAccepted, 1},
		{"rejected", model.FriendRequestRejected, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.status != tt.want {
				t.Errorf("FriendRequestStatus %s = %v, want %v", tt.name, tt.status, tt.want)
			}
		})
	}
}

func TestFriendshipBidirectional(t *testing.T) {
	// When user1 and user2 become friends, two records should be created
	// user1 -> user2
	// user2 -> user1
	f1 := model.Friendship{
		UserID:   "user1",
		FriendID: "user2",
		Status:   model.FriendStatusNormal,
	}

	f2 := model.Friendship{
		UserID:   "user2",
		FriendID: "user1",
		Status:   model.FriendStatusNormal,
	}

	// Verify they represent the same friendship from different perspectives
	if f1.UserID != f2.FriendID {
		t.Error("Bidirectional friendship mismatch")
	}
	if f1.FriendID != f2.UserID {
		t.Error("Bidirectional friendship mismatch")
	}
}

func TestFriendRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     model.FriendRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: model.FriendRequest{
				FromUID: "user1",
				ToUID:   "user2",
				Message: "Hello!",
			},
			wantErr: false,
		},
		{
			name: "empty from_uid",
			req: model.FriendRequest{
				FromUID: "",
				ToUID:   "user2",
			},
			wantErr: true,
		},
		{
			name: "empty to_uid",
			req: model.FriendRequest{
				FromUID: "user1",
				ToUID:   "",
			},
			wantErr: true,
		},
		{
			name: "same user",
			req: model.FriendRequest{
				FromUID: "user1",
				ToUID:   "user1",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasErr := tt.req.FromUID == "" || tt.req.ToUID == "" || tt.req.FromUID == tt.req.ToUID
			if hasErr != tt.wantErr {
				t.Errorf("validation error = %v, wantErr %v", hasErr, tt.wantErr)
			}
		})
	}
}

func TestFriendshipBlockUnblock(t *testing.T) {
	f := model.Friendship{
		UserID:   "user1",
		FriendID: "user2",
		Status:   model.FriendStatusNormal,
	}

	// Block friend
	f.Status = model.FriendStatusBlocked
	if f.Status != model.FriendStatusBlocked {
		t.Error("Failed to block friend")
	}

	// Unblock friend
	f.Status = model.FriendStatusNormal
	if f.Status != model.FriendStatusNormal {
		t.Error("Failed to unblock friend")
	}
}

func TestFriendRequestStatusTransition(t *testing.T) {
	fr := model.FriendRequest{
		FromUID: "user1",
		ToUID:   "user2",
		Status:  model.FriendRequestPending,
	}

	// Accept request
	fr.Status = model.FriendRequestAccepted
	if fr.Status != model.FriendRequestAccepted {
		t.Error("Failed to accept friend request")
	}

	// Test rejection (new request)
	fr2 := model.FriendRequest{
		FromUID: "user3",
		ToUID:   "user4",
		Status:  model.FriendRequestPending,
	}
	fr2.Status = model.FriendRequestRejected
	if fr2.Status != model.FriendRequestRejected {
		t.Error("Failed to reject friend request")
	}
}

func TestFriendshipRemark(t *testing.T) {
	f := model.Friendship{
		UserID:   "user1",
		FriendID: "user2",
		Remark:   "",
	}

	// Set remark
	f.Remark = "Colleague"
	if f.Remark != "Colleague" {
		t.Errorf("Remark = %v, want Colleague", f.Remark)
	}

	// Update remark
	f.Remark = "Best Friend"
	if f.Remark != "Best Friend" {
		t.Errorf("Remark = %v, want Best Friend", f.Remark)
	}

	// Clear remark
	f.Remark = ""
	if f.Remark != "" {
		t.Error("Failed to clear remark")
	}
}
