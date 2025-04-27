package encryption

import (
	"bytes"
	"crypto/cipher"
	"crypto/des"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
)

// 生成密钥和 IV
func getDerivedKey(password string, salt []byte, count int) ([]byte, []byte) {
	key := md5.Sum([]byte(password + string(salt)))
	for i := 0; i < count-1; i++ {
		key = md5.Sum(key[:])
	}
	return key[:8], key[8:]
}

// PKCS5 填充
func pKCS5Padding(cipherText []byte, blockSize int) []byte {
	padding := blockSize - len(cipherText)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(cipherText, padText...)
}

// PKCS5 反填充
func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unPadding := int(origData[length-1])
	return origData[:(length - unPadding)]
}

// DES 加密
func DesEncrypt(origData, key, iv []byte) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	origData = pKCS5Padding(origData, block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, iv)
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

// DES 解密
func DesDecrypt(crypted, key, iv []byte) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockMode := cipher.NewCBCDecrypter(block, iv)
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = PKCS5UnPadding(origData)
	return origData, nil
}

// 解密
func Decrypt(msg, password string) (string, error) {
	msgBytes, err := base64.StdEncoding.DecodeString(msg)
	if err != nil {
		return "", err
	}
	salt := msgBytes[:8]
	encText := msgBytes[8:]

	dk, iv := getDerivedKey(password, salt, 1000)

	text, err := DesDecrypt(encText, dk, iv)
	if err != nil {
		return "Decrypt错误", err
	}
	return string(text), nil
}

// 加密
func Encrypt(msg, password string) (string, error) {
	salt := make([]byte, 8)
	_, err := rand.Read(salt)
	if err != nil {
		return "Encrypt错误", err
	}

	dk, iv := getDerivedKey(password, salt, 1000)
	encText, err := DesEncrypt([]byte(msg), dk, iv)
	if err != nil {
		return "", err
	}
	r := append(salt, encText...)
	encodeString := base64.StdEncoding.EncodeToString(r)
	return encodeString, nil
}
