package models

import "github.com/prometheus/alertmanager/template"

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
type TemplateConfig struct {
	CustomPath string `yaml:"custom_path"`
}
type Configuration struct {
	Webhook  string         `yaml:"webhook"`
	Sign     string         `yaml:"sign"`
	Template TemplateConfig `yaml:"template"`
}

type Config struct {
	Conditions    []Condition
	Actions       map[string]Action
	LogConfig     LogConfig
	Configuration map[string]Configuration
}

type LogConfig struct {
	Filename   string `yaml:"filename"`
	MaxSize    int    `yaml:"maxSize"`
	MaxBackups int    `yaml:"maxBackups"`
	MaxAge     int    `yaml:"maxAge"`
	Compress   bool   `yaml:"compress"`
}

type WebhookMessage struct {
	// reference: https://prometheus.io/docs/alerting/latest/notifications/
	template.Data
	OpenIDs []string
	// 用于存储 AlertManager webhook 请求带来的数据，比如 query string
	Meta template.KV
	// 仅内置模板中使用，自定义模板中访问是空数组
	// 目前没有发现在 {{template defined_name .}} 后对其结果进行进一步处理的方式
	// 首先，通过模板，将每个 Alert 转为字符串，大段文本都在 content 字段，需要注意转义。
	FiringAlerts   []string
	ResolvedAlerts []string
	TitlePrefix    string
}

type Alerts template.Alert
