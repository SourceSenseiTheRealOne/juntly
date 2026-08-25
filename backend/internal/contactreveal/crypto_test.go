package contactreveal

import (
	"encoding/base64"
	"errors"
	"testing"
)

func TestCipherRequiresExactServerKeyAndRoundTripsWithoutPlaintextMetadata(t *testing.T) {
	t.Parallel()
	key := make([]byte, 32)
	for index := range key {
		key[index] = byte(index + 1)
	}
	cipher, err := NewCipher(base64.StdEncoding.EncodeToString(key))
	if err != nil {
		t.Fatalf("new cipher: %v", err)
	}
	sealed, err := cipher.Encrypt([]byte("+351912345678"))
	if err != nil || len(sealed.Ciphertext) == 0 || len(sealed.Nonce) == 0 {
		t.Fatalf("sealed/error = %#v/%v", sealed, err)
	}
	opened, err := cipher.Decrypt(sealed)
	if err != nil || string(opened) != "+351912345678" {
		t.Fatalf("opened/error = %q/%v", opened, err)
	}
}

func TestCipherRejectsMalformedOrWrongLengthServerKeys(t *testing.T) {
	t.Parallel()
	for _, key := range []string{"", "not-base64", base64.StdEncoding.EncodeToString(make([]byte, 31))} {
		if _, err := NewCipher(key); !errors.Is(err, ErrInvalidEncryptionKey) {
			t.Fatalf("key %q error = %v", key, err)
		}
	}
}
