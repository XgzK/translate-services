// Package cache 提供翻译结果的缓存功能
package cache

import (
	"context"
	"time"
)

// Cache 定义缓存操作接口
// 支持多种缓存后端实现（Redis、内存等）
type Cache interface {
	// Get 获取缓存值
	// 返回 nil, nil 表示缓存未命中
	Get(ctx context.Context, key string) ([]byte, error)

	// Set 设置缓存值
	// ttl 为 0 表示永不过期
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// Delete 删除缓存
	Delete(ctx context.Context, key string) error

	// Ping 检查连接是否正常
	Ping(ctx context.Context) error

	// Close 关闭连接
	Close() error
}

// CachedTranslation 统一的缓存值结构
// 支持所有翻译服务提供商的结果存储
type CachedTranslation struct {
	// ========== 原始请求信息 ==========
	OriginalText string `json:"original_text"`  // 翻译前的原始文本
	SourceLang   string `json:"source_lang"`    // 源语言代码 (可能是 auto 检测后的结果)
	TargetLang   string `json:"target_lang"`    // 目标语言代码

	// ========== 翻译结果 ==========
	TranslatedText string   `json:"translated_text"`        // 主要翻译结果
	Alternatives   []string `json:"alternatives,omitempty"` // 备选翻译结果列表

	// ========== 服务元信息 ==========
	Service string `json:"service"`         // 翻译平台 (deeplx/google/baidu/openai)
	Model   string `json:"model,omitempty"` // 使用的模型 (可选，如 gpt-4)

	// ========== 缓存元信息 ==========
	CachedAt int64 `json:"cached_at"` // 写入时间戳 (Unix 毫秒)
	Version  int   `json:"version"`   // 缓存格式版本
}

// CacheFormatVersion 当前缓存格式版本
// 升级缓存格式时递增此值
const CacheFormatVersion = 1
