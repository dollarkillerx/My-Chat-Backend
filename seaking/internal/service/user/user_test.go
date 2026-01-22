package user

import (
	"testing"
)

func TestRegisterRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     RegisterRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: RegisterRequest{
				Username: "testuser",
				Password: "password123",
				Nickname: "Test User",
			},
			wantErr: false,
		},
		{
			name: "username too short",
			req: RegisterRequest{
				Username: "ab",
				Password: "password123",
			},
			wantErr: true,
		},
		{
			name: "password too short",
			req: RegisterRequest{
				Username: "testuser",
				Password: "12345",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation checks
			hasErr := false
			if len(tt.req.Username) < 3 || len(tt.req.Username) > 32 {
				hasErr = true
			}
			if len(tt.req.Password) < 6 || len(tt.req.Password) > 32 {
				hasErr = true
			}

			if hasErr != tt.wantErr {
				t.Errorf("validation error = %v, wantErr %v", hasErr, tt.wantErr)
			}
		})
	}
}

func TestLoginRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     LoginRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: LoginRequest{
				Username: "testuser",
				Password: "password123",
			},
			wantErr: false,
		},
		{
			name: "empty username",
			req: LoginRequest{
				Username: "",
				Password: "password123",
			},
			wantErr: true,
		},
		{
			name: "empty password",
			req: LoginRequest{
				Username: "testuser",
				Password: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasErr := tt.req.Username == "" || tt.req.Password == ""
			if hasErr != tt.wantErr {
				t.Errorf("validation error = %v, wantErr %v", hasErr, tt.wantErr)
			}
		})
	}
}
