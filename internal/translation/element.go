package translation

import (
	"fmt"
	"time"
)

// ElementScript 模拟 element.js，参数: 无，返回: 含 TKK 的脚本字符串
func ElementScript() string {
	now := time.Now().Unix() / 3600
	return fmt.Sprintf("var tkk='%d.544157181';", now)
}
