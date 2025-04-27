package handler

import (
	Action "alter-api/action"
	"alter-api/config"
	"sync"

	// "alter-api/config"

	// "alter-api/config"
	"alter-api/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

// AlertHandler 处理 Alertmanager Webhook 数据
func AlertHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received a new alarm")
	defer r.Body.Close()

	// 限制请求体大小，防止读取过大的请求
	const maxRequestBodySize = 2 * 1024 // 2 KB
	var body []byte
	body, err := io.ReadAll(io.LimitReader(r.Body, maxRequestBodySize))
	if err != nil {
		http.Error(w, "read body fails", http.StatusBadRequest)
		return
	}

	log.Printf("Contents of request body (first 2 KB): %s", string(body))
	r.Body = io.NopCloser(bytes.NewReader(body))

	var alert models.Alert
	if err := json.NewDecoder(r.Body).Decode(&alert); err != nil {
		log.Printf("Failure to parse JSON: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// 加载配置文件中的条件和动作
	config, err := config.LoadConfig()
	if err != nil {
		log.Printf("load config fail: %v", err)
		http.Error(w, fmt.Sprintf("load config fail: %v", err), http.StatusInternalServerError)
		return
	}

	// 匹配告警并执行相应的动作
	result, err := matchAndExecute(alert, config.Conditions, config.Actions)
	if err != nil {
		log.Printf("Failed to execute action: %v", err)
		http.Error(w, fmt.Sprintf("Failed to execute action: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Result: %s", result)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(result))
}

// matchAndExecute 匹配告警并执行对应的动作
func matchAndExecute(alert models.Alert, conditions []models.Condition, actions map[string]models.Action) (string, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var resultMessage string
	var firstError error
	// 使用 goroutines 并发执行所有匹配的 action
	for _, condition := range conditions {
		if conditionMatches(alert, condition) {
			action, found := actions[condition.Action]
			if !found {
				continue
			}

			// 对每个条件匹配的 action 使用 goroutine 执行
			wg.Add(1)
			go func(action models.Action, condition models.Condition) {
				defer wg.Done()
				if err := Action.ExecuteAction(action, condition); err != nil {
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
	}

	// 等待所有 goroutines 完成
	wg.Wait()

	if firstError != nil {
		return "Action executed failed", firstError
	}

	return resultMessage, nil
}

// conditionMatches 检查告警是否满足条件
func conditionMatches(alert models.Alert, condition models.Condition) bool {
	for key, value := range condition.Fields {
		if !matchValue(key, value, alert) {
			return false
		}
	}
	return true
}

// matchValue 检查告警中的字段是否匹配
func matchValue(key, value string, alert models.Alert) bool {
	switch strings.ToLower(key) {
	case "receiver":
		return value == alert.Receiver
	case "status":
		return value == alert.Status
	case "externalurl":
		return value == alert.ExternalURL
	case "version":
		return value == alert.Version
	case "groupkey":
		return value == alert.GroupKey
	case "truncatedalerts":
		return value == fmt.Sprintf("%d", alert.TruncatedAlerts)
	default:
		// 检查在 Alerts 中的 Labels 或 Annotations
		for _, alertDetail := range alert.Alerts {
			if matchInLabelsOrAnnotations(key, value, alertDetail) {
				return true
			}
		}

		// 检查 GroupLabels, CommonLabels 和 CommonAnnotations
		if matchInGroupOrCommon(key, value, alert) {
			return true
		}
	}

	return false
}

// matchInLabelsOrAnnotations 检查 key-value 是否匹配在 Labels 或 Annotations 中
func matchInLabelsOrAnnotations(key, value string, alertDetail models.AlertDetails) bool {
	if v, ok := alertDetail.Labels[key]; ok && v == value {
		return true
	}
	if v, ok := alertDetail.Annotations[key]; ok && v == value {
		return true
	}
	return false
}

// matchInGroupOrCommon 检查 key-value 是否匹配在 GroupLabels, CommonLabels 或 CommonAnnotations 中
func matchInGroupOrCommon(key, value string, alert models.Alert) bool {
	if v, ok := alert.GroupLabels[key]; ok && v == value {
		return true
	}
	if v, ok := alert.CommonLabels[key]; ok && v == value {
		return true
	}
	if v, ok := alert.CommonAnnotations[key]; ok && v == value {
		return true
	}
	return false
}
