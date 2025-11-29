package langutil

import "testing"

// TestIncludes 测试切片包含检查，参数: 测试实例，返回: 无
func TestIncludes(t *testing.T) {
	tests := []struct {
		name   string
		params []string
		target string
		want   bool
	}{
		{
			name:   "包含目标 - 首位",
			params: []string{"t", "rm", "bd"},
			target: "t",
			want:   true,
		},
		{
			name:   "包含目标 - 中间",
			params: []string{"t", "rm", "bd"},
			target: "rm",
			want:   true,
		},
		{
			name:   "包含目标 - 末位",
			params: []string{"t", "rm", "bd"},
			target: "bd",
			want:   true,
		},
		{
			name:   "不包含目标",
			params: []string{"t", "rm"},
			target: "bd",
			want:   false,
		},
		{
			name:   "空切片",
			params: []string{},
			target: "t",
			want:   false,
		},
		{
			name:   "nil 切片",
			params: nil,
			target: "t",
			want:   false,
		},
		{
			name:   "空目标",
			params: []string{"t", "rm", "bd"},
			target: "",
			want:   false,
		},
		{
			name:   "单元素匹配",
			params: []string{"only"},
			target: "only",
			want:   true,
		},
		{
			name:   "单元素不匹配",
			params: []string{"only"},
			target: "other",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Includes(tt.params, tt.target)
			if got != tt.want {
				t.Errorf("Includes() = %v, want %v", got, tt.want)
			}
		})
	}
}

// BenchmarkIncludes 性能基准测试，参数: 基准实例，返回: 无
func BenchmarkIncludes(b *testing.B) {
	params := []string{"t", "rm", "bd", "qca", "ex", "at", "md", "ss"}
	target := "ss" // 最后一个元素

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Includes(params, target)
	}
}
