package handler

import (
	Action "alter-api/action"
	"alter-api/config"
	"alter-api/models"
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	Template "text/template"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/alertmanager/template"
)

func HealthHandler(c *gin.Context) {
	c.String(http.StatusOK, "OK")
}
func GenSign(secret string, timestamp int64) (string, error) {
	//timestamp + key 做sha256, 再进行base64 encode
	stringToSign := fmt.Sprintf("%v", timestamp) + "\n" + secret
	var data []byte
	h := hmac.New(sha256.New, []byte(stringToSign))
	_, err := h.Write(data)
	if err != nil {
		return "", err
	}
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return signature, nil
}
func AlertHandler(c *gin.Context) {
	body, err := c.GetRawData()
	if err != nil {
		log.Printf("Failed to read body: %v", err)
		c.String(http.StatusBadRequest, "Invalid body")
		return
	}

	log.Printf("Request body: %s", string(body)) // 转成 string 打印，防止乱码或 byte 错误

	var data template.Data
	if err := json.Unmarshal(body, &data); err != nil {
		log.Printf("Failed to bind JSON: %v", err)
		c.String(http.StatusBadRequest, "Invalid JSON")
		return
	}
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Printf("load config fail: %v", err)
		c.String(http.StatusInternalServerError, fmt.Sprintf("load config fail: %v", err))
		return
	}

	var result string
	if len(cfg.Conditions) == 0 || len(cfg.Actions) == 0 {
		log.Printf("No conditions or actions defined, skipping execution.")
		result = "No conditions or actions configured"
	} else {
		// 传入告警的msg给action
		alert := data.Alerts[0]
		msg := re.ReplaceAllString(alert.Annotations["description"], " ")
		result, err = matchAndExecute(data, cfg.Conditions, cfg.Actions, msg)
		if err != nil {
			log.Printf("Failed to execute action: %v", err)
			c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to execute action: %v", err))
			return
		}
	}

	sendFeiShu(data, cfg.Configuration)

	log.Printf("Result: %s", result)
	c.String(http.StatusOK, result)
}

var (
	re               = regexp.MustCompile("[\\n\"]+")
	feishuTmpl       *Template.Template
	loadTemplateOnce sync.Once
	loadTemplateErr  error
)

// 加载并缓存模板（只加载一次）
func loadFeishuTemplate(path string) error {
	loadTemplateOnce.Do(func() {
		tmplBytes, err := os.ReadFile(path)
		if err != nil {
			loadTemplateErr = fmt.Errorf("failed to read template: %w", err)
			return
		}
		feishuTmpl, loadTemplateErr = Template.New("feishu").Parse(string(tmplBytes))
	})
	return loadTemplateErr
}

func sendFeiShu(data template.Data, configs map[string]models.Configuration) string {
	cfg, ok := configs["Feishu"]
	if !ok || cfg.Webhook == "" || cfg.Template.CustomPath == "" {
		log.Printf("Feishu webhook or template path not configured")
		return ""
	}

	if len(data.Alerts) == 0 {
		log.Println("No alerts found in data")
		return ""
	}

	timestamp := time.Now().Unix()
	sign, err := GenSign(cfg.Sign, timestamp)
	if err != nil {
		log.Printf("Failed to genSign: %v", err)
		return ""
	}

	alert := data.Alerts[0]
	msg := re.ReplaceAllString(alert.Annotations["description"], " ")

	tmplData := struct {
		Sign      string
		Timestamp int64
		Alertname string
		Msg       string
		URL       string
		Now       string
	}{
		Sign:      sign,
		Timestamp: timestamp,
		Alertname: alert.Labels["alertname"],
		Msg:       msg,
		URL:       alert.GeneratorURL,
		Now:       time.Now().Format("2006-01-02 15:04:05"),
	}

	if err := loadFeishuTemplate(cfg.Template.CustomPath); err != nil {
		log.Printf("Failed to load template: %v", err)
		return ""
	}

	var buf bytes.Buffer
	if err := feishuTmpl.Execute(&buf, tmplData); err != nil {
		log.Printf("Template execution failed: %v", err)
		return ""
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		log.Printf("Failed to unmarshal template output: %v", err)
		return ""
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal JSON: %v", err)
		return ""
	}

	resp, err := http.Post(cfg.Webhook, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Failed to send Feishu webhook: %v", err)
		return ""
	}
	defer resp.Body.Close()

	log.Printf("Feishu webhook response: %s", resp.Status)
	return tmplData.Alertname
}

