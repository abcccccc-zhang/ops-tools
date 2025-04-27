package handler

import (
	"alter-api/config"
	// rl "alter-api/config" // 确保你引入了 config 包
	"fmt"
	"log"
	"net/http"
)

// Reloadcfg 处理器用于处理 /reload 路由
func Reloadcfg(w http.ResponseWriter, r *http.Request) {
	fmt.Println("-----------------------------------------")
	// 调用 ReloadConfig 来重新加载配置
	err := config.ReloadConfig() // 假设 ReloadConfig 返回错误
	if err != nil {
		// 如果发生错误，返回500服务器错误
		http.Error(w, "Failed to reload config", http.StatusInternalServerError)
		log.Printf("Error reloading config: %v", err)
		return
	}
	// 配置重新加载成功后，记录日志
	log.Println("Config reloaded successfully")

	// 配置重新加载成功后，记录日志
	log.Println("Config reloaded successfully and log configuration updated.")
	// 返回一个简单的响应给客户端
	w.Write([]byte("Config reloaded successfully"))
}
