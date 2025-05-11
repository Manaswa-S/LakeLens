package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/pbkdf2"
)

func deriveKeySha256(password string) ([]byte, error) {
	keySha := sha256.Sum256([]byte(password))
	return keySha[:], nil
}

func deriveKeyPBKDF2(password string, salt string) ([]byte, error) {
	key := pbkdf2.Key([]byte(password), []byte(salt), 100000, 32, sha256.New)
	return key, nil
}


// TODO: implement the encryption/decryption 
func EncryptStringAESGSM(plaintext string) (string, error) {

	password, exists := os.LookupEnv("ClientDataEncryptionPassword")
	if !exists || len(password) < 16 {
		return "", errors.New("client data encryption password does not exist or is not valid")
	}

	key, err := deriveKeySha256(password)
	if err != nil {
		return "", fmt.Errorf("error generating encryption key : %v", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("error creating new cipher block : %v", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("error wrapping cipher block in GSM : %v", err)
	}

	nonce := make([]byte, aesGCM.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return "", fmt.Errorf("error generating nonce : %v", err)
	}

	cipherText := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)

	plaintext = base64.StdEncoding.EncodeToString(cipherText)

	return plaintext, nil
}

func DecryptStringAESGSM(cipherText string) (string, error) {

	password, exists := os.LookupEnv("ClientDataEncryptionPassword")
	if !exists || len(password) < 16 {
		return "", errors.New("client data encryption password does not exist or is not valid")
	}

	key, err := deriveKeySha256(password)
	if err != nil {
		return "", fmt.Errorf("error generating encryption key : %v", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("error creating new cipher block : %v", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("error wrapping cipher block in GSM : %v", err)
	}

	cipherByte, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", fmt.Errorf("failed to decode cipher string to byte : %v", err)
	}

	nonceSize := aesGCM.NonceSize()
	nonce, cipherByte := cipherByte[:nonceSize], cipherByte[nonceSize:]

	plainByte, err := aesGCM.Open(nil, nonce, cipherByte, nil)
	if err != nil {
		return "", fmt.Errorf("failed to open cipher block : %v", err)
	}

	return string(plainByte), nil
}