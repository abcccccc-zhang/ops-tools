package main

import (
	encryption "encrtyption/en-de-crypt"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

type EncryptRequest struct {
	Msg           string `json:"msg" binding:"required"`
	EncryptionKey string `json:"encryption_key" binding:"required"`
	Algorithm     string `json:"algorithm" binding:"required"` // 修改字段名
}

type DecryptRequest struct {
	EncryptedString string `json:"encrypted_string" binding:"required"`
	EncryptionKey   string `json:"encryption_key" binding:"required"`
	Algorithm       string `json:"algorithm" binding:"required"` // 修改字段名
}

func main() {
	r := gin.Default()
	// r.Use(cors.Default())

	// 使用 POST 方法定义加密和解密路由
	r.POST("/api/encrypt", handleEncryption)
	r.POST("/api/decrypt", handleDecryption)

	fmt.Println("Server is running on :8080")
	if err := r.Run(":8080"); err != nil {
		fmt.Printf("Failed to start server: %s\n", err)
	}
}

// handleEncryption 处理加密请求
func handleEncryption(c *gin.Context) {
	var req EncryptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{Error: "请求参数无效"})
		return
	}

	var eMsg string
	var err error
	log.Printf("[INFO] Algorithm: %s, EncryptionKey: %s, Msg: %s ---------", req.Algorithm, req.EncryptionKey, req.Msg)

	if req.Algorithm == "PBEWithMD5AndTripleDES" {
		eMsg, err = encryption.Encrypt(req.Msg, req.EncryptionKey) // 使用算法 A
	} else {
		eMsg, err = encryption.Encrypt_aes(req.Msg, req.EncryptionKey) // 使用算法 B
	}
	for {
		if req.Algorithm == "PBEWithMD5AndTripleDES" {
			eMsg, err = encryption.Encrypt(req.Msg, req.EncryptionKey) // 使用算法 A
			if err != nil {
				c.JSON(http.StatusInternalServerError, Response{Error: fmt.Sprintf("加密失败: %v", err)})
				return
			}
			if !strings.Contains(eMsg, "/") {
				break
			}
		} else {
			eMsg, err = encryption.Encrypt_aes(req.Msg, req.EncryptionKey) // 使用算法 B
			if err != nil {
				c.JSON(http.StatusInternalServerError, Response{Error: fmt.Sprintf("加密失败: %v", err)})
				return
			}
			break
		}
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Error: fmt.Sprintf("加密失败: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		// "message":          req.Msg,
		// "encryption_key":   req.EncryptionKey,
		"encrypted_string": eMsg,
	})
}

// handleDecryption 处理解密请求
func handleDecryption(c *gin.Context) {
	var req DecryptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{Error: "请求参数无效"})
		return
	}
	var dMsg string
	var err error
	log.Printf("[INFO] Algorithm: %s, EncryptedString: %s, EncryptionKey: %s ---------", req.Algorithm, req.EncryptedString, req.EncryptionKey)

	if req.Algorithm == "PBEWithMD5AndTripleDES" {
		dMsg, err = encryption.Decrypt(req.EncryptedString, req.EncryptionKey) // 使用算法
	} else {
		dMsg, err = encryption.Decrypt_aes(req.EncryptedString, req.EncryptionKey) // 使用算法
	}
	// dMsg, err := encryption.Decrypt(req.EncryptedString, req.EncryptionKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Error: fmt.Sprintf("解密失败: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		// "encrypted_string": req.EncryptedString,
		// "encryption_key":   req.EncryptionKey,
		"decrypted_msg": dMsg,
	})
}
