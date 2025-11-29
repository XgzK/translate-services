package translation

import (
	"fmt"
	"strings"

	"github.com/XgzK/translate-services/internal/langutil"
)

// BuildDocumentResponse 构造文档翻译响应，参数: HTML内容与检测语言，返回: 嵌套数组结构
func BuildDocumentResponse(html, detected string) [][][]string {
	src := langutil.DetectLanguage(html, detected)
	translated := html
	if strings.TrimSpace(html) != "" && !strings.EqualFold(src, detected) && strings.TrimSpace(detected) != "" {
		translated = fmt.Sprintf("<p>%s (%s)</p>", html, detected)
	}
	return [][][]string{
		{
			{
				translated,
				src,
			},
		},
	}
}
