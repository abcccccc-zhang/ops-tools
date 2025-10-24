package action

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"strings"

	"alter-api/models"
)

// ExecuteAction 根据动作类型执行相应的操作
func ExecuteAction(action models.Action, conditions models.Condition, msg string) error {
	ip := ""
	if strings.Contains(action.Command, "{{ip}}") {
		re := regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
		ip = re.FindString(msg)
	}
	log.Printf("msg: %s", msg)
	switch action.Type {
	case "command":
		cmdStr := strings.ReplaceAll(action.Command, "{{ip}}", ip)
		log.Printf("exec command: %s", cmdStr)
		cmd := exec.Command("sh", "-c", cmdStr)

		output, err := cmd.CombinedOutput()
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
