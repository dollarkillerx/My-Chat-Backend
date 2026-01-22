package errors

import (
	"testing"
)

func TestNew(t *testing.T) {
	err := New(1001, "test error message")

	if err == nil {
		t.Fatal("New returned nil")
	}

	if err.Code != 1001 {
		t.Errorf("Code mismatch: got %d, want 1001", err.Code)
	}

	if err.Message != "test error message" {
		t.Errorf("Message mismatch: got %s, want test error message", err.Message)
	}
}

func TestErrorString(t *testing.T) {
	err := New(1001, "test error message")

	str := err.Error()
	if str == "" {
		t.Error("Error() returned empty string")
	}

	// Should contain the message (format may include code)
	if str != "[1001] test error message" && str != "test error message" {
		t.Errorf("Error() should contain the message, got: %s", str)
	}
}

func TestPreDefinedErrors(t *testing.T) {
	testCases := []struct {
		name    string
		err     *Error
		code    int
		hasMsg  bool
	}{
		{"ErrInvalidParam", ErrInvalidParam, ErrCodeInvalidParam, true},
		{"ErrInternal", ErrInternal, ErrCodeInternal, true},
		{"ErrNotFound", ErrNotFound, ErrCodeNotFound, true},
		{"ErrForbidden", ErrForbidden, ErrCodeForbidden, true},
		{"ErrLoginRequired", ErrLoginRequired, ErrCodeLoginRequired, true},
		{"ErrInvalidToken", ErrInvalidToken, ErrCodeInvalidToken, true},
		{"ErrUserNotFound", ErrUserNotFound, ErrCodeUserNotFound, true},
		{"ErrUserExists", ErrUserExists, ErrCodeUserExists, true},
		{"ErrPasswordWrong", ErrPasswordWrong, ErrCodePasswordWrong, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err == nil {
				t.Errorf("%s is nil", tc.name)
				return
			}

			if tc.err.Code != tc.code {
				t.Errorf("%s code mismatch: got %d, want %d", tc.name, tc.err.Code, tc.code)
			}

			if tc.hasMsg && tc.err.Message == "" {
				t.Errorf("%s has empty message", tc.name)
			}
		})
	}
}

func TestErrorCodes(t *testing.T) {
	// Verify error codes are unique
	codes := map[int]string{
		ErrCodeUnknown:           "ErrCodeUnknown",
		ErrCodeInvalidParam:      "ErrCodeInvalidParam",
		ErrCodeInternal:          "ErrCodeInternal",
		ErrCodeNotFound:          "ErrCodeNotFound",
		ErrCodeForbidden:         "ErrCodeForbidden",
		ErrCodeLoginRequired:     "ErrCodeLoginRequired",
		ErrCodeInvalidToken:      "ErrCodeInvalidToken",
		ErrCodeUserNotFound:      "ErrCodeUserNotFound",
		ErrCodeUserExists:        "ErrCodeUserExists",
		ErrCodePasswordWrong:     "ErrCodePasswordWrong",
		ErrCodeUserDisabled:      "ErrCodeUserDisabled",
		ErrCodeNotInConversation: "ErrCodeNotInConversation",
		ErrCodeConversationFull:  "ErrCodeConversationFull",
		ErrCodeAlreadyFriend:     "ErrCodeAlreadyFriend",
		ErrCodeGroupNotFound:     "ErrCodeGroupNotFound",
		ErrCodeNoPermission:      "ErrCodeNoPermission",
		ErrCodeCannotRevoke:      "ErrCodeCannotRevoke",
		ErrCodeCannotEdit:        "ErrCodeCannotEdit",
	}

	// Check that we have the expected number of unique codes
	if len(codes) < 10 {
		t.Error("Expected at least 10 unique error codes")
	}
}

func TestErrorImplementsErrorInterface(t *testing.T) {
	var err error = New(1001, "test")

	if err == nil {
		t.Error("Error should implement error interface")
	}

	_ = err.Error() // Should not panic
}
