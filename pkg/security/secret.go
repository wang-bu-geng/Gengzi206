package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"os"
	"strings"
)

const (
	// EnvKeyName 是用于配置加密主密钥的环境变量名。
	EnvKeyName = "DRAMA_KEY_SECRET"
	prefix     = "enc:v1:"
)

var (
	aesgcm  cipher.AEAD
	enabled bool
)

// Init 从环境变量读取主密钥并初始化 AES-256-GCM 加密器。
// 未设置 DRAMA_KEY_SECRET 时保持明文存储（向后兼容），返回 false。
func Init() bool {
	key := os.Getenv(EnvKeyName)
	if key == "" {
		enabled = false
		return false
	}

	// 用 SHA-256 将任意长度密钥派生为 32 字节
	sum := sha256.Sum256([]byte(key))
	block, err := aes.NewCipher(sum[:])
	if err != nil {
		enabled = false
		return false
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		enabled = false
		return false
	}
	aesgcm = gcm
	enabled = true
	return true
}

// Enabled 报告加密是否已启用。
func Enabled() bool { return enabled }

// Encrypt 加密明文，返回 "enc:v1:<base64(nonce+ciphertext)>"。
// 未启用加密、明文为空或加密失败时原样返回（避免锁死配置）。
func Encrypt(plain string) string {
	if !enabled || plain == "" {
		return plain
	}
	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return plain
	}
	sealed := aesgcm.Seal(nonce, nonce, []byte(plain), nil)
	return prefix + base64.StdEncoding.EncodeToString(sealed)
}

// Decrypt 解密密文。无 enc:v1: 前缀（遗留明文）或解密失败时原样返回。
func Decrypt(stored string) string {
	if !enabled || stored == "" || !strings.HasPrefix(stored, prefix) {
		return stored
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(stored, prefix))
	if err != nil {
		return stored
	}
	nonceSize := aesgcm.NonceSize()
	if len(raw) < nonceSize {
		return stored
	}
	nonce, ciphertext := raw[:nonceSize], raw[nonceSize:]
	plain, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return stored
	}
	return string(plain)
}
