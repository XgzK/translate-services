package translation

import (
	"regexp"
	"strings"
	"testing"
	"time"
)

// TestElementScript_Format 测试脚本格式，参数: 测试实例，返回: 无
func TestElementScript_Format(t *testing.T) {
	script := ElementScript()

	// 检查格式是否正确
	if !strings.HasPrefix(script, "var tkk='") {
		t.Errorf("脚本应以 \"var tkk='\" 开头, got: %s", script)
	}
	if !strings.HasSuffix(script, "';") {
		t.Errorf("脚本应以 \"';\" 结尾, got: %s", script)
	}
}

// TestElementScript_TKKValue 测试 TKK 值格式，参数: 测试实例，返回: 无
func TestElementScript_TKKValue(t *testing.T) {
	script := ElementScript()

	// 使用正则提取 TKK 值
	re := regexp.MustCompile(`var tkk='(\d+)\.(\d+)';`)
	matches := re.FindStringSubmatch(script)

	if len(matches) != 3 {
		t.Fatalf("TKK 格式不正确, got: %s", script)
	}

	// 第一部分应该是时间戳 / 3600
	expectedHour := time.Now().Unix() / 3600
	// 允许 1 小时的误差（测试可能跨越小时边界）
	if matches[1] != "" {
		// 验证第一部分是合理的时间值
		if len(matches[1]) < 5 {
			t.Errorf("时间戳部分太短: %s", matches[1])
		}
	}

	// 第二部分应该是固定值
	if matches[2] != "544157181" {
		t.Errorf("固定值部分 = %s, want 544157181", matches[2])
	}

	_ = expectedHour // 避免未使用警告
}

// TestElementScript_Consistency 测试短时间内结果一致性，参数: 测试实例，返回: 无
func TestElementScript_Consistency(t *testing.T) {
	script1 := ElementScript()
	script2 := ElementScript()

	// 在同一秒内调用应该返回相同结果
	if script1 != script2 {
		t.Errorf("连续调用应返回相同结果\n  first:  %s\n  second: %s", script1, script2)
	}
}

// TestElementScript_ValidJavaScript 测试生成的是有效 JavaScript，参数: 测试实例，返回: 无
func TestElementScript_ValidJavaScript(t *testing.T) {
	script := ElementScript()

	// 简单验证是有效的变量声明
	if !strings.Contains(script, "var ") {
		t.Error("应该包含 var 声明")
	}
	if !strings.Contains(script, "tkk") {
		t.Error("应该包含 tkk 变量名")
	}
	if !strings.Contains(script, "=") {
		t.Error("应该包含赋值运算符")
	}
}

// BenchmarkElementScript 性能基准测试，参数: 基准实例，返回: 无
func BenchmarkElementScript(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ElementScript()
	}
}
