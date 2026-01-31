package crypto

import (
	"testing"
)

func TestGenerateRSAKeyPair(t *testing.T) {
	publicKey, privateKey, err := GenerateRSAKeyPair()
	if err != nil {
		t.Fatalf("GenerateRSAKeyPair failed: %v", err)
	}

	if publicKey == "" {
		t.Error("GenerateRSAKeyPair returned empty public key")
	}

	if privateKey == "" {
		t.Error("GenerateRSAKeyPair returned empty private key")
	}

	// Keys should be different
	if publicKey == privateKey {
		t.Error("Public and private keys should be different")
	}
}

func TestGenerateRSAKeyPair_Uniqueness(t *testing.T) {
	pub1, priv1, _ := GenerateRSAKeyPair()
	pub2, priv2, _ := GenerateRSAKeyPair()

	if pub1 == pub2 {
		t.Error("Two generated public keys should be different")
	}

	if priv1 == priv2 {
		t.Error("Two generated private keys should be different")
	}
}

func TestRSAEncryptDecrypt(t *testing.T) {
	publicKey, privateKey, err := GenerateRSAKeyPair()
	if err != nil {
		t.Fatalf("GenerateRSAKeyPair failed: %v", err)
	}

	plaintext := []byte("Hello, World! This is a test message for RSA encryption.")

	// Encrypt
	ciphertext, err := RSAEncrypt(publicKey, plaintext)
	if err != nil {
		t.Fatalf("RSAEncrypt failed: %v", err)
	}

	if ciphertext == "" {
		t.Error("RSAEncrypt returned empty ciphertext")
	}

	// Decrypt
	decrypted, err := RSADecrypt(privateKey, ciphertext)
	if err != nil {
		t.Fatalf("RSADecrypt failed: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("Decrypted text mismatch: got %s, want %s", decrypted, plaintext)
	}
}

func TestRSAEncrypt_DifferentCiphertext(t *testing.T) {
	publicKey, _, _ := GenerateRSAKeyPair()
	plaintext := []byte("test message")

	// RSA-OAEP should produce different ciphertext each time due to random padding
	cipher1, _ := RSAEncrypt(publicKey, plaintext)
	cipher2, _ := RSAEncrypt(publicKey, plaintext)

	if cipher1 == cipher2 {
		t.Error("RSA-OAEP should produce different ciphertext for same plaintext")
	}
}

func TestRSADecrypt_WrongKey(t *testing.T) {
	pub1, _, _ := GenerateRSAKeyPair()
	_, priv2, _ := GenerateRSAKeyPair()

	plaintext := []byte("test message")
	ciphertext, _ := RSAEncrypt(pub1, plaintext)

	// Decrypting with wrong private key should fail
	_, err := RSADecrypt(priv2, ciphertext)
	if err == nil {
		t.Error("RSADecrypt should fail with wrong private key")
	}
}

func TestRSAEncrypt_InvalidPublicKey(t *testing.T) {
	_, err := RSAEncrypt("invalid-key", []byte("test"))
	if err == nil {
		t.Error("RSAEncrypt should fail with invalid public key")
	}
}

func TestRSADecrypt_InvalidPrivateKey(t *testing.T) {
	_, err := RSADecrypt("invalid-key", "invalid-ciphertext")
	if err == nil {
		t.Error("RSADecrypt should fail with invalid private key")
	}
}

func TestGenerateSalt(t *testing.T) {
	salt1, err := GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt failed: %v", err)
	}

	if salt1 == "" {
		t.Error("GenerateSalt returned empty string")
	}

	// Generate another salt and verify uniqueness
	salt2, _ := GenerateSalt()
	if salt1 == salt2 {
		t.Error("Two generated salts should be different")
	}
}

func TestDeriveKey(t *testing.T) {
	password := "testPassword123"
	salt, _ := GenerateSalt()

	key1, err := DeriveKey(password, salt)
	if err != nil {
		t.Fatalf("DeriveKey failed: %v", err)
	}

	if len(key1) != DerivedKeySize {
		t.Errorf("DeriveKey length mismatch: got %d, want %d", len(key1), DerivedKeySize)
	}

	// Same password and salt should produce same key
	key2, _ := DeriveKey(password, salt)
	if string(key1) != string(key2) {
		t.Error("DeriveKey should produce same key for same inputs")
	}

	// Different password should produce different key
	key3, _ := DeriveKey("differentPassword", salt)
	if string(key1) == string(key3) {
		t.Error("DeriveKey should produce different key for different password")
	}

	// Different salt should produce different key
	salt2, _ := GenerateSalt()
	key4, _ := DeriveKey(password, salt2)
	if string(key1) == string(key4) {
		t.Error("DeriveKey should produce different key for different salt")
	}
}

