package deeplx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/XgzK/translate-services/internal/langutil"
)

// TestNewGoogleTranslator 测试谷歌翻译适配器创建，参数: 测试实例，返回: 无
func TestNewGoogleTranslator(t *testing.T) {
	tests := []struct {
		name    string
		apiKey  string
		wantErr bool
	}{
		{
			name:    "有效的 API 密钥",
			apiKey:  "sk-valid-key",
			wantErr: false,
		},
		{
			name:    "空 API 密钥",
			apiKey:  "",
			wantErr: true,
		},
		{
			name:    "无效前缀的 API 密钥",
			apiKey:  "invalid-key",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			translator, err := NewGoogleTranslator(tt.apiKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewGoogleTranslator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && translator == nil {
				t.Error("NewGoogleTranslator() 返回了 nil 翻译器")
			}
		})
	}
}

// TestGoogleTranslator_Translate 测试谷歌格式翻译，参数: 测试实例，返回: 无
func TestGoogleTranslator_Translate(t *testing.T) {
	// 创建模拟服务器
	server := httptest.NewServer(http.HandlerFunc(mockServerHandler))
	defer server.Close()

	// 创建适配器
	adapter, err := NewGoogleTranslator(testAPIKey)
	if err != nil {
		t.Fatalf("NewGoogleTranslator() error = %v", err)
	}

	// 设置模拟服务器 URL
	adapter.translator.SetBaseURL(server.URL)

	tests := []struct {
		name       string
		q          string
		sl         string
		tl         string
		dt         []string
		wantErr    bool
		checkField string // 要检查的字段
	}{
		{
			name:       "基本翻译 (t)",
			q:          "Hello, world!",
			sl:         "EN",
			tl:         "ZH",
			dt:         []string{"t"},
			wantErr:    false,
			checkField: "sentences",
		},
		{
			name:       "完整翻译 (所有字段)",
			q:          "Test",
			sl:         "EN",
			tl:         "ZH",
			dt:         []string{"t", "rm", "bd", "qca", "ex"},
			wantErr:    false,
			checkField: "all",
		},
		{
			name:       "自动检测语言",
			q:          "你好",
			sl:         "auto",
			tl:         "EN",
			dt:         []string{"t"},
			wantErr:    false,
			checkField: "ld_result",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := adapter.Translate(context.Background(), tt.q, tt.sl, tt.tl, tt.dt)

			if (err != nil) != tt.wantErr {
				t.Errorf("Translate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if resp == nil {
				t.Fatal("Translate() 返回了 nil 响应")
			}

			// 验证响应结构
			switch tt.checkField {
			case "sentences":
				if len(resp.Sentences) == 0 {
					t.Error("Sentences 为空")
				}
				if resp.Sentences[0].Orig != tt.q {
					t.Errorf("Orig = %v, want %v", resp.Sentences[0].Orig, tt.q)
				}
			case "ld_result":
				if resp.LDResult == nil {
					t.Error("LDResult 为空")
				}
				if len(resp.LDResult.Srclangs) == 0 {
					t.Error("Srclangs 为空")
				}
			case "all":
				// 检查所有字段都存在
				if len(resp.Sentences) == 0 {
					t.Error("Sentences 为空")
				}
				if len(resp.Dict) == 0 {
					t.Error("Dict 为空")
				}
				if resp.Spell == nil {
					t.Error("Spell 为空")
				}
				if resp.Examples == nil {
					t.Error("Examples 为空")
				}
			}

			// 验证源语言检测
			if resp.Src == "" {
				t.Error("Src (源语言) 为空")
			}
		})
	}
}

