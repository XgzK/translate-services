package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/XgzK/translate-services/internal/translation"
	"github.com/XgzK/translate-services/internal/translator/deeplx"
	"github.com/rs/zerolog"
)

// 缓存写入超时常量
const (
	defaultCacheWriteTimeout = 5 * time.Second // 缓存写入默认超时时间
)

// CachedServiceConfig 缓存服务配置
type CachedServiceConfig struct {
	TTL                 time.Duration // 缓存过期时间，0 表示永不过期
	Enabled             bool          // 是否启用缓存
	ShareAcrossServices bool          // 不同服务共享缓存
	WriteTimeout        time.Duration // 缓存写入超时时间（可选）
}

// CachedTranslationService 包装 TranslationService 添加缓存功能
type CachedTranslationService struct {
	service      deeplx.TranslationService // 被包装的翻译服务
	cache        Cache                     // 缓存实现
	keyGenerator *KeyGenerator             // 缓存键生成器
	ttl          time.Duration             // 缓存过期时间
	enabled      bool                      // 是否启用缓存
	writeTimeout time.Duration             // 缓存写入超时时间
	logger       *zerolog.Logger           // 日志器 (修复: 注入 Logger，保持一致性喵～)
}

// CachedServiceOption 缓存服务可选配置函数类型
type CachedServiceOption func(*CachedTranslationService)

// WithLogger 设置日志器，参数: zerolog.Logger 指针，返回: 配置函数
func WithLogger(logger *zerolog.Logger) CachedServiceOption {
	return func(c *CachedTranslationService) {
		c.logger = logger
	}
}

// WithWriteTimeout 设置缓存写入超时，参数: 超时时间，返回: 配置函数
func WithWriteTimeout(timeout time.Duration) CachedServiceOption {
	return func(c *CachedTranslationService) {
		c.writeTimeout = timeout
	}
}

