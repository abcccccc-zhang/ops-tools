package encryption

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// 生成密钥和 IV
func deriveKeyAndIV(password, salt []byte, keyLen, ivLen int) ([]byte, []byte, error) {
	var d, dI []byte
	h := md5.New()
	for len(d) < keyLen+ivLen {
		h.Write(dI)
		h.Write(password)
		h.Write(salt)
		dI = h.Sum(nil)
		h.Reset()
		d = append(d, dI...)
	}
	return d[:keyLen], d[keyLen : keyLen+ivLen], nil
}

// PKCS7 填充
func pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// PKCS7 反填充
func pkcs7UnPadding(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("invalid data")
	}
	padding := int(data[length-1])
	if padding > length {
		return nil, errors.New("invalid padding size")
	}
	return data[:(length - padding)], nil
}

// AES 加密（CBC 模式）
func aesEncrypt(data, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	data = pkcs7Padding(data, aes.BlockSize)
	cipherText := make([]byte, len(data))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherText, data)
	return cipherText, nil
}

// AES 解密（CBC 模式）
func aesDecrypt(data, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(data)%aes.BlockSize != 0 {
		return nil, errors.New("invalid ciphertext length")
	}
	plainText := make([]byte, len(data))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plainText, data)
	return pkcs7UnPadding(plainText)
}

// Encrypt 加密函数
func Encrypt_aes(plainText, password string) (string, error) {
	// 生成8字节的随机盐
	salt := make([]byte, 8)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", err
	}

	// 从密码和盐值中派生密钥和 IV
	key, iv, err := deriveKeyAndIV([]byte(password), salt, 32, aes.BlockSize)
	if err != nil {
		return "", err
	}

	// 加密明文
	cipherText, err := aesEncrypt([]byte(plainText), key, iv)
	if err != nil {
		return "", err
	}

	// 输出格式: "Salted__" + salt + cipherText
	finalData := append([]byte("Salted__"), salt...)
	finalData = append(finalData, cipherText...)

	return base64.StdEncoding.EncodeToString(finalData), nil
}

// Decrypt 解密函数
func Decrypt_aes(encryptedText, password string) (string, error) {
	// Base64 解码
	data, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", err
	}

	// 检查前缀 "Salted__"
	if len(data) < 16 || string(data[:8]) != "Salted__" {
		return "", errors.New("invalid encrypted text")
	}

	// 提取盐值
	salt := data[8:16]
	cipherText := data[16:]

	// 从密码和盐值中派生密钥和 IV
	key, iv, err := deriveKeyAndIV([]byte(password), salt, 32, aes.BlockSize)
	if err != nil {
		return "", err
	}

	// 解密
	plainText, err := aesDecrypt(cipherText, key, iv)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}
