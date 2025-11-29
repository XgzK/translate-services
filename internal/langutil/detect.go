package langutil

import "strings"

// DetectLanguage 简单语言检测，参数: 文本与请求语言，返回: 推断语言代码
func DetectLanguage(text, requested string) string {
	if strings.TrimSpace(requested) != "" && !strings.EqualFold(requested, "auto") {
		return NormalizeLanguageCode(requested)
	}

	// 简单的启发式检测
	for _, r := range text {
		if IsCJK(r) {
			return "zh-CN"
		}
		if IsCyrillic(r) {
			return "ru"
		}
		if IsJapanese(r) {
			return "ja"
		}
		if IsKorean(r) {
			return "ko"
		}
	}

	return "en"
}

// NormalizeLanguageCode 规范化语言代码，参数: 原始代码字符串，返回: 标准化语言代码
func NormalizeLanguageCode(code string) string {
	code = strings.ToLower(code)

	// 语言代码转换为谷歌格式
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

// IsCJK 判断字符是否为中日韩文字，参数: rune，返回: 布尔
func IsCJK(r rune) bool {
	return (r >= 0x4E00 && r <= 0x9FFF) ||
		(r >= 0x3400 && r <= 0x4DBF) ||
		(r >= 0x20000 && r <= 0x2A6DF)
}

// IsCyrillic 判断字符是否为西里尔字母，参数: rune，返回: 布尔
func IsCyrillic(r rune) bool {
	return r >= 0x0400 && r <= 0x04FF
}

// IsJapanese 判断字符是否为日语假名，参数: rune，返回: 布尔
func IsJapanese(r rune) bool {
	return (r >= 0x3040 && r <= 0x309F) || // 平假名
		(r >= 0x30A0 && r <= 0x30FF) // 片假名
}

// IsKorean 判断字符是否为韩文，参数: rune，返回: 布尔
func IsKorean(r rune) bool {
	return r >= 0xAC00 && r <= 0xD7AF
}
