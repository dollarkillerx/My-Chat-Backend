package crypto

import (
	"bytes"
	"testing"
)

func TestAESEncrypt(t *testing.T) {
	key, err := GenerateAESKey(32)
	if err != nil {
		t.Fatalf("GenerateAESKey failed: %v", err)
	}

	plaintext := []byte("Hello, World! This is a test message.")

	ciphertext, err := AESEncrypt(plaintext, key)
	if err != nil {
		t.Fatalf("AESEncrypt failed: %v", err)
	}

	if ciphertext == "" {
		t.Error("AESEncrypt returned empty ciphertext")
	}

	// Ciphertext should be different from plaintext
	if ciphertext == string(plaintext) {
		t.Error("Ciphertext should be different from plaintext")
	}
}

func TestAESDecrypt(t *testing.T) {
	key, _ := GenerateAESKey(32)
	plaintext := []byte("Hello, World! This is a test message.")

	ciphertext, _ := AESEncrypt(plaintext, key)

	decrypted, err := AESDecrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("AESDecrypt failed: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("Decrypted text mismatch: got %s, want %s", decrypted, plaintext)
	}
}

func TestAESEncryptDecrypt_EmptyPlaintext(t *testing.T) {
	key, _ := GenerateAESKey(32)
	plaintext := []byte("")

	ciphertext, err := AESEncrypt(plaintext, key)
	if err != nil {
		t.Fatalf("AESEncrypt failed for empty plaintext: %v", err)
	}

	decrypted, err := AESDecrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("AESDecrypt failed for empty plaintext: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Error("Decrypted empty text should be empty")
	}
}

func TestAESEncryptDecrypt_LargePlaintext(t *testing.T) {
	key, _ := GenerateAESKey(32)
	// Create a large plaintext (10KB)
	plaintext := make([]byte, 10*1024)
	for i := range plaintext {
		plaintext[i] = byte(i % 256)
	}

	ciphertext, err := AESEncrypt(plaintext, key)
	if err != nil {
		t.Fatalf("AESEncrypt failed for large plaintext: %v", err)
	}

	decrypted, err := AESDecrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("AESDecrypt failed for large plaintext: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Error("Decrypted large text should match original")
	}
}

func TestAESDecrypt_WrongKey(t *testing.T) {
	key1, _ := GenerateAESKey(32)
	key2, _ := GenerateAESKey(32)
	plaintext := []byte("test message")

	ciphertext, _ := AESEncrypt(plaintext, key1)

	// Decrypting with wrong key should fail
	_, err := AESDecrypt(ciphertext, key2)
	if err == nil {
		t.Error("AESDecrypt should fail with wrong key")
	}
}

func TestAESDecrypt_InvalidCiphertext(t *testing.T) {
	key, _ := GenerateAESKey(32)

	// Invalid base64
	_, err := AESDecrypt("not-valid-base64!!!", key)
	if err == nil {
		t.Error("AESDecrypt should fail with invalid base64")
	}

	// Too short ciphertext
	_, err = AESDecrypt("YWJj", key) // "abc" in base64
	if err == nil {
		t.Error("AESDecrypt should fail with too short ciphertext")
	}
}

func TestAESEncrypt_DifferentCiphertext(t *testing.T) {
	key, _ := GenerateAESKey(32)
	plaintext := []byte("test message")

	// AES-GCM should produce different ciphertext each time due to random nonce
	cipher1, _ := AESEncrypt(plaintext, key)
	cipher2, _ := AESEncrypt(plaintext, key)

	if cipher1 == cipher2 {
		t.Error("AES-GCM should produce different ciphertext for same plaintext")
	}

	// But both should decrypt to same plaintext
	decrypted1, _ := AESDecrypt(cipher1, key)
	decrypted2, _ := AESDecrypt(cipher2, key)

	if !bytes.Equal(decrypted1, decrypted2) {
		t.Error("Both ciphertexts should decrypt to same plaintext")
	}
}

func TestGenerateAESKey(t *testing.T) {
	tests := []struct {
		size    int
		wantErr bool
	}{
		{16, false},  // AES-128
		{24, false},  // AES-192
		{32, false},  // AES-256
		{15, true},   // Invalid
		{33, true},   // Invalid
		{0, true},    // Invalid
	}

	for _, tt := range tests {
		key, err := GenerateAESKey(tt.size)
		if tt.wantErr {
			if err == nil {
				t.Errorf("GenerateAESKey(%d) should fail", tt.size)
			}
		} else {
			if err != nil {
				t.Errorf("GenerateAESKey(%d) failed: %v", tt.size, err)
			}
			if len(key) != tt.size {
				t.Errorf("GenerateAESKey(%d) returned key of length %d", tt.size, len(key))
			}
		}
	}
}

func TestGenerateAESKey_Uniqueness(t *testing.T) {
	key1, _ := GenerateAESKey(32)
	key2, _ := GenerateAESKey(32)

	if bytes.Equal(key1, key2) {
		t.Error("Two generated AES keys should be different")
	}
}

func TestAESEncrypt_InvalidKeySize(t *testing.T) {
	invalidKey := []byte("short") // 5 bytes, invalid for AES
	plaintext := []byte("test")

	_, err := AESEncrypt(plaintext, invalidKey)
	if err == nil {
		t.Error("AESEncrypt should fail with invalid key size")
	}
}

// Benchmark tests
func BenchmarkAESEncrypt(b *testing.B) {
	key, _ := GenerateAESKey(32)
	plaintext := []byte("benchmark test message for AES encryption")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AESEncrypt(plaintext, key)
	}
}

func BenchmarkAESDecrypt(b *testing.B) {
	key, _ := GenerateAESKey(32)
	plaintext := []byte("benchmark test message for AES decryption")
	ciphertext, _ := AESEncrypt(plaintext, key)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AESDecrypt(ciphertext, key)
	}
}

func BenchmarkGenerateAESKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateAESKey(32)
	}
}
