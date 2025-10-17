package models

import "time"

// Config 应用程序配置
type Config struct {
	Stock   StockConfig   `yaml:"stock"`
	WeChat  WeChatConfig  `yaml:"wechat"`
	Monitor MonitorConfig `yaml:"monitor"`
}

// StockConfig 股票相关配置
type StockConfig struct {
	Code      string  `yaml:"code"`      // 股票代码，如 "sh000001"
	Name      string  `yaml:"name"`      // 股票名称
	Threshold float64 `yaml:"threshold"` // 波动阈值，如 0.8 表示 0.8%
}

// WeChatConfig 企业微信配置
type WeChatConfig struct {
	WebhookURL string `yaml:"webhook_url"` // 企业微信机器人 webhook URL
}

// MonitorConfig 监视器配置
type MonitorConfig struct {
	Interval time.Duration `yaml:"interval"` // 检查间隔
}