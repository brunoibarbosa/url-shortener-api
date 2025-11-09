package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"log"
)

func EncryptURL(original string, secretKey string) string {
	block, err := aes.NewCipher([]byte(secretKey))
	if err != nil {
		log.Fatal(err)
	}

	plainText := []byte(original)
	cipherText := make([]byte, aes.BlockSize+len(plainText))

	iv := cipherText[:aes.BlockSize]

	if _, err := rand.Read(iv); err != nil {
		log.Fatal(err)
	}

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], plainText)

	return hex.EncodeToString(cipherText)
}

func DecryptURL(encrypted string, secretKey string) string {
	block, err := aes.NewCipher([]byte(secretKey))
	if err != nil {
		log.Fatal(err)
	}

	cipherText, err := hex.DecodeString(encrypted)
	if err != nil {
		log.Fatal(err)
	}

	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(cipherText, cipherText)

	return string(cipherText)
}
