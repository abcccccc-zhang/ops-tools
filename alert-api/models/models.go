package models

type Alert struct {
	Receiver          string            `json:"receiver"`
	Status            string            `json:"status"`
	Alerts            []AlertDetails    `json:"alerts"`
	GroupLabels       map[string]string `json:"groupLabels"`
	CommonLabels      map[string]string `json:"commonLabels"`
	CommonAnnotations map[string]string `json:"commonAnnotations"`
	ExternalURL       string            `json:"externalURL"`
	Version           string            `json:"version"`
	GroupKey          string            `json:"groupKey"`
	TruncatedAlerts   int               `json:"truncatedAlerts"`
}

type AlertDetails struct {
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

type Condition struct {
	Fields map[string]string `yaml:"fields"` // 动态存储所有字段
	Action string            `yaml:"action"` // 执行动作的名称
}

type Action struct {
	Type    string            `yaml:"type"`
	Command string            `yaml:"command"`
	URL     string            `yaml:"url"`
	Method  string            `yaml:"method"`
	Headers map[string]string `yaml:"headers"`
	Body    map[string]string `yaml:"body"`
}

type Config struct {
	Conditions []Condition
	Actions    map[string]Action
	LogConfig  LogConfig
}

type LogConfig struct {
	Filename   string `yaml:"filename"`
	MaxSize    int    `yaml:"maxSize"`
	MaxBackups int    `yaml:"maxBackups"`
	MaxAge     int    `yaml:"maxAge"`
	Compress   bool   `yaml:"compress"`
}
