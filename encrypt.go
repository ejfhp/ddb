package ddb

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

// Key should be 32 bytes (AES-256).
func AESEncrypt(key [32]byte, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("cannot initialize key: %w", err)
	}
	// Never use more than 2^32 random nonces with a given key because of the risk of a repeat.
	nonce := make([]byte, noncesize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("cannot generate nonce: %w", err)
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("cannot initialize encryption: %w", err)
	}
	crypted := make([]byte, 0)
	crypted = append(crypted, nonce...)
	cipherdata := aesgcm.Seal(nil, nonce, data, nil)
	crypted = append(crypted, cipherdata...)
	return crypted, nil
}

// Key should be 32 bytes (AES-256).
func AESDecrypt(key [32]byte, encrypted []byte) ([]byte, error) {
	nonce := encrypted[:noncesize]
	cipherdata := encrypted[noncesize:]
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("cannot initialize key: %w", err)
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("cannot initialize encryption: %w", err)
	}
	plaindata, err := aesgcm.Open(nil, nonce, cipherdata, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot decrypt: %w", err)
	}
	return plaindata, err
}
