package translation

import (
	"fmt"
	"strings"
)

// BuildResponse 构造响应，参数: 文本q、源语言sl、目标语言tl、数据段dt，返回: 模拟的翻译响应
func BuildResponse(q, sl, tl string, dt []string) Response {
	detected := detectLanguage(q, sl)
	resp := Response{
		Src: detected,
		LDResult: &LanguageDetectionResult{
			Srclangs:            []string{detected},
			SrclangsConfidences: []float64{0.99},
		},
	}

	if includes(dt, "t") {
		transText := q
		if !strings.EqualFold(detected, tl) && strings.TrimSpace(tl) != "" {
			transText = fmt.Sprintf("%s (%s)", q, tl)
		}
		resp.Sentences = append(resp.Sentences, Sentence{
			Orig:    q,
			Trans:   transText,
			Backend: 1,
		})
	}

	if includes(dt, "rm") {
		resp.Sentences = append(resp.Sentences, Sentence{
			SrcTranslit: q,
			Translit:    strings.ToUpper(q),
		})
	}

	if includes(dt, "bd") {
		resp.Dict = []Dictionary{
			{
				Pos: "noun",
				Entry: []DictEntry{
					{
						Word:               fmt.Sprintf("%s-%s", q, tl),
						ReverseTranslation: []string{q, "sample"},
						Score:              0.95,
					},
				},
			},
		}
		resp.AlternativeTranslations = []AlternativeTranslation{
			{
				SrcPhrase:     q,
				RawSrcSegment: q,
				Alternative: []Alternative{
					{
						WordPostproc:      fmt.Sprintf("%s alt", q),
						Score:             0.88,
						HasPrecedingSpace: true,
						AttachToNextToken: false,
					},
				},
			},
		}
	}

	if includes(dt, "qca") {
		resp.Spell = &SpellCheck{
			SpellRes: strings.TrimSpace(q),
		}
	}

	if includes(dt, "ex") {
		resp.Examples = &Examples{
			Examples: []Example{
				{
					Text:       fmt.Sprintf("<b>%s</b> example usage.", q),
					SourceType: 1,
					LabelInfo: &LabelInfo{
						Subject: []string{"general"},
					},
				},
			},
		}
	}

	return resp
}

// includes 判断切片中是否包含目标，参数: 字符串切片与目标值，返回: 是否包含
func includes(params []string, target string) bool {
	for _, v := range params {
		if v == target {
			return true
		}
	}
	return false
}

// normalizeDetectedLang 规范化源语言，参数: 源语言字符串sl，返回: 标准化语言代码
func normalizeDetectedLang(sl string) string {
	if strings.EqualFold(sl, "auto") || strings.TrimSpace(sl) == "" {
		return "en"
	}
	return sl
}

// detectLanguage 简单检测语言，参数: 文本和请求的语言，返回: 推断语言代码
func detectLanguage(text, requested string) string {
	if strings.TrimSpace(requested) != "" && !strings.EqualFold(requested, "auto") {
		return requested
	}

	for _, r := range text {
		if isCJK(r) {
			return "zh-CN"
		}
		if isCyrillic(r) {
			return "ru"
		}
	}

	return "en"
}

// isCJK 判断字符是否为中日韩字符，参数: rune，返回: 布尔值
func isCJK(r rune) bool {
	return (r >= 0x4E00 && r <= 0x9FFF) ||
		(r >= 0x3400 && r <= 0x4DBF) ||
		(r >= 0x20000 && r <= 0x2A6DF)
}

// isCyrillic 判断字符是否为西里尔字符，参数: rune，返回: 布尔值
func isCyrillic(r rune) bool {
	return r >= 0x0400 && r <= 0x04FF
}
