package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

const (
	// KeyPrefix 缓存键前缀
	KeyPrefix = "translate"
	// SharedServiceName 共享缓存的服务名
	SharedServiceName = "shared"
)

// KeyGenerator 缓存键生成器
type KeyGenerator struct {
	shareAcrossServices bool
}

// NewKeyGenerator 创建缓存键生成器
func NewKeyGenerator(shareAcrossServices bool) *KeyGenerator {
	return &KeyGenerator{
		shareAcrossServices: shareAcrossServices,
	}
}

// Generate 生成缓存键
// service: 服务提供商标识 (如 deeplx, google, baidu)
// text: 待翻译文本
// sourceLang: 源语言代码
// targetLang: 目标语言代码
// model: 翻译模型 (可选)
//
// 返回格式:
//   - 隔离模式: translate:{service}:{hash}
//   - 共享模式: translate:shared:{hash}
func (g *KeyGenerator) Generate(service, text, sourceLang, targetLang, model string) string {
	hash := g.computeHash(text, sourceLang, targetLang, model)

	if g.shareAcrossServices {
		return fmt.Sprintf("%s:%s:%s", KeyPrefix, SharedServiceName, hash)
	}
	return fmt.Sprintf("%s:%s:%s", KeyPrefix, strings.ToLower(service), hash)
}

// computeHash 计算输入内容的哈希值
// 使用 SHA256 并取前 16 个十六进制字符 (8 字节)
func (g *KeyGenerator) computeHash(text, sourceLang, targetLang, model string) string {
	// 规范化输入，确保相同内容产生相同的哈希
	normalized := fmt.Sprintf("%s|%s|%s|%s",
		strings.TrimSpace(text),
		strings.ToLower(strings.TrimSpace(sourceLang)),
		strings.ToLower(strings.TrimSpace(targetLang)),
		strings.ToLower(strings.TrimSpace(model)),
	)

	// 计算 SHA256 哈希
	hash := sha256.Sum256([]byte(normalized))

	// 取前 8 字节 (16 个十六进制字符)
	return hex.EncodeToString(hash[:8])
}

// GenerateCacheKey 便捷函数：生成缓存键 (默认隔离模式)
func GenerateCacheKey(service, text, sourceLang, targetLang, model string) string {
	return NewKeyGenerator(false).Generate(service, text, sourceLang, targetLang, model)
}

// GenerateSharedCacheKey 便捷函数：生成共享缓存键
func GenerateSharedCacheKey(text, sourceLang, targetLang, model string) string {
	return NewKeyGenerator(true).Generate("", text, sourceLang, targetLang, model)
}
