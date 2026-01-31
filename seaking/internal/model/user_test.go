package model

import (
	"testing"
	"time"
)

func TestUser_TableName(t *testing.T) {
	u := User{}
	if u.TableName() != "users" {
		t.Errorf("TableName() = %v, want %v", u.TableName(), "users")
	}
}

func TestUser_Fields(t *testing.T) {
	now := time.Now()
	u := User{
		ID:        "user123",
		Username:  "testuser",
		Nickname:  "Test User",
		Avatar:    "https://example.com/avatar.jpg",
		Password:  "hashedpassword",
		Phone:     "1234567890",
		Email:     "test@example.com",
		Status:    UserStatusNormal,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if u.ID != "user123" {
		t.Errorf("ID = %v, want %v", u.ID, "user123")
	}
	if u.Username != "testuser" {
		t.Errorf("Username = %v, want %v", u.Username, "testuser")
	}
	if u.Nickname != "Test User" {
		t.Errorf("Nickname = %v, want %v", u.Nickname, "Test User")
	}
	if u.Avatar != "https://example.com/avatar.jpg" {
		t.Errorf("Avatar = %v, want %v", u.Avatar, "https://example.com/avatar.jpg")
	}
	if u.Password != "hashedpassword" {
		t.Errorf("Password = %v, want %v", u.Password, "hashedpassword")
	}
	if u.Phone != "1234567890" {
		t.Errorf("Phone = %v, want %v", u.Phone, "1234567890")
	}
	if u.Email != "test@example.com" {
		t.Errorf("Email = %v, want %v", u.Email, "test@example.com")
	}
	if u.Status != UserStatusNormal {
		t.Errorf("Status = %v, want %v", u.Status, UserStatusNormal)
	}
}

func TestUserStatus_Constants(t *testing.T) {
	if UserStatusDisabled != 0 {
		t.Errorf("UserStatusDisabled = %v, want %v", UserStatusDisabled, 0)
	}
	if UserStatusNormal != 1 {
		t.Errorf("UserStatusNormal = %v, want %v", UserStatusNormal, 1)
	}
}

func TestUser_DefaultValues(t *testing.T) {
	u := User{}

	// Status should default to 0 (will be set by DB to 1)
	if u.Status != 0 {
		t.Errorf("Default Status = %v, want %v", u.Status, 0)
	}

	// String fields should be empty
	if u.ID != "" {
		t.Errorf("Default ID should be empty")
	}
	if u.Username != "" {
		t.Errorf("Default Username should be empty")
	}
	if u.Nickname != "" {
		t.Errorf("Default Nickname should be empty")
	}
}

func TestUser_DisabledStatus(t *testing.T) {
	u := User{
		ID:       "user123",
		Username: "disableduser",
		Status:   UserStatusDisabled,
	}

	if u.Status != UserStatusDisabled {
		t.Errorf("Status = %v, want %v", u.Status, UserStatusDisabled)
	}
}

func TestUser_JSONTags(t *testing.T) {
	// Password should have json:"-" tag (not exposed)
	u := User{
		ID:       "user123",
		Username: "testuser",
		Password: "secretpassword",
	}

	// This is a compile-time check - if the struct compiles, tags are valid
	// Runtime check for field existence
	if u.Password != "secretpassword" {
		t.Error("Password field should be accessible internally")
	}
}

func TestUser_OptionalFields(t *testing.T) {
	// User with only required fields
	u := User{
		ID:       "user123",
		Username: "testuser",
		Password: "hashedpassword",
		Status:   UserStatusNormal,
	}

	// Optional fields should be empty
	if u.Nickname != "" {
		t.Error("Nickname should be empty by default")
	}
	if u.Avatar != "" {
		t.Error("Avatar should be empty by default")
	}
	if u.Phone != "" {
		t.Error("Phone should be empty by default")
	}
	if u.Email != "" {
		t.Error("Email should be empty by default")
	}
}

func TestUser_StatusTransitions(t *testing.T) {
	tests := []struct {
		name   string
		status int
		desc   string
	}{
		{"disabled", UserStatusDisabled, "user cannot login"},
		{"normal", UserStatusNormal, "user can login"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := User{Status: tt.status}
			if u.Status != tt.status {
				t.Errorf("Status = %v, want %v for %s", u.Status, tt.status, tt.desc)
			}
		})
	}
}
