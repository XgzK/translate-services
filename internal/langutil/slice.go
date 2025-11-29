package langutil

// Includes 检查切片是否包含目标，参数: 字符串切片与目标，返回: 是否包含
func Includes(params []string, target string) bool {
	for _, v := range params {
		if v == target {
			return true
		}
	}
	return false
}
