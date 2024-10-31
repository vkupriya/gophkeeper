package helpers

import (
	"fmt"

	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
)

func generateRandom(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, fmt.Errorf("error generating random bytes: %w", err)
	}
	return b, nil
}

func Encrypt(key string, b *[]byte) ([]byte, error) {
	k := sha256.Sum256([]byte(key))
	kb := k[:]

	aesblock, err := aes.NewCipher(kb)
	if err != nil {
		return nil, fmt.Errorf("error creating new block cipher: %w", err)
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return nil, fmt.Errorf("error creating new GCM for block cipher: %w", err)
	}

	nonce, err := generateRandom(aesgcm.NonceSize())
	if err != nil {
		return nil, fmt.Errorf("error generating random nonce: %w", err)
	}

	bytesEncrypted := aesgcm.Seal(nonce, nonce, *b, nil)
	return bytesEncrypted, nil
}

func Decrypt(key string, b []byte) (*[]byte, error) {
	k := sha256.Sum256([]byte(key))
	kb := k[:]

	aesblock, err := aes.NewCipher(kb)
	if err != nil {
		return nil, fmt.Errorf("error creating new block cipher: %w", err)
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return nil, fmt.Errorf("error creating new GCM for block cipher: %w", err)
	}

	nonce, bytesEncrypted := b[:aesgcm.NonceSize()], b[aesgcm.NonceSize():]

	bytesDecrypted, err := aesgcm.Open(nil, nonce, bytesEncrypted, nil)
	if err != nil {
		return nil, fmt.Errorf("error decrypting data: %w", err)
	}

	return &bytesDecrypted, nil
}
