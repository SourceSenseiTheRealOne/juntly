package contactreveal

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

var ErrInvalidEncryptionKey = errors.New("invalid contact encryption key")

type SealedContact struct {
	Ciphertext []byte
	Nonce      []byte
}

type Cipher interface {
	Encrypt([]byte) (SealedContact, error)
	Decrypt(SealedContact) ([]byte, error)
}

type gcmCipher struct{ gcm cipher.AEAD }

func NewCipher(encodedKey string) (Cipher, error) {
	key, err := base64.StdEncoding.DecodeString(encodedKey)
	if err != nil || len(key) != 32 {
		return nil, ErrInvalidEncryptionKey
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, ErrInvalidEncryptionKey
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, ErrInvalidEncryptionKey
	}
	return gcmCipher{gcm: gcm}, nil
}

func (c gcmCipher) Encrypt(plaintext []byte) (SealedContact, error) {
	nonce := make([]byte, c.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return SealedContact{}, err
	}
	return SealedContact{Ciphertext: c.gcm.Seal(nil, nonce, plaintext, nil), Nonce: nonce}, nil
}

func (c gcmCipher) Decrypt(sealed SealedContact) ([]byte, error) {
	if len(sealed.Nonce) != c.gcm.NonceSize() {
		return nil, errors.New("invalid contact nonce")
	}
	return c.gcm.Open(nil, sealed.Nonce, sealed.Ciphertext, nil)
}
