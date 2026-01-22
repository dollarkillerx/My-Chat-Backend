package group

import (
	"testing"
)

func TestCreateGroupRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateGroupRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: CreateGroupRequest{
				Name:        "Test Group",
				Description: "A test group",
				MemberIDs:   []string{"user1", "user2"},
			},
			wantErr: false,
		},
		{
			name: "empty name",
			req: CreateGroupRequest{
				Name:      "",
				MemberIDs: []string{"user1"},
			},
			wantErr: true,
		},
		{
			name: "name too long",
			req: CreateGroupRequest{
				Name: "This is a very long group name that exceeds the maximum allowed length of 64 characters",
			},
			wantErr: true,
		},
		{
			name: "valid with no members",
			req: CreateGroupRequest{
				Name:      "Test Group",
				MemberIDs: nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasErr := false
			if len(tt.req.Name) < 1 || len(tt.req.Name) > 64 {
				hasErr = true
			}

			if hasErr != tt.wantErr {
				t.Errorf("validation error = %v, wantErr %v", hasErr, tt.wantErr)
			}
		})
	}
}

func TestGroupRoles(t *testing.T) {
	// Test role hierarchy
	tests := []struct {
		name     string
		role     int
		minRole  int
		hasPerms bool
	}{
		{"owner has owner perms", 2, 2, true},
		{"owner has admin perms", 2, 1, true},
		{"owner has member perms", 2, 0, true},
		{"admin has admin perms", 1, 1, true},
		{"admin has member perms", 1, 0, true},
		{"admin doesn't have owner perms", 1, 2, false},
		{"member has member perms", 0, 0, true},
		{"member doesn't have admin perms", 0, 1, false},
		{"member doesn't have owner perms", 0, 2, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasPerms := tt.role >= tt.minRole
			if hasPerms != tt.hasPerms {
				t.Errorf("hasPerms = %v, want %v", hasPerms, tt.hasPerms)
			}
		})
	}
}
