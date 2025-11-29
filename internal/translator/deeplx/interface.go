package deeplx

import (
	"context"

	"github.com/XgzK/translate-services/internal/translation"
)

// TranslationService 通用翻译服务接口，参数: 无，返回: 无
// 所有翻译服务提供商都应实现此接口
type TranslationService interface {
	// Translate 执行翻译并返回谷歌格式响应，参数: 上下文、文本、源语言、目标语言、数据类型，返回: 翻译响应与错误
	// q: 待翻译文本
	// sl: 源语言代码 (可选，为 "auto" 或空则自动检测)
	// tl: 目标语言代码
	// dt: 请求的数据类型 (t=翻译, rm=音译, bd=词典, qca=拼写检查, ex=示例)
	Translate(ctx context.Context, q, sl, tl string, dt []string) (*translation.Response, error)

	// TranslateWithModel 使用指定模型执行翻译，参数: 上下文、文本、源语言、目标语言、数据类型、模型名称，返回: 翻译响应与错误
	// q: 待翻译文本
	// sl: 源语言代码 (可选，为 "auto" 或空则自动检测)
	// tl: 目标语言代码
	// dt: 请求的数据类型 (t=翻译, rm=音译, bd=词典, qca=拼写检查, ex=示例)
	// model: 模型名称 (如: gpt-3.5-turbo, gemini-1.5-pro-latest 等)
	TranslateWithModel(ctx context.Context, q, sl, tl string, dt []string, model string) (*translation.Response, error)

	// GetName 返回服务提供商名称，参数: 无，返回: 名称字符串
	GetName() string

	// IsAvailable 检查服务是否可用，参数: 无，返回: 布尔值
	IsAvailable() bool
}

// TranslationServiceConfig 翻译服务配置 (统一的配置接口喵)
type TranslationServiceConfig struct {
	APIKey  string // API 密钥
	BaseURL string // 基础 URL（可选）
	Timeout int    // 超时时间（秒）
}
