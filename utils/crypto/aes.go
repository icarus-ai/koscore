package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
)

var DEBUG = false

func AESGCMEncrypt(data []byte, key []byte) ([]byte, error) {
	nonce := make([]byte, 12)
	if !DEBUG {
		if _, e := rand.Read(nonce); e != nil {
			return nil, e
		}
	}

	block, e := aes.NewCipher(key)
	if e != nil {
		return nil, e
	}
	aead, e := cipher.NewGCM(block)
	if e != nil {
		return nil, e
	}
	data = aead.Seal(nil, nonce, data, nil)

	result := make([]byte, len(nonce)+len(data))
	copy(result[:len(nonce)], nonce)
	copy(result[len(nonce):], data)

	return result, nil
}

func AESGCMDecrypt(data []byte, key []byte) ([]byte, error) {
	block, e := aes.NewCipher(key)
	if e != nil {
		return nil, e
	}
	aead, e := cipher.NewGCM(block)
	if e != nil {
		return nil, e
	}
	// nonce, text
	return aead.Open(nil, data[:12], data[12:], nil)
}
