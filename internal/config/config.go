package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const defaultConfigPath = "config.yaml"

// Config 服务配置 (统一的配置管理喵～)
type Config struct {
	// 服务端口
	Port string `yaml:"port"`

	// 是否启用调试模式
	Debug bool `yaml:"debug"`

	// 服务器配置
	Server ServerConfig `yaml:"server"`

	// 翻译服务配置
	Translation TranslationConfig `yaml:"translation"`

	// 缓存配置
	Cache CacheConfig `yaml:"cache"`
}

// ServerConfig 服务器配置 (超时与性能相关喵～)
type ServerConfig struct {
	RequestTimeout  int `yaml:"request_timeout"`  // 翻译请求超时 (秒)，默认 8
	MiddlewareTimeout int `yaml:"middleware_timeout"` // 中间件超时 (秒)，默认 12
	ShutdownTimeout int `yaml:"shutdown_timeout"` // 优雅停机超时 (秒)，默认 15
}

// TranslationConfig 翻译服务配置 (灵活选择 API 地址与类型喵)
type TranslationConfig struct {
	ServiceType string `yaml:"service_type"`
	APIKey      string `yaml:"api_key"`
	BaseURL     string `yaml:"base_url"`
	Model       string `yaml:"model"`   // 默认使用的模型 (如: gpt-3.5-turbo, gemini-1.5-pro-latest 等)
	Timeout     int    `yaml:"timeout"` // 翻译请求超时 (秒)，默认 10
}

// CacheConfig Redis 缓存配置 (提升性能，减少 API 调用喵～)
type CacheConfig struct {
	// 基础配置
	Enabled  bool   `yaml:"enabled"`  // 是否启用缓存
	Addr     string `yaml:"addr"`     // Redis 地址，如 "localhost:6379"
	Password string `yaml:"password"` // Redis 密码
	DB       int    `yaml:"db"`       // 数据库编号

	// 缓存策略
	TTL                 string `yaml:"ttl"`                    // 缓存过期时间，如 "24h"，空或 "0" 表示永不过期
	ShareAcrossServices bool   `yaml:"share_across_services"` // 不同服务共享缓存

	// 连接池配置
	PoolSize     int `yaml:"pool_size"`     // 连接池大小，默认 10
	DialTimeout  int `yaml:"dial_timeout"`  // 连接超时 (秒)，默认 5
	ReadTimeout  int `yaml:"read_timeout"`  // 读取超时 (秒)，默认 3
	WriteTimeout int `yaml:"write_timeout"` // 写入超时 (秒)，默认 3
}

// GetTTL 获取 TTL 时间，返回 0 表示永不过期
func (c *CacheConfig) GetTTL() time.Duration {
	if c.TTL == "" || c.TTL == "0" {
		return 0 // 永不过期
	}
	d, err := time.ParseDuration(c.TTL)
	if err != nil {
		return 0 // 解析失败，默认永不过期
	}
	return d
}

// GetPoolSize 获取连接池大小
func (c *CacheConfig) GetPoolSize() int {
	if c.PoolSize <= 0 {
		return 10
	}
	return c.PoolSize
}

// GetDialTimeout 获取连接超时时间
func (c *CacheConfig) GetDialTimeout() time.Duration {
	if c.DialTimeout <= 0 {
		return 5 * time.Second
	}
	return time.Duration(c.DialTimeout) * time.Second
}

// GetReadTimeout 获取读取超时时间
func (c *CacheConfig) GetReadTimeout() time.Duration {
	if c.ReadTimeout <= 0 {
		return 3 * time.Second
	}
	return time.Duration(c.ReadTimeout) * time.Second
}

// GetWriteTimeout 获取写入超时时间
func (c *CacheConfig) GetWriteTimeout() time.Duration {
	if c.WriteTimeout <= 0 {
		return 3 * time.Second
	}
	return time.Duration(c.WriteTimeout) * time.Second
}

// GetRequestTimeout 获取翻译请求超时时间，返回秒数
func (c *ServerConfig) GetRequestTimeout() int {
	if c.RequestTimeout <= 0 {
		return 8 // 默认 8 秒
	}
	return c.RequestTimeout
}

// GetMiddlewareTimeout 获取中间件超时时间，返回秒数
func (c *ServerConfig) GetMiddlewareTimeout() int {
	if c.MiddlewareTimeout <= 0 {
		return 12 // 默认 12 秒
	}
	return c.MiddlewareTimeout
}

// GetShutdownTimeout 获取优雅停机超时时间，返回秒数
func (c *ServerConfig) GetShutdownTimeout() int {
	if c.ShutdownTimeout <= 0 {
		return 15 // 默认 15 秒
	}
	return c.ShutdownTimeout
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
		Cache: CacheConfig{
			Enabled:             false,
			Addr:                "localhost:6379",
			DB:                  0,
			TTL:                 "", // 空表示永不过期
			ShareAcrossServices: true,
			PoolSize:            10,
			DialTimeout:         5,
			ReadTimeout:         3,
			WriteTimeout:        3,
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

	if v := strings.TrimSpace(firstNonEmpty(
		os.Getenv("TRANSLATION_MODEL"),
		os.Getenv("DEEPLX_MODEL"),
	)); v != "" {
		cfg.Translation.Model = v
	}

	// 缓存配置环境变量覆盖
	if v := strings.TrimSpace(os.Getenv("CACHE_ENABLED")); v != "" {
		cfg.Cache.Enabled = parseBool(v)
	}

	if v := strings.TrimSpace(os.Getenv("CACHE_ADDR")); v != "" {
		cfg.Cache.Addr = v
	}

	if v := strings.TrimSpace(os.Getenv("CACHE_PASSWORD")); v != "" {
		cfg.Cache.Password = v
	}

	if v := strings.TrimSpace(os.Getenv("CACHE_DB")); v != "" {
		if db, err := strconv.Atoi(v); err == nil {
			cfg.Cache.DB = db
		}
	}

	if v := strings.TrimSpace(os.Getenv("CACHE_TTL")); v != "" {
		cfg.Cache.TTL = v
	}

	if v := strings.TrimSpace(os.Getenv("CACHE_SHARE_ACROSS_SERVICES")); v != "" {
		cfg.Cache.ShareAcrossServices = parseBool(v)
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
