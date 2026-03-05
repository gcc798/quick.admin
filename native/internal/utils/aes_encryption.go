package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
)

const (
	aesAPIKey = "ab3de6fg"
	aesSecret = "f6fec7aaka0a83tt8997gac486tt5a8fdcc19923"
)

func EncryptWithTimestamp(data interface{}, timestamp string) (string, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	key := generateKey(aesAPIKey, aesSecret, timestamp)
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}
	plaintext := pkcs7Pad(b, aes.BlockSize)
	ciphertext := make([]byte, len(plaintext))
	mode := cipher.NewCBCEncrypter(block, []byte(key))
	mode.CryptBlocks(ciphertext, plaintext)
	return hex.EncodeToString(ciphertext), nil
}

func generateKey(apiKey, secret, timestamp string) string {
	h := md5.Sum([]byte(apiKey + secret + timestamp))
	return hex.EncodeToString(h[:])[8:24]
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padLen := blockSize - len(data)%blockSize
	padding := bytes.Repeat([]byte{byte(padLen)}, padLen)
	return append(data, padding...)
}
