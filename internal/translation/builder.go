package translation

import (
	"fmt"
	"strings"

	"github.com/XgzK/translate-services/internal/langutil"
)

// BuildResponse 构造响应，参数: 文本q、源语言sl、目标语言tl、数据段dt，返回: 模拟的翻译响应
func BuildResponse(q, sl, tl string, dt []string) Response {
	detected := langutil.DetectLanguage(q, sl)
	resp := Response{
		Src: detected,
		LDResult: &LanguageDetectionResult{
			Srclangs:            []string{detected},
			SrclangsConfidences: []float64{0.99},
		},
	}

	if langutil.Includes(dt, "t") {
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

	if langutil.Includes(dt, "rm") {
		resp.Sentences = append(resp.Sentences, Sentence{
			SrcTranslit: q,
			Translit:    strings.ToUpper(q),
		})
	}

	if langutil.Includes(dt, "bd") {
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

	if langutil.Includes(dt, "qca") {
		resp.Spell = &SpellCheck{
			SpellRes: strings.TrimSpace(q),
		}
	}

	if langutil.Includes(dt, "ex") {
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
