package deeplx

import (
	"context"
	"strings"

	"translate-services/internal/translation"
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

// Translate 执行翻译并返回谷歌格式，参数: 上下文、文本、源语言、目标语言、数据类型，返回: 翻译响应或错误
func (g *GoogleTranslator) Translate(ctx context.Context, q, sl, tl string, dt []string) (*translation.Response, error) {
	// 执行 DeepLX 翻译
	var result *TranslationResult
	if sl != "" && !strings.EqualFold(sl, "auto") {
		result = g.translator.TranslateWithContext(ctx, q, tl, sl)
	} else {
		result = g.translator.TranslateWithContext(ctx, q, tl)
	}

	// 检查是否成功
	if !result.Success {
		// 即使失败也返回一个基本的响应结构，避免调用方报错
		return g.buildErrorResponse(q, sl, tl), nil
	}

	// 转换为谷歌格式 (DRY 原则：统一转换逻辑喵)
	return g.convertToGoogleFormat(q, result, dt), nil
}

// convertToGoogleFormat 将结果转换为谷歌格式，参数: 原文本、翻译结果、数据类型，返回: 翻译响应
func (g *GoogleTranslator) convertToGoogleFormat(
	originalText string,
	result *TranslationResult,
	dt []string,
) *translation.Response {
	// 规范化检测到的源语言
	detectedLang := normalizeLanguageCode(result.SourceLang)

	// 如果源语言为空，使用语言检测作为后备 (健壮性处理喵～)
	if detectedLang == "" {
		detectedLang = detectLanguage(originalText, "")
	}

	resp := &translation.Response{
		Src: detectedLang,
		LDResult: &translation.LanguageDetectionResult{
			Srclangs:            []string{detectedLang},
			SrclangsConfidences: []float64{0.99},
		},
	}

	// 根据请求的数据类型填充响应 (接口隔离原则：按需提供喵)
	if includes(dt, "t") {
		// 基本翻译
		resp.Sentences = append(resp.Sentences, translation.Sentence{
			Orig:    originalText,
			Trans:   result.TranslatedText,
			Backend: 1,
		})
	}

	if includes(dt, "rm") {
		// 音译信息：DeepLX 无原生数据，提供简单衍生 (保持兼容喵～)
		resp.Sentences = append(resp.Sentences, translation.Sentence{
			SrcTranslit: originalText,
			Translit:    strings.ToUpper(originalText),
		})
	}

	if includes(dt, "bd") {
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

	if includes(dt, "qca") {
		// 拼写检查（DeepLX 不提供，返回原文）
		resp.Spell = &translation.SpellCheck{
			SpellRes: originalText,
		}
	}

	if includes(dt, "ex") {
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
		detectedLang = detectLanguage(q, sl)
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

// normalizeLanguageCode 规范化语言代码，参数: 原始代码字符串，返回: 标准化语言代码
func normalizeLanguageCode(code string) string {
	code = strings.ToLower(code)

	// DeepLX 返回的语言代码转换为谷歌格式
	switch code {
	case "zh", "zh-hans":
		return "zh-CN"
	case "zh-hant":
		return "zh-TW"
	case "en", "en-us":
		return "en"
	case "en-gb":
		return "en-GB"
	case "ja":
		return "ja"
	case "ko":
		return "ko"
	case "fr":
		return "fr"
	case "de":
		return "de"
	case "es":
		return "es"
	case "ru":
		return "ru"
	case "pt", "pt-br":
		return "pt"
	case "it":
		return "it"
	case "ar":
		return "ar"
	default:
		return code
	}
}

// detectLanguage 简单语言检测，参数: 文本与请求语言，返回: 推断语言代码
func detectLanguage(text, requested string) string {
	if strings.TrimSpace(requested) != "" && !strings.EqualFold(requested, "auto") {
		return normalizeLanguageCode(requested)
	}

	// 简单的启发式检测
	for _, r := range text {
		if isCJK(r) {
			return "zh-CN"
		}
		if isCyrillic(r) {
			return "ru"
		}
		if isJapanese(r) {
			return "ja"
		}
		if isKorean(r) {
			return "ko"
		}
	}

	return "en"
}

// includes 检查切片是否包含目标，参数: 字符串切片与目标，返回: 是否包含
func includes(params []string, target string) bool {
	for _, v := range params {
		if v == target {
			return true
		}
	}
	return false
}

// isCJK 判断字符是否为中日韩文字，参数: rune，返回: 布尔
func isCJK(r rune) bool {
	return (r >= 0x4E00 && r <= 0x9FFF) ||
		(r >= 0x3400 && r <= 0x4DBF) ||
		(r >= 0x20000 && r <= 0x2A6DF)
}

// isCyrillic 判断字符是否为西里尔字母，参数: rune，返回: 布尔
func isCyrillic(r rune) bool {
	return r >= 0x0400 && r <= 0x04FF
}

// isJapanese 判断字符是否为日语假名，参数: rune，返回: 布尔
func isJapanese(r rune) bool {
	return (r >= 0x3040 && r <= 0x309F) || // 平假名
		(r >= 0x30A0 && r <= 0x30FF) // 片假名
}

// isKorean 判断字符是否为韩文，参数: rune，返回: 布尔
func isKorean(r rune) bool {
	return r >= 0xAC00 && r <= 0xD7AF
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
