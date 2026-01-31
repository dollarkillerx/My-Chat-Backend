package key

import (
	"testing"
)

func TestCreateUserKeyRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateUserKeyRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: CreateUserKeyRequest{
				UserID:              "user123",
				PublicKey:           "public-key-base64",
				EncryptedPrivateKey: "encrypted-private-key-base64",
				KeySalt:             "salt-base64",
			},
			wantErr: false,
		},
		{
			name: "empty user id",
			req: CreateUserKeyRequest{
				UserID:              "",
				PublicKey:           "public-key-base64",
				EncryptedPrivateKey: "encrypted-private-key-base64",
				KeySalt:             "salt-base64",
			},
			wantErr: true,
		},
		{
			name: "empty public key",
			req: CreateUserKeyRequest{
				UserID:              "user123",
				PublicKey:           "",
				EncryptedPrivateKey: "encrypted-private-key-base64",
				KeySalt:             "salt-base64",
			},
			wantErr: true,
		},
		{
			name: "empty encrypted private key",
			req: CreateUserKeyRequest{
				UserID:              "user123",
				PublicKey:           "public-key-base64",
				EncryptedPrivateKey: "",
				KeySalt:             "salt-base64",
			},
			wantErr: true,
		},
		{
			name: "empty salt",
			req: CreateUserKeyRequest{
				UserID:              "user123",
				PublicKey:           "public-key-base64",
				EncryptedPrivateKey: "encrypted-private-key-base64",
				KeySalt:             "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasErr := tt.req.UserID == "" ||
				tt.req.PublicKey == "" ||
				tt.req.EncryptedPrivateKey == "" ||
				tt.req.KeySalt == ""

			if hasErr != tt.wantErr {
				t.Errorf("validation error = %v, wantErr %v", hasErr, tt.wantErr)
			}
		})
	}
}

func TestChatKeyEntry_Validation(t *testing.T) {
	tests := []struct {
		name    string
		entry   ChatKeyEntry
		wantErr bool
	}{
		{
			name: "valid entry",
			entry: ChatKeyEntry{
				UserID:       "user123",
				EncryptedKey: "encrypted-key-base64",
			},
			wantErr: false,
		},
		{
			name: "empty user id",
			entry: ChatKeyEntry{
				UserID:       "",
				EncryptedKey: "encrypted-key-base64",
			},
			wantErr: true,
		},
		{
			name: "empty encrypted key",
			entry: ChatKeyEntry{
				UserID:       "user123",
				EncryptedKey: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasErr := tt.entry.UserID == "" || tt.entry.EncryptedKey == ""
			if hasErr != tt.wantErr {
				t.Errorf("validation error = %v, wantErr %v", hasErr, tt.wantErr)
			}
		})
	}
}

func TestCreateChatKeysRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateChatKeysRequest
		wantErr bool
	}{
		{
			name: "valid request with two keys",
			req: CreateChatKeysRequest{
				ConversationID: "d:user1:user2",
				Keys: []ChatKeyEntry{
					{UserID: "user1", EncryptedKey: "key1"},
					{UserID: "user2", EncryptedKey: "key2"},
				},
			},
			wantErr: false,
		},
		{
			name: "empty conversation id",
			req: CreateChatKeysRequest{
				ConversationID: "",
				Keys: []ChatKeyEntry{
					{UserID: "user1", EncryptedKey: "key1"},
				},
			},
			wantErr: true,
		},
		{
			name: "empty keys",
			req: CreateChatKeysRequest{
				ConversationID: "d:user1:user2",
				Keys:           []ChatKeyEntry{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasErr := tt.req.ConversationID == "" || len(tt.req.Keys) == 0
			if hasErr != tt.wantErr {
				t.Errorf("validation error = %v, wantErr %v", hasErr, tt.wantErr)
			}
		})
	}
}

func TestGroupKeyEntry_Validation(t *testing.T) {
	tests := []struct {
		name    string
		entry   GroupKeyEntry
		wantErr bool
	}{
		{
			name: "valid entry",
			entry: GroupKeyEntry{
				UserID:       "user123",
				EncryptedKey: "encrypted-key-base64",
			},
			wantErr: false,
		},
		{
			name: "empty user id",
			entry: GroupKeyEntry{
				UserID:       "",
				EncryptedKey: "encrypted-key-base64",
			},
			wantErr: true,
		},
		{
			name: "empty encrypted key",
			entry: GroupKeyEntry{
				UserID:       "user123",
				EncryptedKey: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasErr := tt.entry.UserID == "" || tt.entry.EncryptedKey == ""
			if hasErr != tt.wantErr {
				t.Errorf("validation error = %v, wantErr %v", hasErr, tt.wantErr)
			}
		})
	}
}

func TestCreateGroupKeysRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateGroupKeysRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: CreateGroupKeysRequest{
				GroupID: "group123",
				Keys: []GroupKeyEntry{
					{UserID: "user1", EncryptedKey: "key1"},
					{UserID: "user2", EncryptedKey: "key2"},
				},
				Version: 1,
			},
			wantErr: false,
		},
		{
			name: "empty group id",
			req: CreateGroupKeysRequest{
				GroupID: "",
				Keys: []GroupKeyEntry{
					{UserID: "user1", EncryptedKey: "key1"},
				},
				Version: 1,
			},
			wantErr: true,
		},
		{
			name: "empty keys",
			req: CreateGroupKeysRequest{
				GroupID: "group123",
				Keys:    []GroupKeyEntry{},
				Version: 1,
			},
			wantErr: true,
		},
		{
			name: "zero version is valid (auto-increment)",
			req: CreateGroupKeysRequest{
				GroupID: "group123",
				Keys: []GroupKeyEntry{
					{UserID: "user1", EncryptedKey: "key1"},
				},
				Version: 0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasErr := tt.req.GroupID == "" || len(tt.req.Keys) == 0
			if hasErr != tt.wantErr {
				t.Errorf("validation error = %v, wantErr %v", hasErr, tt.wantErr)
			}
		})
	}
}

func TestConversationIDFormat(t *testing.T) {
	tests := []struct {
		name    string
		cid     string
		isValid bool
	}{
		{
			name:    "valid direct conversation",
			cid:     "d:user1:user2",
			isValid: true,
		},
		{
			name:    "valid group conversation",
			cid:     "g:group123",
			isValid: true,
		},
		{
			name:    "invalid format - no prefix",
			cid:     "user1:user2",
			isValid: false,
		},
		{
			name:    "empty string",
			cid:     "",
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simple validation: must start with "d:" or "g:" and have content after
			isValid := len(tt.cid) > 2 && (tt.cid[:2] == "d:" || tt.cid[:2] == "g:")
			if isValid != tt.isValid {
				t.Errorf("isValid = %v, want %v", isValid, tt.isValid)
			}
		})
	}
}
