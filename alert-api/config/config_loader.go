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

func readYAMLFile(path string, out interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read file %s failed: %w", path, err)
	}
	return yaml.Unmarshal(data, out)
}

func LoadConfig() (models.Config, error) {
	//mu.Lock()
	//defer mu.Unlock()
	if cacheLoaded {
		return configCache, nil
	}
	// 只在需要初始化时上锁
	mu.Lock()
	defer mu.Unlock()

	// 双重检查，防止并发竞争加载
	if cacheLoaded {
		return configCache, nil
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

	if err := readYAMLFile("conditions.yml", &configCache.Conditions); err != nil {
		log.Printf("[WARN] %v", err)
	}

	if err := readYAMLFile("actions.yml", &configCache.Actions); err != nil {
		log.Printf("[WARN] %v", err)
	}

	if err := readYAMLFile("config.yml", &configCache.Configuration); err != nil {
		log.Printf("[WARN] %v", err)
	}

	cacheLoaded = true
	return configCache, nil
}

func ReloadConfig() error {
	mu.Lock()
	cacheLoaded = false
	configCache = models.Config{}
	mu.Unlock()
	log.Println("Start reloading config...")
	cfg, err := LoadConfig()
	if err != nil {
		log.Printf("Failed to reload config: %v", err)
	} else {
		log.Println("Finish reloading config...")
	}
	InitLogger(cfg.LogConfig)
	return err
}
