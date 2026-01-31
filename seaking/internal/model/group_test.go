package model

import (
	"testing"
	"time"
)

func TestGroup_TableName(t *testing.T) {
	g := Group{}
	if g.TableName() != "groups" {
		t.Errorf("TableName() = %v, want %v", g.TableName(), "groups")
	}
}

func TestGroupMember_TableName(t *testing.T) {
	gm := GroupMember{}
	if gm.TableName() != "group_members" {
		t.Errorf("TableName() = %v, want %v", gm.TableName(), "group_members")
	}
}

func TestGroup_Fields(t *testing.T) {
	now := time.Now()
	g := Group{
		ID:          "group123",
		Name:        "Test Group",
		Avatar:      "https://example.com/group.jpg",
		Description: "This is a test group",
		OwnerID:     "owner123",
		MaxMembers:  500,
		Status:      GroupStatusNormal,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if g.ID != "group123" {
		t.Errorf("ID = %v, want %v", g.ID, "group123")
	}
	if g.Name != "Test Group" {
		t.Errorf("Name = %v, want %v", g.Name, "Test Group")
	}
	if g.OwnerID != "owner123" {
		t.Errorf("OwnerID = %v, want %v", g.OwnerID, "owner123")
	}
	if g.MaxMembers != 500 {
		t.Errorf("MaxMembers = %v, want %v", g.MaxMembers, 500)
	}
	if g.Status != GroupStatusNormal {
		t.Errorf("Status = %v, want %v", g.Status, GroupStatusNormal)
	}
}

func TestGroupMember_Fields(t *testing.T) {
	now := time.Now()
	gm := GroupMember{
		ID:        1,
		GroupID:   "group123",
		UserID:    "user456",
		Role:      GroupRoleAdmin,
		Nickname:  "Admin User",
		Muted:     false,
		MutedAt:   nil,
		JoinedAt:  now,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if gm.ID != 1 {
		t.Errorf("ID = %v, want %v", gm.ID, 1)
	}
	if gm.GroupID != "group123" {
		t.Errorf("GroupID = %v, want %v", gm.GroupID, "group123")
	}
	if gm.UserID != "user456" {
		t.Errorf("UserID = %v, want %v", gm.UserID, "user456")
	}
	if gm.Role != GroupRoleAdmin {
		t.Errorf("Role = %v, want %v", gm.Role, GroupRoleAdmin)
	}
	if gm.Nickname != "Admin User" {
		t.Errorf("Nickname = %v, want %v", gm.Nickname, "Admin User")
	}
	if gm.Muted {
		t.Error("Muted should be false")
	}
	if gm.MutedAt != nil {
		t.Error("MutedAt should be nil")
	}
}

func TestGroupMember_MutedAt(t *testing.T) {
	now := time.Now()
	gm := GroupMember{
		ID:      1,
		GroupID: "group123",
		UserID:  "user456",
		Muted:   true,
		MutedAt: &now,
	}

	if !gm.Muted {
		t.Error("Muted should be true")
	}
	if gm.MutedAt == nil {
		t.Error("MutedAt should not be nil")
	}
	if !gm.MutedAt.Equal(now) {
		t.Error("MutedAt should equal the set time")
	}
}

func TestGroupStatus_Constants(t *testing.T) {
	if GroupStatusDissolved != 0 {
		t.Errorf("GroupStatusDissolved = %v, want %v", GroupStatusDissolved, 0)
	}
	if GroupStatusNormal != 1 {
		t.Errorf("GroupStatusNormal = %v, want %v", GroupStatusNormal, 1)
	}
}

func TestGroupRole_Constants(t *testing.T) {
	if GroupRoleMember != 0 {
		t.Errorf("GroupRoleMember = %v, want %v", GroupRoleMember, 0)
	}
	if GroupRoleAdmin != 1 {
		t.Errorf("GroupRoleAdmin = %v, want %v", GroupRoleAdmin, 1)
	}
	if GroupRoleOwner != 2 {
		t.Errorf("GroupRoleOwner = %v, want %v", GroupRoleOwner, 2)
	}
}

func TestGroupRole_Hierarchy(t *testing.T) {
	// Owner > Admin > Member
	if GroupRoleOwner <= GroupRoleAdmin {
		t.Error("GroupRoleOwner should be greater than GroupRoleAdmin")
	}
	if GroupRoleAdmin <= GroupRoleMember {
		t.Error("GroupRoleAdmin should be greater than GroupRoleMember")
	}
}

func TestGroup_DefaultValues(t *testing.T) {
	g := Group{}

	// Status should default to 0 (will be set by DB to 1)
	if g.Status != 0 {
		t.Errorf("Default Status = %v, want %v", g.Status, 0)
	}

	// MaxMembers should default to 0 (will be set by DB to 500)
	if g.MaxMembers != 0 {
		t.Errorf("Default MaxMembers = %v, want %v", g.MaxMembers, 0)
	}
}

func TestGroupMember_DefaultValues(t *testing.T) {
	gm := GroupMember{}

	// Role should default to 0 (Member)
	if gm.Role != 0 {
		t.Errorf("Default Role = %v, want %v", gm.Role, 0)
	}

	// Muted should default to false
	if gm.Muted != false {
		t.Error("Default Muted should be false")
	}
}
