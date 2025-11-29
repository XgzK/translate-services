package langutil

import "testing"

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
			name:      "日语文本 (平假名)",
			text:      "こんにちは",
			requested: "auto",
			want:      "ja",
		},
		{
			name:      "日语文本 (片假名)",
			text:      "カタカナ",
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
		{
			name:      "空请求语言",
			text:      "Hello World",
			requested: "",
			want:      "en",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectLanguage(tt.text, tt.requested)
			if got != tt.want {
				t.Errorf("DetectLanguage() = %v, want %v", got, tt.want)
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
		{"中文简体 ZH", "ZH", "zh-CN"},
		{"中文简体 zh-hans", "zh-hans", "zh-CN"},
		{"中文繁体", "ZH-HANT", "zh-TW"},
		{"英语", "EN", "en"},
		{"英语 en-us", "en-us", "en"},
		{"英式英语", "EN-GB", "en-GB"},
		{"日语", "JA", "ja"},
		{"韩语", "KO", "ko"},
		{"法语", "FR", "fr"},
		{"德语", "DE", "de"},
		{"西班牙语", "ES", "es"},
		{"俄语", "RU", "ru"},
		{"葡萄牙语", "PT", "pt"},
		{"葡萄牙语巴西", "pt-br", "pt"},
		{"意大利语", "IT", "it"},
		{"阿拉伯语", "AR", "ar"},
		{"未知语言", "unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeLanguageCode(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeLanguageCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIsCJK 测试中日韩字符检测，参数: 测试实例，返回: 无
func TestIsCJK(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want bool
	}{
		{"中文字符 你", '你', true},
		{"中文字符 好", '好', true},
		{"英文字符", 'A', false},
		{"数字", '1', false},
		{"日语平假名", 'あ', false}, // 平假名不是CJK基本区
		{"韩文", '한', false},        // 韩文不是CJK基本区
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsCJK(tt.r)
			if got != tt.want {
				t.Errorf("IsCJK() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIsCyrillic 测试西里尔字符检测，参数: 测试实例，返回: 无
func TestIsCyrillic(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want bool
	}{
		{"俄语字符 П", 'П', true},
		{"俄语字符 р", 'р', true},
		{"英文字符", 'A', false},
		{"中文字符", '中', false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsCyrillic(tt.r)
			if got != tt.want {
				t.Errorf("IsCyrillic() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIsJapanese 测试日语假名检测，参数: 测试实例，返回: 无
func TestIsJapanese(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want bool
	}{
		{"平假名 あ", 'あ', true},
		{"平假名 ん", 'ん', true},
		{"片假名 ア", 'ア', true},
		{"片假名 ン", 'ン', true},
		{"英文字符", 'A', false},
		{"中文字符", '中', false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsJapanese(tt.r)
			if got != tt.want {
				t.Errorf("IsJapanese() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIsKorean 测试韩文检测，参数: 测试实例，返回: 无
func TestIsKorean(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want bool
	}{
		{"韩文 한", '한', true},
		{"韩文 글", '글', true},
		{"英文字符", 'A', false},
		{"中文字符", '中', false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsKorean(tt.r)
			if got != tt.want {
				t.Errorf("IsKorean() = %v, want %v", got, tt.want)
			}
		})
	}
}
