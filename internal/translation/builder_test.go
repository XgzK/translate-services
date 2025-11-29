package translation

import (
	"strings"
	"testing"
)

// TestBuildResponse_BasicTranslation 测试基本翻译响应，参数: 测试实例，返回: 无
func TestBuildResponse_BasicTranslation(t *testing.T) {
	resp := BuildResponse("你好", "auto", "en", []string{"t"})

	if resp.Src != "zh-CN" {
		t.Errorf("Src = %v, want zh-CN", resp.Src)
	}
	if len(resp.Sentences) != 1 {
		t.Fatalf("Sentences count = %d, want 1", len(resp.Sentences))
	}
	if resp.Sentences[0].Orig != "你好" {
		t.Errorf("Orig = %v, want 你好", resp.Sentences[0].Orig)
	}
	if !strings.Contains(resp.Sentences[0].Trans, "你好") {
		t.Errorf("Trans should contain original text")
	}
}

// TestBuildResponse_WithRomanization 测试音译响应，参数: 测试实例，返回: 无
func TestBuildResponse_WithRomanization(t *testing.T) {
	resp := BuildResponse("hello", "en", "zh", []string{"t", "rm"})

	if len(resp.Sentences) != 2 {
		t.Fatalf("Sentences count = %d, want 2 (t + rm)", len(resp.Sentences))
	}
	// 第二个句子应该是音译
	if resp.Sentences[1].Translit != "HELLO" {
		t.Errorf("Translit = %v, want HELLO", resp.Sentences[1].Translit)
	}
}

// TestBuildResponse_WithDictionary 测试词典响应，参数: 测试实例，返回: 无
func TestBuildResponse_WithDictionary(t *testing.T) {
	resp := BuildResponse("test", "en", "zh", []string{"bd"})

	if len(resp.Dict) == 0 {
		t.Fatal("Dict should not be empty")
	}
	if resp.Dict[0].Pos != "noun" {
		t.Errorf("Dict[0].Pos = %v, want noun", resp.Dict[0].Pos)
	}
	if len(resp.AlternativeTranslations) == 0 {
		t.Fatal("AlternativeTranslations should not be empty")
	}
}

// TestBuildResponse_WithSpellCheck 测试拼写检查响应，参数: 测试实例，返回: 无
func TestBuildResponse_WithSpellCheck(t *testing.T) {
	resp := BuildResponse("  hello  ", "en", "zh", []string{"qca"})

	if resp.Spell == nil {
		t.Fatal("Spell should not be nil")
	}
	if resp.Spell.SpellRes != "hello" {
		t.Errorf("SpellRes = %v, want hello (trimmed)", resp.Spell.SpellRes)
	}
}

// TestBuildResponse_WithExamples 测试示例响应，参数: 测试实例，返回: 无
func TestBuildResponse_WithExamples(t *testing.T) {
	resp := BuildResponse("word", "en", "zh", []string{"ex"})

	if resp.Examples == nil {
		t.Fatal("Examples should not be nil")
	}
	if len(resp.Examples.Examples) == 0 {
		t.Fatal("Examples.Examples should not be empty")
	}
	if !strings.Contains(resp.Examples.Examples[0].Text, "word") {
		t.Errorf("Example text should contain 'word'")
	}
}

// TestBuildResponse_AllParams 测试所有参数组合，参数: 测试实例，返回: 无
func TestBuildResponse_AllParams(t *testing.T) {
	resp := BuildResponse("hello", "en", "zh", []string{"t", "rm", "bd", "qca", "ex"})

	// 验证所有字段都被填充
	if len(resp.Sentences) != 2 {
		t.Errorf("Sentences count = %d, want 2", len(resp.Sentences))
	}
	if len(resp.Dict) == 0 {
		t.Error("Dict should not be empty")
	}
	if resp.Spell == nil {
		t.Error("Spell should not be nil")
	}
	if resp.Examples == nil {
		t.Error("Examples should not be nil")
	}
	if resp.LDResult == nil {
		t.Error("LDResult should not be nil")
	}
}

// TestBuildResponse_EmptyDt 测试空数据类型，参数: 测试实例，返回: 无
func TestBuildResponse_EmptyDt(t *testing.T) {
	resp := BuildResponse("hello", "en", "zh", []string{})

	if len(resp.Sentences) != 0 {
		t.Errorf("Sentences should be empty, got %d", len(resp.Sentences))
	}
	if len(resp.Dict) != 0 {
		t.Error("Dict should be empty")
	}
}

// TestBuildResponse_LanguageDetection 测试语言检测，参数: 测试实例，返回: 无
func TestBuildResponse_LanguageDetection(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		sl      string
		wantSrc string
	}{
		{"中文自动检测", "你好世界", "auto", "zh-CN"},
		{"日语自动检测", "こんにちは", "auto", "ja"},
		{"韩语自动检测", "안녕하세요", "auto", "ko"},
		{"俄语自动检测", "Привет", "auto", "ru"},
		{"英语自动检测", "Hello", "auto", "en"},
		{"指定语言", "任意文本", "FR", "fr"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := BuildResponse(tt.text, tt.sl, "en", []string{"t"})
			if resp.Src != tt.wantSrc {
				t.Errorf("Src = %v, want %v", resp.Src, tt.wantSrc)
			}
		})
	}
}

// TestBuildResponse_LDResult 测试语言检测结果结构，参数: 测试实例，返回: 无
func TestBuildResponse_LDResult(t *testing.T) {
	resp := BuildResponse("hello", "en", "zh", []string{"t"})

	if resp.LDResult == nil {
		t.Fatal("LDResult should not be nil")
	}
	if len(resp.LDResult.Srclangs) != 1 {
		t.Errorf("Srclangs length = %d, want 1", len(resp.LDResult.Srclangs))
	}
	if len(resp.LDResult.SrclangsConfidences) != 1 {
		t.Errorf("SrclangsConfidences length = %d, want 1", len(resp.LDResult.SrclangsConfidences))
	}
	if resp.LDResult.SrclangsConfidences[0] != 0.99 {
		t.Errorf("Confidence = %v, want 0.99", resp.LDResult.SrclangsConfidences[0])
	}
}