// TestNormalizeLanguageCode 测试语言代码规范化，参数: 测试实例，返回: 无
func TestNormalizeLanguageCode(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "中文简体",
			input: "ZH",
			want:  "zh-CN",
		},
		{
			name:  "中文简体 (ZH-HANS)",
			input: "zh-hans",
			want:  "zh-CN",
		},
		{
			name:  "中文繁体",
			input: "ZH-HANT",
			want:  "zh-TW",
		},
		{
			name:  "英语",
			input: "EN",
			want:  "en",
		},
		{
			name:  "日语",
			input: "JA",
			want:  "ja",
		},
		{
			name:  "其他语言",
			input: "FR",
			want:  "fr",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := langutil.NormalizeLanguageCode(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeLanguageCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDetectLanguage 测试语言检测，参数: 测试实例，返回: 无
func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		requested string
		want      string
	}{
		{
			name:      "中文文本",
			text:      "你好世界",
			requested: "auto",
			want:      "zh-CN",
		},
		{
			name:      "日语文本",
			text:      "こんにちは",
			requested: "auto",
			want:      "ja",
		},
		{
			name:      "韩文文本",
			text:      "안녕하세요",
			requested: "auto",
			want:      "ko",
		},
		{
			name:      "俄语文本",
			text:      "Привет",
			requested: "auto",
			want:      "ru",
		},
		{
			name:      "英文文本",
			text:      "Hello",
			requested: "auto",
			want:      "en",
		},
		{
			name:      "指定语言",
			text:      "任意文本",
			requested: "FR",
			want:      "fr",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := langutil.DetectLanguage(tt.text, tt.requested)
			if got != tt.want {
				t.Errorf("DetectLanguage() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIncludes 测试切片包含助手，参数: 测试实例，返回: 无
func TestIncludes(t *testing.T) {
	tests := []struct {
		name   string
		params []string
		target string
		want   bool
	}{
		{
			name:   "包含目标",
			params: []string{"t", "rm", "bd"},
			target: "t",
			want:   true,
		},
		{
			name:   "不包含目标",
			params: []string{"t", "rm"},
			target: "bd",
			want:   false,
		},
		{
			name:   "空切片",
			params: []string{},
			target: "t",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := langutil.Includes(tt.params, tt.target)
			if got != tt.want {
				t.Errorf("Includes() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestConvertToGoogleFormat 测试格式转换，参数: 测试实例，返回: 无
func TestConvertToGoogleFormat(t *testing.T) {
	adapter, _ := NewGoogleTranslator(testAPIKey)

	result := &TranslationResult{
		Success:        true,
		TranslatedText: "你好，世界！",
		SourceLang:     "EN",
		TargetLang:     "ZH",
	}

	tests := []struct {
		name string
		dt   []string
		want int // 期望的 sentences 数量
	}{
		{
			name: "只翻译",
			dt:   []string{"t"},
			want: 1,
		},
		{
			name: "翻译和音译",
			dt:   []string{"t", "rm"},
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := adapter.convertToGoogleFormat("Hello, world!", result, tt.dt)

			if len(resp.Sentences) != tt.want {
				t.Errorf("sentences 数量 = %v, want %v", len(resp.Sentences), tt.want)
			}

			if resp.Src == "" {
				t.Error("Src 为空")
			}

			if resp.LDResult == nil {
				t.Error("LDResult 为空")
			}
		})
	}
}

// TestBuildErrorResponse 测试错误响应构建，参数: 测试实例，返回: 无
func TestBuildErrorResponse(t *testing.T) {
	adapter, _ := NewGoogleTranslator(testAPIKey)

	resp := adapter.buildErrorResponse("Hello", "en", "zh")

	if resp == nil {
		t.Fatal("buildErrorResponse() 返回了 nil")
	}

	if len(resp.Sentences) == 0 {
		t.Error("错误响应的 Sentences 为空")
	}

	// 错误时应该返回原文
	if resp.Sentences[0].Trans != "Hello" {
		t.Errorf("错误响应的 Trans = %v, want %v", resp.Sentences[0].Trans, "Hello")
	}

}

// ExampleGoogleTranslator_Translate 使用示例，参数: 无，返回: 无
func ExampleGoogleTranslator_Translate() {
	// 创建适配器
	adapter, err := NewGoogleTranslator("sk-your-api-key")
	if err != nil {
		panic(err)
	}

	// 执行翻译，返回谷歌格式
	resp, err := adapter.Translate(
		context.Background(),
		"Hello, world!",
		"EN",
		"ZH",
		[]string{"t"}, // 只需要翻译文本
	)

	if err != nil {
		panic(err)
	}

	// 访问翻译结果
	if len(resp.Sentences) > 0 {
		println(resp.Sentences[0].Trans)
	}
}

// BenchmarkGoogleTranslate 性能基准测试，参数: 基准实例，返回: 无
func BenchmarkGoogleTranslate(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(mockServerHandler))
	defer server.Close()

	adapter, _ := NewGoogleTranslator(testAPIKey)
	adapter.translator.SetBaseURL(server.URL)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		adapter.Translate(context.Background(), "Benchmark test", "EN", "ZH", []string{"t"})
	}
}
