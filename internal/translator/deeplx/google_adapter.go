package deeplx

import (
	"context"
	"strings"

	"github.com/XgzK/translate-services/internal/langutil"
	"github.com/XgzK/translate-services/internal/translation"
)

// GoogleTranslator 谷歌翻译接口适配器 (适配器模式，让 DeepLX 兼容谷歌格式喵～)
// 实现 TranslationService 接口
type GoogleTranslator struct {
	translator *DeepLXTranslator
	name       string
}

// NewGoogleTranslator 创建谷歌翻译适配器，参数: API 密钥，返回: 适配器指针或错误
func NewGoogleTranslator(apiKey string) (*GoogleTranslator, error) {
	translator, err := NewTranslator(apiKey)
	if err != nil {
		return nil, err
	}

	return &GoogleTranslator{
		translator: translator,
		name:       "DeepLX",
	}, nil
}

// NewGoogleTranslatorWithConfig 使用配置创建适配器，参数: 配置对象，返回: 适配器指针或错误
func NewGoogleTranslatorWithConfig(config *TranslationServiceConfig) (*GoogleTranslator, error) {
	translator, err := NewTranslatorWithConfig(config)
	if err != nil {
		return nil, err
	}

	return &GoogleTranslator{
		translator: translator,
		name:       "DeepLX",
	}, nil
}

// NewGoogleTranslatorWithClient 使用自定义客户端创建适配器，参数: API 密钥与客户端对象，返回: 适配器指针或错误
func NewGoogleTranslatorWithClient(apiKey string, client interface{}) (*GoogleTranslator, error) {
	// 这里可以扩展支持自定义 HTTP 客户端
	translator, err := NewTranslator(apiKey)
	if err != nil {
		return nil, err
	}

	return &GoogleTranslator{
		translator: translator,
	}, nil
}

// translateFunc 翻译函数类型定义，用于抽象不同翻译方法
type translateFunc func(ctx context.Context, text, targetLang string, sourceLang ...string) *TranslationResult

// doTranslate 执行翻译的公共逻辑 (DRY 原则：抽取重复代码喵～)
// 参数: 上下文、文本、源语言、目标语言、数据类型、翻译函数，返回: 翻译响应或错误
func (g *GoogleTranslator) doTranslate(ctx context.Context, q, sl, tl string, dt []string, fn translateFunc) (*translation.Response, error) {
	var result *TranslationResult
	if sl != "" && !strings.EqualFold(sl, "auto") {
		result = fn(ctx, q, tl, sl)
	} else {
		result = fn(ctx, q, tl)
	}

	if !result.Success {
		// 即使失败也返回一个基本的响应结构，避免调用方报错
		return g.buildErrorResponse(q, sl, tl), nil
	}

	return g.convertToGoogleFormat(q, result, dt), nil
}

// Translate 执行翻译并返回谷歌格式，参数: 上下文、文本、源语言、目标语言、数据类型，返回: 翻译响应或错误
func (g *GoogleTranslator) Translate(ctx context.Context, q, sl, tl string, dt []string) (*translation.Response, error) {
	return g.doTranslate(ctx, q, sl, tl, dt, g.translator.TranslateWithContext)
}

// TranslateWithModel 使用指定模型执行翻译并返回谷歌格式，参数: 上下文、文本、源语言、目标语言、数据类型、模型名称，返回: 翻译响应或错误
func (g *GoogleTranslator) TranslateWithModel(ctx context.Context, q, sl, tl string, dt []string, model string) (*translation.Response, error) {
	// 使用闭包捕获 model 参数，适配统一的 translateFunc 签名
	fn := func(ctx context.Context, text, targetLang string, sourceLang ...string) *TranslationResult {
		return g.translator.TranslateWithModelContext(ctx, text, targetLang, model, sourceLang...)
	}
	return g.doTranslate(ctx, q, sl, tl, dt, fn)
}

// convertToGoogleFormat 将结果转换为谷歌格式，参数: 原文本、翻译结果、数据类型，返回: 翻译响应
func (g *GoogleTranslator) convertToGoogleFormat(
	originalText string,
	result *TranslationResult,
	dt []string,
) *translation.Response {
	// 规范化检测到的源语言
	detectedLang := langutil.NormalizeLanguageCode(result.SourceLang)

	// 如果源语言为空，使用语言检测作为后备 (健壮性处理喵～)
	if detectedLang == "" {
		detectedLang = langutil.DetectLanguage(originalText, "")
	}

	resp := &translation.Response{
		Src: detectedLang,
		LDResult: &translation.LanguageDetectionResult{
			Srclangs:            []string{detectedLang},
			SrclangsConfidences: []float64{0.99},
		},
	}

	// 根据请求的数据类型填充响应 (接口隔离原则：按需提供喵)
	if langutil.Includes(dt, "t") {
		// 基本翻译
		resp.Sentences = append(resp.Sentences, translation.Sentence{
			Orig:    originalText,
			Trans:   result.TranslatedText,
			Backend: 1,
		})
	}

	if langutil.Includes(dt, "rm") {
		// 音译信息：DeepLX 无原生数据，提供简单衍生 (保持兼容喵～)
		resp.Sentences = append(resp.Sentences, translation.Sentence{
			SrcTranslit: originalText,
			Translit:    strings.ToUpper(originalText),
		})
	}

	if langutil.Includes(dt, "bd") {
		// 词典和替代翻译（DeepLX 不提供详细词典，使用简化版本）
		resp.Dict = []translation.Dictionary{
			{
				Pos: "translation",
				Entry: []translation.DictEntry{
					{
						Word:               result.TranslatedText,
						ReverseTranslation: []string{originalText},
						Score:              0.95,
					},
				},
			},
		}
	}

	if langutil.Includes(dt, "qca") {
		// 拼写检查（DeepLX 不提供，返回原文）
		resp.Spell = &translation.SpellCheck{
			SpellRes: originalText,
		}
	}

	if langutil.Includes(dt, "ex") {
		// 示例（DeepLX 不提供，返回空）
		resp.Examples = &translation.Examples{
			Examples: []translation.Example{},
		}
	}

	return resp
}

// buildErrorResponse 构建错误响应，参数: 文本、源语言、目标语言，返回: 基本翻译响应
func (g *GoogleTranslator) buildErrorResponse(q, sl, tl string) *translation.Response {
	detectedLang := sl
	if detectedLang == "" || strings.EqualFold(detectedLang, "auto") {
		detectedLang = langutil.DetectLanguage(q, sl)
	}

	return &translation.Response{
		Src: detectedLang,
		Sentences: []translation.Sentence{
			{
				Orig:  q,
				Trans: q, // 翻译失败时返回原文
			},
		},
		LDResult: &translation.LanguageDetectionResult{
			Srclangs:            []string{detectedLang},
			SrclangsConfidences: []float64{0.5},
		},
	}
}

// ========== TranslationService 接口实现 ==========

// GetName 返回服务提供商名称，参数: 无，返回: 名称字符串
func (g *GoogleTranslator) GetName() string {
	return g.name
}

// IsAvailable 检查服务是否可用，参数: 无，返回: 布尔
func (g *GoogleTranslator) IsAvailable() bool {
	// 检查翻译器是否已初始化
	return g.translator != nil && g.translator.apiKey != ""
}

// SetName 设置服务名称，参数: 名称字符串，返回: 无
func (g *GoogleTranslator) SetName(name string) {
	g.name = name
}
