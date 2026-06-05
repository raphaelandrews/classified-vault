package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

var vaultKey []byte

func InitVault(password string, salt []byte) {
	if password == "" {
		password = "classified-vault-default-key-change-me"
	}
	if len(salt) == 0 {
		salt = []byte("pelican-town-salt-v1")
	}
	vaultKey = pbkdf2.Key([]byte(password), salt, 100000, 32, sha256.New)
}

func Encrypt(plain string) (string, error) {
	if vaultKey == nil {
		return "", fmt.Errorf("vault not initialized")
	}

	block, err := aes.NewCipher(vaultKey)
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create gcm: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plain), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func Decrypt(encoded string) (string, error) {
	if vaultKey == nil {
		return "", fmt.Errorf("vault not initialized")
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("decode base64: %w", err)
	}

	block, err := aes.NewCipher(vaultKey)
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create gcm: %w", err)
	}

	if len(ciphertext) < gcm.NonceSize() {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce := ciphertext[:gcm.NonceSize()]
	ciphertext = ciphertext[gcm.NonceSize():]

	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}

	return string(plain), nil
}

func IsEncrypted(encoded string) bool {
	_, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return false
	}
	return len(encoded) > 50
}
