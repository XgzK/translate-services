package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestValidate 测试配置校验逻辑，参数: 测试实例，返回: 无
func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: Config{
				Port:  "8080",
				Debug: false,
				Translation: TranslationConfig{
					ServiceType: "deeplx",
					APIKey:      "sk-test",
					BaseURL:     "https://example.com",
				},
			},
			wantErr: false,
		},
		{
			name: "missing api key",
			cfg: Config{
				Port:  "8080",
				Debug: false,
				Translation: TranslationConfig{
					ServiceType: "deeplx",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid port",
			cfg: Config{
				Port: "70000",
				Translation: TranslationConfig{
					ServiceType: "deeplx",
					APIKey:      "sk-test",
				},
			},
			wantErr: true,
		},
		{
			name: "missing service type",
			cfg: Config{
				Port: "8080",
				Translation: TranslationConfig{
					APIKey: "sk-test",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestLoadFromFile 测试从文件加载配置，参数: 测试实例，返回: 无
func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	data := `
port: "9000"
debug: true
translation:
  service_type: "custom"
  api_key: "sk-file"
  base_url: "https://custom.example.com"
`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatalf("写入配置文件失败: %v", err)
	}

	t.Setenv("CONFIG_FILE", path)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Port != "9000" || !cfg.Debug {
		t.Fatalf("Load() 未解析顶层字段: %#v", cfg)
	}
	if cfg.Translation.ServiceType != "custom" ||
		cfg.Translation.APIKey != "sk-file" ||
		cfg.Translation.BaseURL != "https://custom.example.com" {
		t.Fatalf("Load() 未解析 translation 字段: %#v", cfg.Translation)
	}
}

// TestLoadEnvOverrides 测试环境变量覆盖配置，参数: 测试实例，返回: 无
func TestLoadEnvOverrides(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CONFIG_FILE", filepath.Join(dir, "missing.yaml"))
	t.Setenv("PORT", "9100")
	t.Setenv("DEBUG", "true")
	t.Setenv("TRANSLATION_SERVICE", "custom")
	t.Setenv("TRANSLATION_API_KEY", "sk-env")
	t.Setenv("TRANSLATION_BASE_URL", "https://env.example.com")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Port != "9100" || !cfg.Debug {
		t.Fatalf("环境变量未覆盖顶层字段: %#v", cfg)
	}
	if cfg.Translation.ServiceType != "custom" ||
		cfg.Translation.APIKey != "sk-env" ||
		cfg.Translation.BaseURL != "https://env.example.com" {
		t.Fatalf("环境变量未覆盖 translation 字段: %#v", cfg.Translation)
	}
}
