package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
)

// AES256Encrypt encrypts plaintext using AES-256-CBC encryption.
// Returns format: {base64(iv)}:{base64(ciphertext)}
func AES256Encrypt(plaintext, key string) (string, error) {
	if len(key) != 32 {
		return "", errors.New("key must be 32 bytes for AES-256")
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Generate random IV
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", fmt.Errorf("failed to generate IV: %w", err)
	}

	// Apply PKCS7 padding
	plaintextBytes := []byte(plaintext)
	padding := aes.BlockSize - len(plaintextBytes)%aes.BlockSize
	padtext := make([]byte, len(plaintextBytes)+padding)
	copy(padtext, plaintextBytes)
	for i := len(plaintextBytes); i < len(padtext); i++ {
		padtext[i] = byte(padding)
	}

	// Encrypt
	ciphertext := make([]byte, len(padtext))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, padtext)

	// Return {iv}:{ciphertext} format
	ivBase64 := base64.StdEncoding.EncodeToString(iv)
	ciphertextBase64 := base64.StdEncoding.EncodeToString(ciphertext)

	return fmt.Sprintf("%s:%s", ivBase64, ciphertextBase64), nil
}

// AES256Decrypt decrypts ciphertext using AES-256-CBC encryption.
// Input format: {base64(iv)}:{base64(ciphertext)}
func AES256Decrypt(ciphertext, key string) (string, error) {
	if len(key) != 32 {
		return "", errors.New("key must be 32 bytes for AES-256")
	}

	// Parse {iv}:{ciphertext}
	parts := strings.SplitN(ciphertext, ":", 2)
	if len(parts) != 2 {
		return "", errors.New("invalid ciphertext format, expected {iv}:{ciphertext}")
	}

	ivBase64, ciphertextBase64 := parts[0], parts[1]

	// Base64 decode
	iv, err := base64.StdEncoding.DecodeString(ivBase64)
	if err != nil {
		return "", fmt.Errorf("failed to decode IV: %w", err)
	}

	ciphertextBytes, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	// Decrypt
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	if len(ciphertextBytes) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}

	if len(ciphertextBytes)%aes.BlockSize != 0 {
		return "", errors.New("ciphertext is not a multiple of block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertextBytes))
	mode.CryptBlocks(plaintext, ciphertextBytes)

	// Remove PKCS7 padding
	padding := int(plaintext[len(plaintext)-1])
	if padding < 1 || padding > aes.BlockSize {
		return "", errors.New("invalid padding")
	}

	plaintext = plaintext[:len(plaintext)-padding]

	return string(plaintext), nil
}
