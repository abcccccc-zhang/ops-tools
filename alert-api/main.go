package main

import (
	"alter-api/handler"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/health", handler.HealthHandler)
	r.POST("/alert", handler.AlertHandler)
	r.GET("/reload", handler.Reloadcfg)

	log.Println("Starting server on :5000  access /health /reload /alert v1")
	r.Run(":5000")
}
