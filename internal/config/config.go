package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"stockWatcher/internal/models"
	"time"

	"gopkg.in/yaml.v3"
)

// Load 加载配置文件
func Load(filename string) (*models.Config, error) {
	// 检查配置文件是否存在
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// 如果不存在，创建默认配置文件
		if err := createDefaultConfig(filename); err != nil {
			return nil, fmt.Errorf("创建默认配置文件失败: %v", err)
		}
		return nil, fmt.Errorf("配置文件不存在，已创建默认配置文件 %s，请填写企业微信webhookURL后重新启动", filename)
	}

	// 读取配置文件
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 解析YAML
	var config models.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 验证配置
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("配置验证失败: %v", err)
	}

	// 设置默认值
	setDefaults(&config)

	return &config, nil
}

// createDefaultConfig 创建默认配置文件
func createDefaultConfig(filename string) error {
	defaultConfig := `# 股票指数监视器配置文件

# 股票配置
stock:
  code: "sh000001"        # 上证指数代码
  name: "上证指数"        # 股票名称
  threshold: 0.8          # 波动阈值(百分比)，0.8表示0.8%

# 企业微信机器人配置
wechat:
  webhook_url: ""         # 请填写企业微信机器人的webhook URL
  # 如何获取企业微信机器人webhook：
  # 1. 在企业微信群中添加机器人
  # 2. 复制机器人的webhook URL并填写到此处

# 监视器配置
monitor:
  interval: 30s          # 检查间隔，支持时间单位: s(秒), m(分钟), h(小时)
`

	// 确保目录存在
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}

	// 写入默认配置
	return ioutil.WriteFile(filename, []byte(defaultConfig), 0644)
}

// validateConfig 验证配置
func validateConfig(config *models.Config) error {
	// 验证股票配置
	if config.Stock.Code == "" {
		return fmt.Errorf("股票代码不能为空")
	}

	// 验证企业微信配置
	if config.WeChat.WebhookURL == "" {
		return fmt.Errorf("企业微信webhookURL不能为空，请先配置企业微信机器人")
	}

	// 验证监视器配置
	if config.Monitor.Interval <= 0 {
		return fmt.Errorf("监视器间隔必须大于0")
	}

	return nil
}

// setDefaults 设置默认值
func setDefaults(config *models.Config) {
	// 设置默认股票名称
	if config.Stock.Name == "" {
		nameMap := map[string]string{
			"sh000001": "上证指数",
			"sz399001": "深证成指",
			"sz399006": "创业板指",
		}
		if name, exists := nameMap[config.Stock.Code]; exists {
			config.Stock.Name = name
		}
	}

	// 设置默认波动阈值
	if config.Stock.Threshold == 0 {
		config.Stock.Threshold = 0.8
	}

	// 设置默认检查间隔
	if config.Monitor.Interval == 0 {
		config.Monitor.Interval = 30 * time.Second
	}
}

// Save 保存配置到文件
func Save(config *models.Config, filename string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	if err := ioutil.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	return nil
}