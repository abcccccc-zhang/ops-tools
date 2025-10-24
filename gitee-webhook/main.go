package main

import (
	"encoding/json"
	"github.com/Knetic/govaluate"
	"gopkg.in/yaml.v2"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"

	"gopkg.in/natefinch/lumberjack.v2"
)

type Config struct {
	Password  string `yaml:"password"`
	Condition string `yaml:"condition"` // 条件表达式
	Action    struct {
		ShellScript string `yaml:"shell-script"`
	} `yaml:"action"`
	Log struct {
		Filename   string `yaml:"filename"`
		MaxSize    int    `yaml:"max_size"`
		MaxBackups int    `yaml:"max_backups"`
		MaxAge     int    `yaml:"max_age"`
		Compress   bool   `yaml:"compress"`
	} `yaml:"log"`
}

type GiteeWebhookPayload struct {
	Action   string `json:"action"`
	Password string `json:"password"`
	Ref      string `json:"ref"` // 顶层 ref 字段

	HookName string `json:"hook_name"` // hook 名称
	//PushData struct {
	//	Ref string `json:"ref"`
	//} `json:"push_data"`
	Issue struct {
		Title string `json:"title"`
	} `json:"issue"`
	Comment struct {
		Body string `json:"body"`
	} `json:"comment"`
}

var config Config

func loadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, &config)
}

func initLogger() {
	_ = os.MkdirAll("logs", 0755)
	log.SetOutput(io.MultiWriter(os.Stdout, &lumberjack.Logger{
		Filename:   config.Log.Filename,
		MaxSize:    config.Log.MaxSize,
		MaxBackups: config.Log.MaxBackups,
		MaxAge:     config.Log.MaxAge,
		Compress:   config.Log.Compress,
	}))
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Read body error", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	log.Println("Received WebHook payload:")
	log.Println(string(body))

	var payload GiteeWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "JSON unmarshal error", http.StatusBadRequest)
		log.Println("Unmarshal error:", err)
		return
	}

	// 密码校验
	if payload.Password != config.Password {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		log.Println("Password mismatch")
		return
	}

	// 判断是否是 push_hooks 且分支是
	branch := payload.Ref
	log.Printf("Hook: %s, Ref: %s\n", payload.HookName, branch)
	parameters := map[string]interface{}{
		"hookname": payload.HookName,
		"branch":   branch,
	}
	// 表达式解析
	expr, err := govaluate.NewEvaluableExpression(config.Condition)
	if err != nil {
		log.Printf("Condition expression error: %v", err)
		http.Error(w, "Condition expression error", http.StatusInternalServerError)
		return
	}

	result, err := expr.Evaluate(parameters)
	if err != nil {
		log.Printf("Condition evaluation error: %v", err)
		http.Error(w, "Condition evaluation error", http.StatusInternalServerError)
		return
	}

	trigger := false
	if boolResult, ok := result.(bool); ok {
		trigger = boolResult
	}

	if trigger {
		log.Println("Condition matched, executing script")
		//cmd := exec.Command("sh", config.Action.ShellScript)
		//cmd.Stdout = os.Stdout
		//cmd.Stderr = os.Stderr
		//if err := cmd.Run(); err != nil {
		//	log.Printf("Script execution failed: %v", err)
		//	http.Error(w, "Script execution failed", http.StatusInternalServerError)
		//	return
		//}
		cmd := exec.Command("sh", config.Action.ShellScript)

		stdoutPipe, _ := cmd.StdoutPipe()
		stderrPipe, _ := cmd.StderrPipe()

		if err := cmd.Start(); err != nil {
			log.Printf("Script start failed: %v", err)
			http.Error(w, "Script start failed", http.StatusInternalServerError)
			return
		}

		// 读取 stdout 和 stderr 并写入日志
		go func() {
			stdoutBytes, _ := io.ReadAll(stdoutPipe)
			if len(stdoutBytes) > 0 {
				log.Printf("Script stdout:\n%s", string(stdoutBytes))
			}
		}()

		go func() {
			stderrBytes, _ := io.ReadAll(stderrPipe)
			if len(stderrBytes) > 0 {
				log.Printf("Script stderr:\n%s", string(stderrBytes))
			}
		}()

		if err := cmd.Wait(); err != nil {
			log.Printf("Script execution failed: %v", err)
			http.Error(w, "Script execution failed", http.StatusInternalServerError)
			return
		}
	} else {
		log.Println("Condition not matched, skipping execution")
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Webhook received"))
}

func main() {
	if err := loadConfig("config.yaml"); err != nil {
		log.Fatalf("Failed to load config.yaml: %v", err)
	}

	initLogger()
	log.Println("Webhook Server Starting at :7878")

	http.HandleFunc("/webhook", webhookHandler)
	log.Fatal(http.ListenAndServe(":7878", nil))
}
