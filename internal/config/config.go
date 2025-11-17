package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const defaultConfigPath = "config.yaml"

// Config 服务配置 (统一的配置管理喵～)
type Config struct {
	// 服务端口
	Port string `yaml:"port"`

	// 是否启用调试模式
	Debug bool `yaml:"debug"`

	// 翻译服务配置
	Translation TranslationConfig `yaml:"translation"`
}

// TranslationConfig 翻译服务配置 (灵活选择 API 地址与类型喵)
type TranslationConfig struct {
	ServiceType string `yaml:"service_type"`
	APIKey      string `yaml:"api_key"`
	BaseURL     string `yaml:"base_url"`
}

// Load 从配置文件与环境变量加载配置，参数: 无，返回: 配置指针与可能的错误
func Load() (*Config, error) {
	cfg := defaultConfig()

	if err := loadFromFile(cfg); err != nil {
		return nil, err
	}

	applyEnvOverrides(cfg)
	return cfg, nil
}

// Validate 验证配置，参数: 接收者 Config，返回: 校验失败时的错误
func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("配置不能为空")
	}

	if err := validatePort(c.Port); err != nil {
		return err
	}

	if err := validateTranslation(&c.Translation); err != nil {
		return err
	}

	return nil
}

// validateTranslation 校验翻译配置，参数: TranslationConfig 指针，返回: 验证失败的错误
func validateTranslation(t *TranslationConfig) error {
	if t == nil {
		return fmt.Errorf("translation 配置不能为空")
	}

	if strings.TrimSpace(t.ServiceType) == "" {
		return fmt.Errorf("translation.service_type 未设置")
	}

	if strings.TrimSpace(t.APIKey) == "" {
		return fmt.Errorf("translation.api_key 未设置")
	}

	return nil
}

// validatePort 校验端口，参数: 端口字符串，返回: 无效端口的错误
func validatePort(port string) error {
	port = strings.TrimSpace(port)
	if port == "" {
		return fmt.Errorf("端口配置无效")
	}

	value, err := strconv.Atoi(port)
	if err != nil || value <= 0 || value > 65535 {
		return fmt.Errorf("端口配置无效: %s", port)
	}

	return nil
}

// defaultConfig 构建默认配置，参数: 无，返回: 默认配置指针
func defaultConfig() *Config {
	return &Config{
		Port:  "8080",
		Debug: false,
		Translation: TranslationConfig{
			ServiceType: "deeplx",
		},
	}
}

// loadFromFile 从文件加载配置，参数: 目标配置指针，返回: 读取或解析时的错误
func loadFromFile(cfg *Config) error {
	path := strings.TrimSpace(os.Getenv("CONFIG_FILE"))
	if path == "" {
		path = defaultConfigPath
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	return nil
}

// applyEnvOverrides 应用环境变量覆盖，参数: 目标配置指针，返回: 无
func applyEnvOverrides(cfg *Config) {
	if v := strings.TrimSpace(os.Getenv("PORT")); v != "" {
		cfg.Port = v
	}

	if v := strings.TrimSpace(os.Getenv("DEBUG")); v != "" {
		cfg.Debug = parseBool(v)
	}

	if v := strings.TrimSpace(firstNonEmpty(
		os.Getenv("TRANSLATION_SERVICE"),
		os.Getenv("DEEPLX_SERVICE"),
	)); v != "" {
		cfg.Translation.ServiceType = v
	}

	if v := strings.TrimSpace(firstNonEmpty(
		os.Getenv("TRANSLATION_API_KEY"),
		os.Getenv("DEEPLX_API_KEY"),
	)); v != "" {
		cfg.Translation.APIKey = v
	}

	if v := strings.TrimSpace(firstNonEmpty(
		os.Getenv("TRANSLATION_BASE_URL"),
		os.Getenv("DEEPLX_BASE_URL"),
	)); v != "" {
		cfg.Translation.BaseURL = v
	}
}

// parseBool 解析布尔环境变量，参数: 字符串，返回: 布尔值
func parseBool(v string) bool {
	v = strings.ToLower(strings.TrimSpace(v))
	switch v {
	case "1", "true", "t", "yes", "y", "on":
		return true
	default:
		return false
	}
}

// firstNonEmpty 返回首个非空字符串，参数: 可变字符串列表，返回: 第一个非空值
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
