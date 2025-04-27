package action

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"

	"alter-api/models"
)

// ExecuteAction 根据动作类型执行相应的操作
func ExecuteAction(action models.Action, conditions models.Condition) error {
	switch action.Type {
	case "command":
		cmd := exec.Command("sh", "-c", action.Command)
		log.Printf("exec command: %s", action.Command)

		output, err := cmd.CombinedOutput() // 捕获标准输出和错误输出
		if err != nil {
			log.Printf("exec command fail: %v", err)
			log.Printf("command output: %s", string(output))
			return err
		}

		log.Printf("command output: %s", string(output))
		return nil

	case "webhook":
		client := &http.Client{}
		data, err := json.Marshal(action.Body)
		if err != nil {
			log.Printf("Failed to serialize webhook request body: %v", err)
			return err
		}

		req, err := http.NewRequest(action.Method, action.URL, bytes.NewBuffer(data))
		if err != nil {
			log.Printf("create webhook fail: %v", err)
			return err
		}
		for key, value := range action.Headers {
			req.Header.Set(key, value)
		}

		log.Printf("send webhook req to %s,method: %s", action.URL, action.Method)
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("send webhook req fail: %v", err)
			return err
		}
		defer resp.Body.Close()

		return nil

	default:
		log.Printf("unknow type: %s", action.Type)
		return fmt.Errorf("unknow type: %s", action.Type)
	}
}
