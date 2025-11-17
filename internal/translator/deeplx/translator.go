package deeplx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// TranslationRequest 翻译请求结构，参数: 无，返回: 无
type TranslationRequest struct {
	Text       string `json:"text"`
	SourceLang string `json:"source_lang,omitempty"` // omitempty: 为空时不发送
	TargetLang string `json:"target_lang"`
}

// TranslationResponse DeepLX API 响应结构，参数: 无，返回: 无
type TranslationResponse struct {
	Alternatives []string `json:"alternatives"`
	Code         int      `json:"code"`
	Data         string   `json:"data"`
	ID           int64    `json:"id"`
	Method       string   `json:"method"`
	SourceLang   string   `json:"source_lang"`
	TargetLang   string   `json:"target_lang"`
}

// TranslationResult 翻译结果封装，参数: 无，返回: 无
type TranslationResult struct {
	Success        bool
	TranslatedText string
	SourceLang     string
	TargetLang     string
	ErrorMessage   string
	RawResponse    *TranslationResponse
}

// Translator 翻译器接口，参数: 无，返回: 无
type Translator interface {
	Translate(text, targetLang string, sourceLang ...string) *TranslationResult
	TranslateWithModel(text, targetLang, model string, sourceLang ...string) *TranslationResult
}

// DeepLXTranslator DeepLX 翻译器实现，参数: 无，返回: 无
type DeepLXTranslator struct {
	apiKey          string
	baseURL         string
	httpClient      *http.Client // 复用 HTTP 客户端，提高性能喵
	requestTimeout  time.Duration
	maxRetryAttempt int
}

// NewTranslator 创建翻译器实例，参数: API 密钥，返回: DeepLXTranslator 指针或错误
func NewTranslator(apiKey string) (*DeepLXTranslator, error) {
	if apiKey == "" || !strings.HasPrefix(apiKey, "sk-") {
		return nil, fmt.Errorf("API 密钥必须以 sk- 开头")
	}

	return &DeepLXTranslator{
		apiKey:  apiKey,
		baseURL: "https://deeplx.jayogo.com/translate",
		httpClient: &http.Client{
			Timeout: 30 * time.Second, // 设置超时，避免长时间等待
		},
		requestTimeout:  10 * time.Second,
		maxRetryAttempt: 2,
	}, nil
}

// NewTranslatorWithClient 使用自定义客户端创建翻译器，参数: API 密钥与 HTTP 客户端，返回: DeepLXTranslator 指针或错误
func NewTranslatorWithClient(apiKey string, client *http.Client) (*DeepLXTranslator, error) {
	if apiKey == "" || !strings.HasPrefix(apiKey, "sk-") {
		return nil, fmt.Errorf("API 密钥必须以 sk- 开头")
	}

	return &DeepLXTranslator{
		apiKey:          apiKey,
		baseURL:         "https://deeplx.jayogo.com/translate",
		httpClient:      client,
		requestTimeout:  10 * time.Second,
		maxRetryAttempt: 2,
	}, nil
}

// Translate 执行翻译，参数: 文本、目标语言、可选源语言，返回: 翻译结果
func (t *DeepLXTranslator) Translate(text, targetLang string, sourceLang ...string) *TranslationResult {
	return t.TranslateWithContext(context.Background(), text, targetLang, sourceLang...)
}

// TranslateWithModel 使用指定模型翻译，参数: 文本、目标语言、模型、可选源语言，返回: 翻译结果
func (t *DeepLXTranslator) TranslateWithModel(text, targetLang, model string, sourceLang ...string) *TranslationResult {
	return t.TranslateWithModelContext(context.Background(), text, targetLang, model, sourceLang...)
}

// TranslateWithContext 带 context 的翻译请求，参数: 上下文、文本、目标语言、可选源语言，返回: 翻译结果
func (t *DeepLXTranslator) TranslateWithContext(ctx context.Context, text, targetLang string, sourceLang ...string) *TranslationResult {
	req := TranslationRequest{
		Text:       text,
		TargetLang: strings.ToUpper(targetLang),
	}

	if len(sourceLang) > 0 && sourceLang[0] != "" {
		req.SourceLang = strings.ToUpper(sourceLang[0])
	}

	return t.doRequest(ctx, req, "")
}

// TranslateWithModelContext 带 context 的模型翻译请求，参数: 上下文、文本、目标语言、模型、可选源语言，返回: 翻译结果
func (t *DeepLXTranslator) TranslateWithModelContext(ctx context.Context, text, targetLang, model string, sourceLang ...string) *TranslationResult {
	req := TranslationRequest{
		Text:       text,
		TargetLang: strings.ToUpper(targetLang),
	}

	if len(sourceLang) > 0 && sourceLang[0] != "" {
		req.SourceLang = strings.ToUpper(sourceLang[0])
	}

	return t.doRequest(ctx, req, model)
}

