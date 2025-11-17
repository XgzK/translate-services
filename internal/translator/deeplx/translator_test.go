package deeplx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

// 测试用的 API 密钥常量
const testAPIKey = "sk-test-key-123456"

// TestNewTranslator 测试翻译器创建，参数: 测试实例，返回: 无
func TestNewTranslator(t *testing.T) {
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
			translator, err := NewTranslator(tt.apiKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTranslator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && translator == nil {
				t.Error("NewTranslator() 返回了 nil 翻译器")
			}
		})
	}
}

// TestNewTranslatorWithClient 测试使用自定义客户端创建翻译器，参数: 测试实例，返回: 无
func TestNewTranslatorWithClient(t *testing.T) {
	customClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	translator, err := NewTranslatorWithClient(testAPIKey, customClient)
	if err != nil {
		t.Fatalf("NewTranslatorWithClient() error = %v", err)
	}

	if translator.httpClient != customClient {
		t.Error("翻译器未使用自定义 HTTP 客户端")
	}
}

// TestSetBaseURL 测试设置自定义 URL，参数: 测试实例，返回: 无
func TestSetBaseURL(t *testing.T) {
	translator, _ := NewTranslator(testAPIKey)

	customURL := "https://custom.example.com/api"
	translator.SetBaseURL(customURL)

	if translator.baseURL != customURL {
		t.Errorf("SetBaseURL() baseURL = %v, want %v", translator.baseURL, customURL)
	}

	// 测试去除尾部斜杠
	translator.SetBaseURL(customURL + "/")
	if translator.baseURL != customURL {
		t.Errorf("SetBaseURL() 未正确去除尾部斜杠")
	}
}

// TestBuildURL 测试 URL 构建逻辑，参数: 测试实例，返回: 无
func TestBuildURL(t *testing.T) {
	translator, _ := NewTranslator(testAPIKey)

	tests := []struct {
		name     string
		model    string
		expected string
	}{
		{
			name:     "默认 URL",
			model:    "",
			expected: "https://deeplx.jayogo.com/translate/sk-test-key-123456",
		},
		{
			name:     "指定模型 URL",
			model:    "gpt-4",
			expected: "https://deeplx.jayogo.com/translate/sk-test-key-123456/gpt-4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := translator.buildURL(tt.model)
			if url != tt.expected {
				t.Errorf("buildURL() = %v, want %v", url, tt.expected)
			}
		})
	}
}

// mockServerHandler 模拟服务器处理函数，参数: 响应与请求，返回: 无
func mockServerHandler(w http.ResponseWriter, r *http.Request) {
	// 验证请求方法
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 验证 Content-Type
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Invalid Content-Type", http.StatusBadRequest)
		return
	}

	// 解析请求
	var req TranslationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// 简单的模拟翻译逻辑
	translatedText := "模拟翻译: " + req.Text
	sourceLang := req.SourceLang
	if sourceLang == "" {
		sourceLang = "AUTO"
	}

	// 返回模拟响应
	resp := TranslationResponse{
		Alternatives: []string{},
		Code:         200,
		Data:         translatedText,
		ID:           12345,
		Method:       "Mock",
		SourceLang:   sourceLang,
		TargetLang:   req.TargetLang,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// TestTranslate 测试基本翻译功能，参数: 测试实例，返回: 无
func TestTranslate(t *testing.T) {
	// 创建模拟服务器
	server := httptest.NewServer(http.HandlerFunc(mockServerHandler))
	defer server.Close()

	// 创建翻译器并设置为模拟服务器 URL
	translator, _ := NewTranslator(testAPIKey)
	translator.SetBaseURL(server.URL)

	tests := []struct {
		name       string
		text       string
		targetLang string
		sourceLang string
		wantErr    bool
	}{
		{
			name:       "指定源语言翻译",
			text:       "Hello, world!",
			targetLang: "ZH",
			sourceLang: "EN",
			wantErr:    false,
		},
		{
			name:       "自动检测源语言",
			text:       "你好，世界！",
			targetLang: "EN",
			sourceLang: "",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result *TranslationResult
			if tt.sourceLang != "" {
				result = translator.Translate(tt.text, tt.targetLang, tt.sourceLang)
			} else {
				result = translator.Translate(tt.text, tt.targetLang)
			}

			if result.Success != !tt.wantErr {
				t.Errorf("Translate() success = %v, wantErr %v, error: %s",
					result.Success, tt.wantErr, result.ErrorMessage)
				return
			}

			if result.Success {
				if result.TranslatedText == "" {
					t.Error("翻译结果为空")
				}
				if result.TargetLang != strings.ToUpper(tt.targetLang) {
					t.Errorf("目标语言 = %v, want %v", result.TargetLang, tt.targetLang)
				}
			}
		})
	}
}

// TestTranslateWithModel 测试指定模型翻译，参数: 测试实例，返回: 无
func TestTranslateWithModel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(mockServerHandler))
	defer server.Close()

	translator, _ := NewTranslator(testAPIKey)
	translator.SetBaseURL(server.URL)

	result := translator.TranslateWithModel("Test text", "ZH", "gpt-4", "EN")

	if !result.Success {
		t.Errorf("TranslateWithModel() failed: %s", result.ErrorMessage)
	}
}

// TestTranslateError 测试错误处理，参数: 测试实例，返回: 无
func TestTranslateError(t *testing.T) {
	// 创建返回错误的模拟服务器
	errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer errorServer.Close()

	translator, _ := NewTranslator(testAPIKey)
	translator.SetBaseURL(errorServer.URL)

	result := translator.Translate("Test", "ZH")

	if result.Success {
		t.Error("应该返回错误但返回了成功")
	}

	if result.ErrorMessage == "" {
		t.Error("错误信息为空")
	}
}

// TestTranslateTimeout 测试超时处理，参数: 测试实例，返回: 无
func TestTranslateTimeout(t *testing.T) {
	// 创建会超时的服务器
	timeoutServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // 延迟响应
	}))
	defer timeoutServer.Close()

	// 创建超时时间很短的客户端
	shortTimeoutClient := &http.Client{
		Timeout: 100 * time.Millisecond,
	}

	translator, _ := NewTranslatorWithClient(testAPIKey, shortTimeoutClient)
	translator.SetBaseURL(timeoutServer.URL)

	result := translator.Translate("Test", "ZH")

	if result.Success {
		t.Error("应该因为超时而失败")
	}

	if !strings.Contains(result.ErrorMessage, "请求失败") {
		t.Errorf("错误信息不符合预期: %s", result.ErrorMessage)
	}
}

// BenchmarkTranslate 性能基准测试，参数: 基准测试实例，返回: 无
func BenchmarkTranslate(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(mockServerHandler))
	defer server.Close()

	translator, _ := NewTranslator(testAPIKey)
	translator.SetBaseURL(server.URL)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		translator.Translate("Benchmark test", "ZH", "EN")
	}
}

// ExampleDeepLXTranslator_Translate 使用示例，参数: 无，返回: 无
func ExampleDeepLXTranslator_Translate() {
	// 注意：这是示例代码，实际使用时需要真实的 API 密钥
	apiKey := os.Getenv("DEEPLX_API_KEY")
	if apiKey == "" {
		apiKey = "sk-your-api-key"
	}

	translator, err := NewTranslator(apiKey)
	if err != nil {
		panic(err)
	}

	// 基本翻译
	result := translator.Translate("Hello, world!", "ZH", "EN")
	if result.Success {
		println(result.TranslatedText)
	}
}
