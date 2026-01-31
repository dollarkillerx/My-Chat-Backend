package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"

	"golang.org/x/crypto/pbkdf2"
)

const (
	// RSAKeySize RSA密钥长度
	RSAKeySize = 2048
	// PBKDF2Iterations PBKDF2迭代次数
	PBKDF2Iterations = 100000
	// SaltSize 盐值长度
	SaltSize = 16
	// DerivedKeySize 派生密钥长度
	DerivedKeySize = 32
)

// GenerateRSAKeyPair 生成RSA密钥对
// 返回: 公钥(PEM格式Base64), 私钥(PEM格式Base64), error
func GenerateRSAKeyPair() (string, string, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, RSAKeySize)
	if err != nil {
		return "", "", err
	}

	// 编码私钥为PEM
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	// 编码公钥为PEM
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", err
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return base64.StdEncoding.EncodeToString(publicKeyPEM),
		base64.StdEncoding.EncodeToString(privateKeyPEM),
		nil
}

// RSAEncrypt 使用公钥加密数据
// publicKeyB64: Base64编码的PEM格式公钥
// plaintext: 待加密数据
func RSAEncrypt(publicKeyB64 string, plaintext []byte) (string, error) {
	publicKeyPEM, err := base64.StdEncoding.DecodeString(publicKeyB64)
	if err != nil {
		return "", err
	}

	block, _ := pem.Decode(publicKeyPEM)
	if block == nil {
		return "", errors.New("failed to decode public key PEM")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", err
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return "", errors.New("not an RSA public key")
	}

	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, rsaPub, plaintext, nil)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// RSADecrypt 使用私钥解密数据
// privateKeyB64: Base64编码的PEM格式私钥
// ciphertextB64: Base64编码的密文
func RSADecrypt(privateKeyB64 string, ciphertextB64 string) ([]byte, error) {
	privateKeyPEM, err := base64.StdEncoding.DecodeString(privateKeyB64)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		return nil, errors.New("failed to decode private key PEM")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return nil, err
	}

	return rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, ciphertext, nil)
}

// GenerateSalt 生成随机盐值
func GenerateSalt() (string, error) {
	salt := make([]byte, SaltSize)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(salt), nil
}

// DeriveKey 使用PBKDF2从密码派生密钥
// password: 用户密码
// saltB64: Base64编码的盐值
func DeriveKey(password string, saltB64 string) ([]byte, error) {
	salt, err := base64.StdEncoding.DecodeString(saltB64)
	if err != nil {
		return nil, err
	}
	return pbkdf2.Key([]byte(password), salt, PBKDF2Iterations, DerivedKeySize, sha256.New), nil
}

// EncryptPrivateKey 使用密码加密私钥
// privateKeyB64: Base64编码的私钥
// password: 用户密码
// 返回: 加密后的私钥(Base64), 盐值(Base64), error
func EncryptPrivateKey(privateKeyB64 string, password string) (string, string, error) {
	salt, err := GenerateSalt()
	if err != nil {
		return "", "", err
	}

	derivedKey, err := DeriveKey(password, salt)
	if err != nil {
		return "", "", err
	}

	privateKeyBytes, err := base64.StdEncoding.DecodeString(privateKeyB64)
	if err != nil {
		return "", "", err
	}

	encryptedKey, err := AESEncrypt(privateKeyBytes, derivedKey)
	if err != nil {
		return "", "", err
	}

	return encryptedKey, salt, nil
}

// DecryptPrivateKey 使用密码解密私钥
// encryptedKeyB64: 加密后的私钥(Base64)
// password: 用户密码
// saltB64: Base64编码的盐值
func DecryptPrivateKey(encryptedKeyB64 string, password string, saltB64 string) (string, error) {
	derivedKey, err := DeriveKey(password, saltB64)
	if err != nil {
		return "", err
	}

	privateKeyBytes, err := AESDecrypt(encryptedKeyB64, derivedKey)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(privateKeyBytes), nil
}

// GenerateSymmetricKey 生成对称密钥 (用于消息加密)
func GenerateSymmetricKey() (string, error) {
	key, err := GenerateAESKey(32) // AES-256
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}