func matchAndExecute(data template.Data, conditions []models.Condition, actions map[string]models.Action, msg string) (string, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var resultMessage string
	var firstError error
	// matched := false
	// for _, condition := range conditions {
	// 	if conditionMatches(data, condition) {
	// 		matched = true
	// 		action, found := actions[condition.Action]
	// 		if !found {
	// 			continue
	// 		}
	// 		wg.Add(1)
	// 		go func(action models.Action, condition models.Condition) {
	// 			defer wg.Done()
	// 			if err := Action.ExecuteAction(action, condition, msg); err != nil {
	// 				mu.Lock()
	// 				if firstError == nil {
	// 					firstError = err
	// 				}
	// 				mu.Unlock()
	// 				log.Printf("Action execution failed: %v", err)
	// 			} else {
	// 				mu.Lock()
	// 				if resultMessage == "" {
	// 					resultMessage = "Action executed successfully"
	// 				}
	// 				mu.Unlock()
	// 			}
	// 		}(action, condition)
	// 	}
	// }
	// if !matched {
	// 	resultMessage = "unmatched conditions"
	// }
	for _, condition := range conditions {
		if conditionMatches(data, condition) {
			action, found := actions[condition.Action]
			if !found {
				continue
			}
			wg.Add(1)
			go func(action models.Action, condition models.Condition) {
				defer wg.Done()
				if err := Action.ExecuteAction(action, condition, msg); err != nil {
					mu.Lock()
					if firstError == nil {
						firstError = err
					}
					mu.Unlock()
					log.Printf("Action execution failed: %v", err)
				} else {
					mu.Lock()
					if resultMessage == "" {
						resultMessage = "Action executed successfully"
					}
					mu.Unlock()
				}
			}(action, condition)
		}
		resultMessage = "unmatched conditions "
	}
	wg.Wait()

	if firstError != nil {
		return "Action executed failed", firstError
	}
	return resultMessage, nil
}

func conditionMatches(data template.Data, condition models.Condition) bool {
	for key, value := range condition.Fields {
		if !matchValue(key, value, data) {
			log.Printf("condition not match: %s=%s", key, value)
			return false
		}
		log.Printf("condition match: %s=%s", key, value)
	}
	return true
}

func matchValue(key, value string, data template.Data) bool {
	keyLower := strings.ToLower(key)
	switch keyLower {
	case "receiver":
		return value == data.Receiver
	case "status":
		return value == data.Status
	case "externalurl":
		return value == data.ExternalURL
	default:
		for _, alert := range data.Alerts {
			if v, ok := alert.Labels[key]; ok && v == value {
				return true
			}
			if v, ok := alert.Annotations[key]; ok && v == value {
				return true
			}
		}
		if v, ok := data.GroupLabels[key]; ok && v == value {
			return true
		}
		if v, ok := data.CommonLabels[key]; ok && v == value {
			return true
		}
		if v, ok := data.CommonAnnotations[key]; ok && v == value {
			return true
		}
	}
	return false
}

// func matchValue(key, value string, data template.Data) bool {
// 	keyLower := strings.ToLower(key)
// 	switch keyLower {
// 	case "receiver":
// 		return value == data.Receiver
// 	case "status":
// 		return value == data.Status
// 	case "externalurl":
// 		return value == data.ExternalURL
// 	default:
// 		for _, alert := range data.Alerts {
// 			for k, v := range alert.Labels {
// 				fmt.Printf("matchValue: %s=%s", k, v)
// 				if strings.ToLower(k) == keyLower && v == value {
// 					return true
// 				}
// 			}
// 			for k, v := range alert.Annotations {
// 				fmt.Printf("matchValue: %s=%s", k, v)
// 				if strings.ToLower(k) == keyLower && v == value {
// 					return true
// 				}
// 			}
// 		}
// 		for k, v := range data.GroupLabels {
// 			fmt.Printf("matchValue: %s=%s", k, v)
// 			if strings.ToLower(k) == keyLower && v == value {
// 				return true
// 			}
// 		}
// 		for k, v := range data.CommonLabels {
// 			if strings.ToLower(k) == keyLower && v == value {
// 				return true
// 			}
// 		}
// 		for k, v := range data.CommonAnnotations {
// 			if strings.ToLower(k) == keyLower && v == value {
// 				return true
// 			}
// 		}
// 	}
// 	return false
// }
