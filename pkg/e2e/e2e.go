package e2e

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
)

type Service struct {
	secretKey []byte
}

func New(secretKey string) Service {
	return Service{
		secretKey: []byte(secretKey),
	}
}

func (s Service) Encrypt(content []byte) ([]byte, error) {
	aes, err := aes.NewCipher(s.secretKey)
	if err != nil {
		return nil, fmt.Errorf("cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(aes)
	if err != nil {
		return nil, fmt.Errorf("gcm: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("nonce: %w", err)
	}

	return gcm.Seal(nonce, nonce, content, nil), nil
}

func (s Service) Decrypt(content []byte) ([]byte, error) {
	aes, err := aes.NewCipher(s.secretKey)
	if err != nil {
		return nil, fmt.Errorf("cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(aes)
	if err != nil {
		return nil, fmt.Errorf("gcm: %w", err)
	}

	nonceSize := gcm.NonceSize()

	return gcm.Open(nil, content[:nonceSize], content[nonceSize:], nil)
}
