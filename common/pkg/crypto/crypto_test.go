package crypto

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "testPassword123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hash == "" {
		t.Error("HashPassword returned empty hash")
	}

	if hash == password {
		t.Error("HashPassword returned plaintext password")
	}
}

func TestCheckPassword(t *testing.T) {
	password := "testPassword123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Correct password should pass
	if !CheckPassword(password, hash) {
		t.Error("CheckPassword failed for correct password")
	}

	// Wrong password should fail
	if CheckPassword("wrongPassword", hash) {
		t.Error("CheckPassword passed for wrong password")
	}
}

func TestHashPasswordDifferentHashes(t *testing.T) {
	password := "testPassword123"

	hash1, _ := HashPassword(password)
	hash2, _ := HashPassword(password)

	// bcrypt should generate different hashes for same password
	if hash1 == hash2 {
		t.Error("HashPassword generated same hash for same password (bcrypt should use random salt)")
	}

	// But both should verify correctly
	if !CheckPassword(password, hash1) {
		t.Error("First hash verification failed")
	}
	if !CheckPassword(password, hash2) {
		t.Error("Second hash verification failed")
	}
}

func TestCheckPasswordEmptyInputs(t *testing.T) {
	// Empty password with valid hash
	hash, _ := HashPassword("validPassword")
	if CheckPassword("", hash) {
		t.Error("CheckPassword should fail for empty password")
	}

	// Valid password with empty hash
	if CheckPassword("validPassword", "") {
		t.Error("CheckPassword should fail for empty hash")
	}
}

func TestSHA256Hash(t *testing.T) {
	data := []byte("hello world")
	hash := SHA256Hash(data)

	if hash == "" {
		t.Error("SHA256Hash returned empty string")
	}

	// SHA256 hash should be 64 hex characters
	if len(hash) != 64 {
		t.Errorf("SHA256Hash length mismatch: got %d, want 64", len(hash))
	}

	// Same data should produce same hash
	hash2 := SHA256Hash(data)
	if hash != hash2 {
		t.Error("SHA256Hash should produce same hash for same data")
	}

	// Different data should produce different hash
	hash3 := SHA256Hash([]byte("different data"))
	if hash == hash3 {
		t.Error("SHA256Hash should produce different hash for different data")
	}
}

func TestHMACSHA256(t *testing.T) {
	data := []byte("hello world")
	key := []byte("secret-key")

	signature := HMACSHA256(data, key)

	if signature == "" {
		t.Error("HMACSHA256 returned empty string")
	}

	// Same data and key should produce same signature
	signature2 := HMACSHA256(data, key)
	if signature != signature2 {
		t.Error("HMACSHA256 should produce same signature for same inputs")
	}

	// Different key should produce different signature
	signature3 := HMACSHA256(data, []byte("different-key"))
	if signature == signature3 {
		t.Error("HMACSHA256 should produce different signature for different key")
	}
}

func TestVerifyHMACSHA256(t *testing.T) {
	data := []byte("hello world")
	key := []byte("secret-key")

	signature := HMACSHA256(data, key)

	// Correct signature should verify
	if !VerifyHMACSHA256(data, key, signature) {
		t.Error("VerifyHMACSHA256 failed for correct signature")
	}

	// Wrong signature should not verify
	if VerifyHMACSHA256(data, key, "wrong-signature") {
		t.Error("VerifyHMACSHA256 passed for wrong signature")
	}

	// Wrong key should not verify
	if VerifyHMACSHA256(data, []byte("wrong-key"), signature) {
		t.Error("VerifyHMACSHA256 passed for wrong key")
	}

	// Wrong data should not verify
	if VerifyHMACSHA256([]byte("wrong data"), key, signature) {
		t.Error("VerifyHMACSHA256 passed for wrong data")
	}
}