func TestEncryptDecryptPrivateKey(t *testing.T) {
	_, privateKey, _ := GenerateRSAKeyPair()
	password := "testPassword123"

	// Encrypt private key
	encryptedKey, salt, err := EncryptPrivateKey(privateKey, password)
	if err != nil {
		t.Fatalf("EncryptPrivateKey failed: %v", err)
	}

	if encryptedKey == "" {
		t.Error("EncryptPrivateKey returned empty encrypted key")
	}

	if salt == "" {
		t.Error("EncryptPrivateKey returned empty salt")
	}

	// Encrypted key should be different from original
	if encryptedKey == privateKey {
		t.Error("Encrypted key should be different from original")
	}

	// Decrypt private key
	decryptedKey, err := DecryptPrivateKey(encryptedKey, password, salt)
	if err != nil {
		t.Fatalf("DecryptPrivateKey failed: %v", err)
	}

	if decryptedKey != privateKey {
		t.Error("Decrypted key should match original private key")
	}
}

func TestDecryptPrivateKey_WrongPassword(t *testing.T) {
	_, privateKey, _ := GenerateRSAKeyPair()
	password := "correctPassword"
	wrongPassword := "wrongPassword"

	encryptedKey, salt, _ := EncryptPrivateKey(privateKey, password)

	// Decrypting with wrong password should fail
	_, err := DecryptPrivateKey(encryptedKey, wrongPassword, salt)
	if err == nil {
		t.Error("DecryptPrivateKey should fail with wrong password")
	}
}

func TestGenerateSymmetricKey(t *testing.T) {
	key1, err := GenerateSymmetricKey()
	if err != nil {
		t.Fatalf("GenerateSymmetricKey failed: %v", err)
	}

	if key1 == "" {
		t.Error("GenerateSymmetricKey returned empty key")
	}

	// Generate another key and verify uniqueness
	key2, _ := GenerateSymmetricKey()
	if key1 == key2 {
		t.Error("Two generated symmetric keys should be different")
	}
}

func TestEndToEndEncryption(t *testing.T) {
	// Simulate the full encryption flow

	// 1. User A generates key pair
	pubA, privA, _ := GenerateRSAKeyPair()
	password := "userAPassword"

	// 2. Encrypt private key with password (for storage)
	encryptedPrivA, saltA, _ := EncryptPrivateKey(privA, password)

	// 3. User B generates key pair
	pubB, privB, _ := GenerateRSAKeyPair()

	// 4. Generate symmetric key for chat
	chatKey, _ := GenerateSymmetricKey()

	// 5. Encrypt chat key for both users
	encryptedKeyForA, _ := RSAEncrypt(pubA, []byte(chatKey))
	encryptedKeyForB, _ := RSAEncrypt(pubB, []byte(chatKey))

	// 6. User A retrieves and decrypts their private key (simulating new device login)
	decryptedPrivA, _ := DecryptPrivateKey(encryptedPrivA, password, saltA)

	// 7. User A decrypts chat key
	decryptedChatKeyA, _ := RSADecrypt(decryptedPrivA, encryptedKeyForA)

	// 8. User B decrypts chat key
	decryptedChatKeyB, _ := RSADecrypt(privB, encryptedKeyForB)

	// 9. Both should have the same chat key
	if string(decryptedChatKeyA) != chatKey {
		t.Error("User A's decrypted chat key doesn't match original")
	}

	if string(decryptedChatKeyB) != chatKey {
		t.Error("User B's decrypted chat key doesn't match original")
	}

	if string(decryptedChatKeyA) != string(decryptedChatKeyB) {
		t.Error("Both users should have the same chat key")
	}
}

func TestAESEncryptDecrypt_WithDerivedKey(t *testing.T) {
	password := "testPassword"
	salt, _ := GenerateSalt()
	key, _ := DeriveKey(password, salt)

	message := []byte("This is a secret message")

	// Encrypt
	ciphertext, err := AESEncrypt(message, key)
	if err != nil {
		t.Fatalf("AESEncrypt failed: %v", err)
	}

	// Decrypt
	decrypted, err := AESDecrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("AESDecrypt failed: %v", err)
	}

	if string(decrypted) != string(message) {
		t.Errorf("Decrypted message mismatch: got %s, want %s", decrypted, message)
	}
}

// Benchmark tests
func BenchmarkGenerateRSAKeyPair(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateRSAKeyPair()
	}
}

func BenchmarkRSAEncrypt(b *testing.B) {
	pub, _, _ := GenerateRSAKeyPair()
	plaintext := []byte("benchmark test message")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RSAEncrypt(pub, plaintext)
	}
}

func BenchmarkRSADecrypt(b *testing.B) {
	pub, priv, _ := GenerateRSAKeyPair()
	plaintext := []byte("benchmark test message")
	ciphertext, _ := RSAEncrypt(pub, plaintext)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RSADecrypt(priv, ciphertext)
	}
}

func BenchmarkDeriveKey(b *testing.B) {
	password := "benchmarkPassword"
	salt, _ := GenerateSalt()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DeriveKey(password, salt)
	}
}
