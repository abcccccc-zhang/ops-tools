package config

import (
	"alter-api/models"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/yaml.v2"
)

var configCache models.Config
var cacheLoaded bool
var mu sync.Mutex

func InitLogger(logConfig models.LogConfig) {
	// 读取 logs.yaml 配置
	// var logConfig models.LogConfig
	// config, err := config.LoadConfig() // 使用 LoadConfig 函数来加载配置
	// if err != nil {
	// 	// // 使用默认配置
	// 	// logConfig = models.LogConfig{
	// 	// 	Filename:   "logs/alert_handler.log",
	// 	// 	MaxSize:    10,
	// 	// 	MaxBackups: 5,
	// 	// 	MaxAge:     1,
	// 	// 	Compress:   true,
	// 	// }
	// 	log.Printf("Error reading logs.yaml: %v. Using default log configuration.", err)
	// } else {
	// 	logConfig = config.LogConfig // 从加载的配置中获取日志配置
	// }

	// 确保 logs 目录存在，如果不存在则创建
	if err := ensureLogsDirExists(); err != nil {
		log.Fatalf("Failed to create logs directory: %v", err)
	}

	// 初始化 lumberjack 日志记录器
	logFilePath := fmt.Sprintf(logConfig.Filename)
	lumberjackLogger := &lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    logConfig.MaxSize,
		MaxBackups: logConfig.MaxBackups,
		MaxAge:     logConfig.MaxAge,
		Compress:   logConfig.Compress,
	}

	// 设置日志输出到控制台和日志文件
	log.SetOutput(io.MultiWriter(os.Stdout, lumberjackLogger))
}

func ensureLogsDirExists() error {
	_, err := os.Stat("logs")
	if os.IsNotExist(err) {
		return os.Mkdir("logs", 0755) // 如果没有 logs 目录，创建它
	}
	return nil
}

func LoadConfig() (models.Config, error) {
	if cacheLoaded {
		return configCache, nil
	}

	conditionFile, err := os.ReadFile("conditions.yml")
	if err != nil {
		log.Printf("Error reading conditions file: %v", err)
		return configCache, err
	}
	err = yaml.Unmarshal(conditionFile, &configCache.Conditions)
	if err != nil {
		log.Printf("Error unmarshalling conditions file: %v", err)
		return configCache, err
	}

	actionFile, err := os.ReadFile("actions.yml")
	if err != nil {
		log.Printf("Error reading actions file: %v", err)
		return configCache, err
	}
	err = yaml.Unmarshal(actionFile, &configCache.Actions)
	if err != nil {
		log.Printf("Error unmarshalling actions file: %v", err)
		return configCache, err
	}

	logConfigFile, err := os.ReadFile("logs.yml")
	if err != nil {
		// 如果读取 logs.yaml 失败，使用默认的日志配置
		log.Printf("Error reading logs config file: %v. Using default log configuration.", err)
		configCache.LogConfig = models.LogConfig{
			Filename:   "logs/alert_handler.log",
			MaxSize:    10,
			MaxBackups: 7,
			MaxAge:     7,
			Compress:   true,
		}
		InitLogger(configCache.LogConfig)
	} else {
		err = yaml.Unmarshal(logConfigFile, &configCache.LogConfig)
		if err != nil {
			// 如果 unmarshalling 失败，也使用默认配置
			log.Printf("Error unmarshalling logs config file: %v. Using default log configuration.", err)
			configCache.LogConfig = models.LogConfig{
				Filename:   "logs/alert_handler.log",
				MaxSize:    10,
				MaxBackups: 7,
				MaxAge:     7,
				Compress:   true,
			}
		}
		InitLogger(configCache.LogConfig)
	}

	cacheLoaded = true
	return configCache, nil
}

func ReloadConfig() error {
	mu.Lock()
	defer mu.Unlock()
	// 清理掉旧的 configCache 数据
	configCache = models.Config{}
	cacheLoaded = false

	log.Println("Start reloading config...")
	cfg, err := LoadConfig()
	if err != nil {
		log.Printf("Failed to reload config: %v", err)
	} else {
		log.Println("Finish reloading config...")
	}
	InitLogger(cfg.LogConfig)
	// 配置重新加载成功后，初始化日志
	// cfg, err := LoadConfig()
	// if err != nil {
	// 	// 如果加载新配置失败，返回500服务器错误
	// 	// http.Error(w, "Failed to load new config", http.StatusInternalServerError)
	// 	log.Printf("Error loading new config: %v", err)
	// 	// return
	// }

	// // 使用新加载的日志配置来初始化日志
	// handler.InitLogger(cfg.LogConfig) // 这里调用 InitLogger 函数来更新日志配置
	return err
}
