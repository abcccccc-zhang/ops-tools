package main

import (
	"alter-api/handler"
	"log"
	"net/http"
	_ "net/http/pprof"
)

func main() {
	// config, err := config.LoadConfig()
	// if err != nil {
	// 	log.Fatalf("Failed to load initial config: %v", err)
	// }
	// handler.InitLogger(config.LogConfig)
	// go func() {
	// 	log.Println(http.ListenAndServe(":6060", nil))
	// }()
	// 其他初始化代码...
	log.Println("Starting server on :5000")
	http.HandleFunc("/alert", handler.AlertHandler)
	http.HandleFunc("/reload", handler.Reloadcfg)
	log.Fatal(http.ListenAndServe(":5000", nil))
}
