package translation

// Response 表示翻译响应，参数: 无，返回: 无
type Response struct {
	Src                     string                   `json:"src"`
	Sentences               []Sentence               `json:"sentences,omitempty"`
	Dict                    []Dictionary             `json:"dict,omitempty"`
	Spell                   *SpellCheck              `json:"spell,omitempty"`
	LDResult                *LanguageDetectionResult `json:"ld_result,omitempty"`
	AlternativeTranslations []AlternativeTranslation `json:"alternative_translations,omitempty"`
	Examples                *Examples                `json:"examples,omitempty"`
}

// Sentence 表示单句翻译结果，参数: 无，返回: 无
type Sentence struct {
	Orig        string `json:"orig,omitempty"`
	Trans       string `json:"trans,omitempty"`
	Backend     int    `json:"backend,omitempty"`
	SrcTranslit string `json:"src_translit,omitempty"`
	Translit    string `json:"translit,omitempty"`
}

// Dictionary 表示词典条目列表，参数: 无，返回: 无
type Dictionary struct {
	Pos   string      `json:"pos"`
	Entry []DictEntry `json:"entry"`
}

// DictEntry 表示词典的具体翻译项，参数: 无，返回: 无
type DictEntry struct {
	Word               string   `json:"word"`
	ReverseTranslation []string `json:"reverse_translation"`
	Score              float64  `json:"score"`
}

// SpellCheck 拼写检查结果，参数: 无，返回: 无
type SpellCheck struct {
	SpellRes string `json:"spell_res"`
}

// LanguageDetectionResult 语言检测结果，参数: 无，返回: 无
type LanguageDetectionResult struct {
	Srclangs            []string  `json:"srclangs"`
	SrclangsConfidences []float64 `json:"srclangs_confidences"`
}

// AlternativeTranslation 备选翻译，参数: 无，返回: 无
type AlternativeTranslation struct {
	SrcPhrase     string        `json:"src_phrase"`
	RawSrcSegment string        `json:"raw_src_segment"`
	Alternative   []Alternative `json:"alternative"`
}

// Alternative 单个备选翻译，参数: 无，返回: 无
type Alternative struct {
	WordPostproc      string  `json:"word_postproc"`
	Score             float64 `json:"score"`
	HasPrecedingSpace bool    `json:"has_preceding_space"`
	AttachToNextToken bool    `json:"attach_to_next_token"`
}

// Examples 示例集合，参数: 无，返回: 无
type Examples struct {
	Examples []Example `json:"example"`
}

// Example 单个例句，参数: 无，返回: 无
type Example struct {
	Text       string     `json:"text"`
	SourceType int        `json:"source_type"`
	LabelInfo  *LabelInfo `json:"label_info,omitempty"`
}

// LabelInfo 标签信息，参数: 无，返回: 无
type LabelInfo struct {
	Subject []string `json:"subject"`
}