// SetBaseURL 设置自定义基础 URL，参数: 新的基础地址，返回: 无
func (t *DeepLXTranslator) SetBaseURL(baseURL string) {
	t.baseURL = strings.TrimSuffix(baseURL, "/")
}

// doRequest 执行 HTTP 请求，参数: 上下文、翻译请求、模型名称，返回: 翻译结果
func (t *DeepLXTranslator) doRequest(ctx context.Context, req TranslationRequest, model string) *TranslationResult {
	// 构建 URL
	url := t.buildURL(model)

	// 序列化请求体
	jsonData, err := json.Marshal(req)
	if err != nil {
		return &TranslationResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("序列化请求失败: %v", err),
		}
	}

	if ctx == nil {
		ctx = context.Background()
	}

	var lastErr string

	for attempt := 0; attempt <= t.maxRetryAttempt; attempt++ {
		if err := ctx.Err(); err != nil {
			return &TranslationResult{
				Success:      false,
				ErrorMessage: fmt.Sprintf("请求已取消: %v", err),
			}
		}

		reqCtx := ctx
		var cancel context.CancelFunc
		if t.requestTimeout > 0 {
			reqCtx, cancel = context.WithTimeout(ctx, t.requestTimeout)
		}

		// 创建 HTTP 请求
		httpReq, err := http.NewRequestWithContext(reqCtx, "POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			if cancel != nil {
				cancel()
			}
			return &TranslationResult{
				Success:      false,
				ErrorMessage: fmt.Sprintf("创建请求失败: %v", err),
			}
		}

		httpReq.Header.Set("Content-Type", "application/json")

		// 发送请求
		resp, err := t.httpClient.Do(httpReq)
		if err != nil {
			if cancel != nil {
				cancel()
			}
			lastErr = fmt.Sprintf("请求失败: %v", err)
			if t.shouldRetry(err) && attempt < t.maxRetryAttempt {
				time.Sleep(t.backoff(attempt))
				continue
			}
			return &TranslationResult{
				Success:      false,
				ErrorMessage: lastErr,
			}
		}

		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if cancel != nil {
			cancel()
		}
		if readErr != nil {
			lastErr = fmt.Sprintf("读取响应失败: %v", readErr)
			if attempt < t.maxRetryAttempt {
				time.Sleep(t.backoff(attempt))
				continue
			}
			return &TranslationResult{
				Success:      false,
				ErrorMessage: lastErr,
			}
		}

		// 检查状态码
		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body))
			if t.shouldRetryStatus(resp.StatusCode) && attempt < t.maxRetryAttempt {
				time.Sleep(t.backoff(attempt))
				continue
			}
			return &TranslationResult{
				Success:      false,
				ErrorMessage: lastErr,
			}
		}

		// 解析响应
		var translationResp TranslationResponse
		if err := json.Unmarshal(body, &translationResp); err != nil {
			lastErr = fmt.Sprintf("解析响应失败: %v", err)
			if attempt < t.maxRetryAttempt {
				time.Sleep(t.backoff(attempt))
				continue
			}
			return &TranslationResult{
				Success:      false,
				ErrorMessage: lastErr,
			}
		}

		return &TranslationResult{
			Success:        true,
			TranslatedText: translationResp.Data,
			SourceLang:     translationResp.SourceLang,
			TargetLang:     translationResp.TargetLang,
			RawResponse:    &translationResp,
		}
	}

	return &TranslationResult{
		Success:      false,
		ErrorMessage: lastErr,
	}
}

// buildURL 构建请求 URL，参数: 模型名称，返回: 完整 URL 字符串
func (t *DeepLXTranslator) buildURL(model string) string {
	if model != "" {
		return fmt.Sprintf("%s/%s/%s", t.baseURL, t.apiKey, model)
	}
	return fmt.Sprintf("%s/%s", t.baseURL, t.apiKey)
}

// shouldRetry 判断错误是否需重试，参数: 错误对象，返回: 布尔
func (t *DeepLXTranslator) shouldRetry(err error) bool {
	if err == nil {
		return false
	}
	if ne, ok := err.(net.Error); ok && (ne.Timeout() || ne.Temporary()) {
		return true
	}
	return false
}

// shouldRetryStatus 判断状态码是否需重试，参数: 状态码，返回: 布尔
func (t *DeepLXTranslator) shouldRetryStatus(status int) bool {
	// 对 5xx 等服务器错误进行重试
	return status >= 500 && status < 600
}

// backoff 计算退避时间，参数: 重试次数，返回: 时间间隔
func (t *DeepLXTranslator) backoff(attempt int) time.Duration {
	// 线性退避，避免过长阻塞
	return time.Duration(200*(attempt+1)) * time.Millisecond
}
