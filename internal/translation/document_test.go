package translation

import (
	"strings"
	"testing"
)

// TestBuildDocumentResponse_Basic 测试基本文档响应，参数: 测试实例，返回: 无
func TestBuildDocumentResponse_Basic(t *testing.T) {
	result := BuildDocumentResponse("<p>Hello</p>", "zh")

	if len(result) != 1 {
		t.Fatalf("外层长度 = %d, want 1", len(result))
	}
	if len(result[0]) != 1 {
		t.Fatalf("中层长度 = %d, want 1", len(result[0]))
	}
	if len(result[0][0]) != 2 {
		t.Fatalf("内层长度 = %d, want 2", len(result[0][0]))
	}
}

// TestBuildDocumentResponse_ChineseContent 测试中文内容，参数: 测试实例，返回: 无
func TestBuildDocumentResponse_ChineseContent(t *testing.T) {
	// 使用 auto 让函数自动检测语言
	result := BuildDocumentResponse("你好世界", "auto")

	// 检测语言应该是 zh-CN
	detected := result[0][0][1]
	if detected != "zh-CN" {
		t.Errorf("detected = %v, want zh-CN", detected)
	}
}

// TestBuildDocumentResponse_SameLanguage 测试相同语言，参数: 测试实例，返回: 无
func TestBuildDocumentResponse_SameLanguage(t *testing.T) {
	result := BuildDocumentResponse("Hello", "en")

	// 当检测语言与目标语言相同时，不应包装
	translated := result[0][0][0]
	if strings.Contains(translated, "<p>") {
		t.Error("相同语言时不应添加包装标签")
	}
}

// TestBuildDocumentResponse_DifferentLanguage 测试不同语言，参数: 测试实例，返回: 无
func TestBuildDocumentResponse_DifferentLanguage(t *testing.T) {
	result := BuildDocumentResponse("Hello", "zh")

	translated := result[0][0][0]
	// 当需要翻译时，应该包装
	if !strings.Contains(translated, "Hello") {
		t.Error("翻译结果应包含原文")
	}
}

// TestBuildDocumentResponse_EmptyContent 测试空内容，参数: 测试实例，返回: 无
func TestBuildDocumentResponse_EmptyContent(t *testing.T) {
	result := BuildDocumentResponse("", "zh")

	if len(result) != 1 || len(result[0]) != 1 || len(result[0][0]) != 2 {
		t.Error("空内容应返回正确结构")
	}
}

// TestBuildDocumentResponse_WhitespaceContent 测试空白内容，参数: 测试实例，返回: 无
func TestBuildDocumentResponse_WhitespaceContent(t *testing.T) {
	result := BuildDocumentResponse("   ", "zh")

	translated := result[0][0][0]
	// 空白内容不应被包装
	if strings.Contains(translated, "<p>") {
		t.Error("空白内容不应添加包装标签")
	}
}

// TestBuildDocumentResponse_HTMLContent 测试 HTML 内容，参数: 测试实例，返回: 无
func TestBuildDocumentResponse_HTMLContent(t *testing.T) {
	html := "<div><span>Test</span></div>"
	result := BuildDocumentResponse(html, "zh")

	translated := result[0][0][0]
	if !strings.Contains(translated, "Test") {
		t.Error("HTML 内容应被保留")
	}
}

// TestBuildDocumentResponse_AutoDetect 测试自动检测语言，参数: 测试实例，返回: 无
func TestBuildDocumentResponse_AutoDetect(t *testing.T) {
	tests := []struct {
		name       string
		html       string
		wantDetect string
	}{
		{"中文", "你好", "zh-CN"},
		{"日语", "こんにちは", "ja"},
		{"韩语", "안녕", "ko"},
		{"俄语", "Привет", "ru"},
		{"英语", "Hello", "en"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildDocumentResponse(tt.html, "auto")
			detected := result[0][0][1]
			if detected != tt.wantDetect {
				t.Errorf("detected = %v, want %v", detected, tt.wantDetect)
			}
		})
	}
}
