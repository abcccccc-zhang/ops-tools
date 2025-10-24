package web

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func StartServer() {
	r := gin.Default()

	r.POST("/api/getpackagelist", HandleListPackage)
	r.POST("/api/generateDownloadURL", HandleGenerateDownloadURL)
	fmt.Println("Server is running on :8080")
	if err := r.Run(":8080"); err != nil {
		fmt.Printf("Failed to start server: %s\n", err)
	}
}
