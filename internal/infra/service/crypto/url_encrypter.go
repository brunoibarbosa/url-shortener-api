package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
)

type URLEncrypter struct {
	secretKey string
}

func NewURLEncrypter(secretKey string) *URLEncrypter {
	return &URLEncrypter{
		secretKey,
	}
}

func (e *URLEncrypter) Encrypt(text string) (string, error) {
	block, err := aes.NewCipher([]byte(e.secretKey))
	if err != nil {
		return "", err
	}

	plainText := []byte(text)
	cipherText := make([]byte, aes.BlockSize+len(plainText))

	iv := cipherText[:aes.BlockSize]

	if _, err := rand.Read(iv); err != nil {
		return "", err
	}

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], plainText)

	return hex.EncodeToString(cipherText), nil
}

func (e *URLEncrypter) Decrypt(text string) (string, error) {
	block, err := aes.NewCipher([]byte(e.secretKey))
	if err != nil {
		return "", err
	}

	cipherText, err := hex.DecodeString(text)
	if err != nil {
		return "", err
	}

	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(cipherText, cipherText)

	return string(cipherText), nil
}