// NewCachedTranslationService 创建缓存翻译服务
func NewCachedTranslationService(
	service deeplx.TranslationService,
	cache Cache,
	cfg CachedServiceConfig,
	opts ...CachedServiceOption,
) *CachedTranslationService {
	// 设置默认写入超时
	writeTimeout := cfg.WriteTimeout
	if writeTimeout <= 0 {
		writeTimeout = defaultCacheWriteTimeout
	}

	c := &CachedTranslationService{
		service:      service,
		cache:        cache,
		keyGenerator: NewKeyGenerator(cfg.ShareAcrossServices),
		ttl:          cfg.TTL,
		enabled:      cfg.Enabled,
		writeTimeout: writeTimeout,
	}

	// 应用可选配置
	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Translate 实现 TranslationService 接口
func (c *CachedTranslationService) Translate(
	ctx context.Context,
	q, sl, tl string,
	dt []string,
) (*translation.Response, error) {
	return c.TranslateWithModel(ctx, q, sl, tl, dt, "")
}

// TranslateWithModel 实现 TranslationService 接口
func (c *CachedTranslationService) TranslateWithModel(
	ctx context.Context,
	q, sl, tl string,
	dt []string,
	model string,
) (*translation.Response, error) {
	// 缓存未启用或缓存实例为空，直接调用底层服务
	if !c.enabled || c.cache == nil {
		return c.service.TranslateWithModel(ctx, q, sl, tl, dt, model)
	}

	// 生成缓存键
	serviceName := c.service.GetName()
	key := c.keyGenerator.Generate(serviceName, q, sl, tl, model)

	// 尝试从缓存获取
	if cached, err := c.getFromCache(ctx, key); err == nil && cached != nil {
		c.logDebug().
			Str("key", key).
			Str("service", serviceName).
			Msg("cache hit")
		return c.buildResponseFromCache(cached), nil
	}

	// 缓存未命中，调用翻译服务
	c.logDebug().
		Str("key", key).
		Str("service", serviceName).
		Msg("cache miss, calling translation service")

	resp, err := c.service.TranslateWithModel(ctx, q, sl, tl, dt, model)
	if err != nil {
		return nil, err
	}

	// 异步写入缓存（带超时控制，不阻塞响应喵～）
	go c.saveToCacheWithTimeout(key, q, sl, tl, model, resp)

	return resp, nil
}

// GetName 返回服务名称
func (c *CachedTranslationService) GetName() string {
	return "cached-" + c.service.GetName()
}

// IsAvailable 检查服务是否可用
func (c *CachedTranslationService) IsAvailable() bool {
	return c.service.IsAvailable()
}

// getFromCache 从缓存获取翻译结果
func (c *CachedTranslationService) getFromCache(ctx context.Context, key string) (*CachedTranslation, error) {
	data, err := c.cache.Get(ctx, key)
	if err != nil {
		c.logWarn().Err(err).Str("key", key).Msg("cache get failed")
		return nil, err
	}
	if data == nil {
		return nil, nil // 缓存未命中
	}

	var cached CachedTranslation
	if err := json.Unmarshal(data, &cached); err != nil {
		c.logWarn().Err(err).Str("key", key).Msg("cache unmarshal failed, ignoring corrupted data")
		return nil, err
	}

	// 检查缓存版本兼容性
	if cached.Version != CacheFormatVersion {
		c.logDebug().
			Int("cached_version", cached.Version).
			Int("current_version", CacheFormatVersion).
			Msg("cache version mismatch, ignoring old data")
		return nil, nil
	}

	return &cached, nil
}

// saveToCacheWithTimeout 带超时控制的缓存保存 (修复: 添加超时控制喵～)
func (c *CachedTranslationService) saveToCacheWithTimeout(
	key, originalText, sourceLang, targetLang, model string,
	resp *translation.Response,
) {
	// 创建带超时的 context
	ctx, cancel := context.WithTimeout(context.Background(), c.writeTimeout)
	defer cancel()

	c.saveToCache(ctx, key, originalText, sourceLang, targetLang, model, resp)
}

// saveToCache 保存翻译结果到缓存
func (c *CachedTranslationService) saveToCache(
	ctx context.Context,
	key, originalText, sourceLang, targetLang, model string,
	resp *translation.Response,
) {
	cached := c.buildCachedTranslation(originalText, sourceLang, targetLang, model, resp)

	data, err := json.Marshal(cached)
	if err != nil {
		c.logWarn().Err(err).Str("key", key).Msg("cache marshal failed")
		return
	}

	if err := c.cache.Set(ctx, key, data, c.ttl); err != nil {
		// 检查是否为超时错误
		if ctx.Err() == context.DeadlineExceeded {
			c.logWarn().Str("key", key).Dur("timeout", c.writeTimeout).Msg("cache write timeout")
		} else {
			c.logWarn().Err(err).Str("key", key).Msg("cache set failed")
		}
		return
	}

	c.logDebug().
		Str("key", key).
		Str("service", c.service.GetName()).
		Dur("ttl", c.ttl).
		Msg("cache saved")
}

// buildCachedTranslation 从 Response 构建缓存结构
func (c *CachedTranslationService) buildCachedTranslation(
	originalText, sourceLang, targetLang, model string,
	resp *translation.Response,
) *CachedTranslation {
	cached := &CachedTranslation{
		OriginalText: originalText,
		SourceLang:   resp.Src, // 使用实际检测的源语言
		TargetLang:   targetLang,
		Service:      c.service.GetName(),
		Model:        model,
		CachedAt:     time.Now().UnixMilli(),
		Version:      CacheFormatVersion,
	}

	// 如果 Src 为空，使用请求的源语言
	if cached.SourceLang == "" {
		cached.SourceLang = sourceLang
	}

	// 提取主翻译结果
	if len(resp.Sentences) > 0 {
		var translatedText string
		for _, sentence := range resp.Sentences {
			translatedText += sentence.Trans
		}
		cached.TranslatedText = translatedText
	}

	// 提取备选翻译
	if len(resp.AlternativeTranslations) > 0 {
		alternatives := make([]string, 0)
		for _, alt := range resp.AlternativeTranslations {
			for _, a := range alt.Alternative {
				if a.WordPostproc != "" && a.WordPostproc != cached.TranslatedText {
					alternatives = append(alternatives, a.WordPostproc)
				}
			}
		}
		if len(alternatives) > 0 {
			cached.Alternatives = alternatives
		}
	}

	return cached
}

// buildResponseFromCache 从缓存构建 Response
func (c *CachedTranslationService) buildResponseFromCache(cached *CachedTranslation) *translation.Response {
	resp := &translation.Response{
		Src: cached.SourceLang,
		Sentences: []translation.Sentence{
			{
				Orig:  cached.OriginalText,
				Trans: cached.TranslatedText,
			},
		},
	}

	// 如果有备选翻译，构建 AlternativeTranslations
	if len(cached.Alternatives) > 0 {
		alternatives := make([]translation.Alternative, 0, len(cached.Alternatives))
		for _, alt := range cached.Alternatives {
			alternatives = append(alternatives, translation.Alternative{
				WordPostproc: alt,
			})
		}
		resp.AlternativeTranslations = []translation.AlternativeTranslation{
			{
				SrcPhrase:   cached.OriginalText,
				Alternative: alternatives,
			},
		}
	}

	return resp
}

// Close 关闭缓存连接
func (c *CachedTranslationService) Close() error {
	if c.cache != nil {
		return c.cache.Close()
	}
	return nil
}

// ========== 日志辅助方法 (统一日志处理喵～) ==========

// nopLogger 空日志器（用于未注入 logger 时）
var nopLogger = zerolog.Nop()

// logDebug 返回 Debug 级别日志事件
func (c *CachedTranslationService) logDebug() *zerolog.Event {
	if c.logger != nil {
		return c.logger.Debug()
	}
	return nopLogger.Debug()
}

// logWarn 返回 Warn 级别日志事件
func (c *CachedTranslationService) logWarn() *zerolog.Event {
	if c.logger != nil {
		return c.logger.Warn()
	}
	return nopLogger.Warn()
}
