package server

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"

	"github.com/XgzK/translate-services/internal/cache"
	"github.com/XgzK/translate-services/internal/config"
	"github.com/XgzK/translate-services/internal/translation"
	"github.com/XgzK/translate-services/internal/translator/deeplx"
)

// Server 服务器结构 (封装翻译服务喵～)
type Server struct {
	echo               *echo.Echo
	translationService deeplx.TranslationService
	config             *config.Config
	logger             *zerolog.Logger
	startedAt          time.Time
	cache              cache.Cache // 可选的缓存实例
}

type Dependencies struct {
	TranslationService deeplx.TranslationService
}

type translateRequest struct {
	Q     string   `json:"q"`
	SL    string   `json:"sl"`
	TL    string   `json:"tl"`
	DT    []string `json:"dt"`
	Model string   `json:"model,omitempty"` // 可选：指定翻译模型
}

// New 构建服务器，参数: 配置、日志器、依赖注入，返回: 初始化好的 Server 或错误
func New(cfg *config.Config, logger *zerolog.Logger, deps *Dependencies) (*Server, error) {
	if cfg == nil {
		return nil, errors.New("config 不能为空")
	}

	if logger == nil {
		nop := zerolog.New(io.Discard)
		logger = &nop
	}

	service, err := selectTranslationService(cfg, deps)
	if err != nil {
		return nil, err
	}

	if !service.IsAvailable() {
		logger.Warn().Msg("翻译服务不可用，请检查 API 密钥")
	} else {
		logger.Info().Str("provider", service.GetName()).Msg("翻译服务初始化完成")
	}

	// 初始化缓存（如果启用）
	var cacheInstance cache.Cache
	if cfg.Cache.Enabled {
		redisCache, err := cache.NewRedisCache(cache.RedisConfig{
			Addr:         cfg.Cache.Addr,
			Password:     cfg.Cache.Password,
			DB:           cfg.Cache.DB,
			PoolSize:     cfg.Cache.GetPoolSize(),
			DialTimeout:  cfg.Cache.GetDialTimeout(),
			ReadTimeout:  cfg.Cache.GetReadTimeout(),
			WriteTimeout: cfg.Cache.GetWriteTimeout(),
		})
		if err != nil {
			// 缓存连接失败，记录警告但继续运行（优雅降级）
			logger.Warn().Err(err).Msg("Redis 缓存连接失败，将以无缓存模式运行")
		} else {
			cacheInstance = redisCache
			logger.Info().
				Str("addr", cfg.Cache.Addr).
				Dur("ttl", cfg.Cache.GetTTL()).
				Bool("share_across_services", cfg.Cache.ShareAcrossServices).
				Msg("Redis 缓存初始化完成")

			// 包装翻译服务，添加缓存功能 (修复: 传入 logger 保持日志一致性喵～)
			service = cache.NewCachedTranslationService(service, cacheInstance, cache.CachedServiceConfig{
				TTL:                 cfg.Cache.GetTTL(),
				Enabled:             true,
				ShareAcrossServices: cfg.Cache.ShareAcrossServices,
			}, cache.WithLogger(logger))
			logger.Info().Str("provider", service.GetName()).Msg("翻译服务已启用缓存")
		}
	}

	e := echo.New()

	s := &Server{
		echo:               e,
		translationService: service,
		config:             cfg,
		logger:             logger,
		startedAt:          time.Now(),
		cache:              cacheInstance,
	}

	s.configureMiddleware()
	s.registerRoutes()

	return s, nil
}

// selectTranslationService 选择翻译服务，参数: 配置和测试依赖，返回: 翻译服务实例或错误
func selectTranslationService(cfg *config.Config, deps *Dependencies) (deeplx.TranslationService, error) {
	if deps != nil && deps.TranslationService != nil {
		return deps.TranslationService, nil
	}

	factory := deeplx.NewFactory()
	serviceType := cfg.Translation.ServiceType
	if strings.TrimSpace(serviceType) == "" {
		serviceType = string(deeplx.ServiceTypeDeepLX)
	}
	service, err := factory.CreateService(
		deeplx.ServiceType(strings.ToLower(serviceType)),
		&deeplx.TranslationServiceConfig{
			APIKey:  cfg.Translation.APIKey,
			BaseURL: cfg.Translation.BaseURL,
		},
	)
	if err != nil {
		return nil, err
	}
	return service, nil
}

// Start 启动服务器，参数: 监听地址字符串，返回: 启动失败的错误
func (s *Server) Start(addr string) error {
	return s.echo.Start(addr)
}

// Shutdown 优雅关闭服务器，参数: 上下文，用于超时控制，返回: 关闭时的错误
func (s *Server) Shutdown(ctx context.Context) error {
	// 关闭缓存连接
	if s.cache != nil {
		if err := s.cache.Close(); err != nil {
			s.logger.Warn().Err(err).Msg("关闭缓存连接失败")
		} else {
			s.logger.Info().Msg("缓存连接已关闭")
		}
	}
	return s.echo.Shutdown(ctx)
}

// translateHandler 处理翻译请求，参数: Echo 上下文，返回: 处理结果的错误
func (s *Server) translateHandler(c echo.Context) error {
	clientIP := c.RealIP()
	payload, err := s.decodeTranslateRequest(c)
	if err != nil {
		return BadRequestWithDetails(c, ErrCodeInvalidRequest, "invalid request payload", err.Error())
	}

	// 获取必需参数
	q := payload.Q
	if strings.TrimSpace(q) == "" {
		return BadRequest(c, ErrCodeMissingParameter, "missing required parameter: q")
	}

	sl := payload.SL
	tl := payload.TL
	dt := payload.DT
	model := payload.Model

	// 如果请求中没有指定模型，使用配置文件中的默认模型
	if model == "" && s.config.Translation.Model != "" {
		model = s.config.Translation.Model
	}

	if strings.TrimSpace(tl) == "" {
		return BadRequest(c, ErrCodeMissingParameter, "missing required parameter: tl")
	}

	if len(dt) == 0 {
		// 默认只返回翻译文本
		dt = []string{"t"}
	}

	// 调试日志：记录请求参数
	logEvent := s.logger.Debug().
		Str("handler", "translate_single").
		Str("ip", clientIP).
		Str("sl", sl).
		Str("tl", tl).
		Int("dt_count", len(dt))

	if model != "" {
		logEvent.Str("model", model)
	}
	logEvent.Msg("收到翻译请求")

	// 调用真实的翻译服务 (浮浮酱的核心改进喵～)，为外部调用增加超时
	requestTimeout := time.Duration(s.config.Server.GetRequestTimeout()) * time.Second
	ctx, cancel := context.WithTimeout(c.Request().Context(), requestTimeout)
	defer cancel()

	var resp *translation.Response

	// 根据是否指定模型选择不同的翻译方法
	if model != "" {
		resp, err = s.translationService.TranslateWithModel(ctx, q, sl, tl, dt, model)
	} else {
		resp, err = s.translationService.Translate(ctx, q, sl, tl, dt)
	}

	if err != nil {
		s.logger.Warn().
			Err(err).
			Str("handler", "translate_single").
			Str("ip", clientIP).
			Msg("翻译失败，返回上游错误")
		return BadGatewayWithDetails(c, ErrCodeTranslationFailed, "translation service unavailable", err.Error())
	}

	if resp == nil {
		s.logger.Error().
			Str("handler", "translate_single").
			Str("ip", clientIP).
			Msg("翻译返回为空")
		return BadGatewayWithDetails(c, ErrCodeServiceUnavailable, "translation service unavailable", "empty response from translation provider")
	}

	// 请求成功日志（保持在 Info，默认可见）
	if len(resp.Sentences) > 0 {
		s.logger.Info().
			Str("handler", "translate_single").
			Str("ip", clientIP).
			Str("requested_sl", sl).
			Str("requested_tl", tl).
			Str("detected_src", resp.Src).
			Str("orig", resp.Sentences[0].Orig).
			Str("trans", resp.Sentences[0].Trans).
			Msg("翻译成功")
	}

	return c.JSON(http.StatusOK, resp)
}

// translateDocumentHandler 处理文档翻译请求，参数: Echo 上下文，返回: 处理结果的错误
func (s *Server) translateDocumentHandler(c echo.Context) error {
	// 首先检查必需的查询参数 (修复：先检查缺失参数再检查格式喵～)
	requiredQueryParams := []string{"client", "sl", "tl", "format", "tk"}
	var missing []string
	for _, key := range requiredQueryParams {
		if strings.TrimSpace(c.QueryParam(key)) == "" {
			missing = append(missing, key)
		}
	}

	// 检查必需的表单参数
	q := c.FormValue("q")
	if strings.TrimSpace(q) == "" {
		missing = append(missing, "q")
	}

	// 如果有缺失参数，立即返回错误
	if len(missing) > 0 {
		return BadRequestWithDetails(c, ErrCodeMissingParameter, "missing required parameters", map[string]interface{}{
			"missing_fields": missing,
		})
	}

	// 参数完整后，再验证格式
	format := c.QueryParam("format")
	if strings.ToLower(format) != "html" {
		return BadRequestWithDetails(c, ErrCodeUnsupportedFormat, "unsupported format", map[string]interface{}{
			"format":    format,
			"supported": []string{"html"},
		})
	}

	resp := translation.BuildDocumentResponse(q, c.QueryParam("sl"))
	return c.JSON(http.StatusOK, resp)
}

// elementHandler 返回元素脚本，参数: Echo 上下文，返回: 处理结果的错误
func (s *Server) elementHandler(c echo.Context) error {
	js := translation.ElementScript()
	return c.Blob(http.StatusOK, "text/javascript; charset=utf-8", []byte(js))
}

// healthHandler 健康检查，参数: Echo 上下文，返回: 处理结果的错误
func (s *Server) healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "ok",
		"uptime": time.Since(s.startedAt).Seconds(),
	})
}

// configureMiddleware 配置中间件，参数: 无（使用接收者），返回: 无
func (s *Server) configureMiddleware() {
	s.echo.HideBanner = true
	s.echo.HidePort = true
	s.echo.Use(middleware.Recover())
	s.echo.Use(middleware.RequestID())
	s.echo.Use(middleware.BodyLimit("2M"))
	s.echo.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: time.Duration(s.config.Server.GetMiddlewareTimeout()) * time.Second,
	}))

	s.echo.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:  true,
		LogURI:     true,
		LogMethod:  true,
		LogLatency: true,
		LogError:   true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			var event *zerolog.Event
			switch {
			case v.Error != nil:
				event = s.logger.Error().Err(v.Error)
			case v.Status >= http.StatusInternalServerError:
				event = s.logger.Error()
			case v.Status >= http.StatusBadRequest:
				event = s.logger.Warn()
			default:
				event = s.logger.Debug()
			}
			event = event.
				Str("method", v.Method).
				Str("uri", v.URI).
				Str("ip", c.RealIP()).
				Int("status", v.Status).
				Dur("latency", v.Latency)
			event.Msg("http_request")
			return nil
		},
	}))

	s.echo.Use(echoprometheus.NewMiddleware("deeplx"))
}

// registerRoutes 注册路由，参数: 无（使用接收者），返回: 无
func (s *Server) registerRoutes() {
	s.echo.GET("/translate_a/element.js", s.elementHandler)
	s.echo.POST("/translate_a/single", s.translateHandler)
	s.echo.POST("/translate_a/t", s.translateDocumentHandler)
	s.echo.GET("/healthz", s.healthHandler)
	s.echo.GET("/metrics", echoprometheus.NewHandler())
}

// decodeTranslateRequest 解析翻译请求参数，参数: Echo 上下文，返回: 翻译请求结构与错误
func (s *Server) decodeTranslateRequest(c echo.Context) (translateRequest, error) {
	var payload translateRequest
	contentType := c.Request().Header.Get("Content-Type")

	if strings.Contains(strings.ToLower(contentType), "application/json") {
		if err := c.Bind(&payload); err != nil {
			return payload, err
		}
	} else {
		payload.Q = c.FormValue("q")
		payload.SL = c.FormValue("sl")
		payload.TL = c.FormValue("tl")

		if formValues, err := c.FormParams(); err == nil && len(formValues["dt"]) > 0 {
			payload.DT = append(payload.DT, formValues["dt"]...)
		}
	}

	if payload.SL == "" {
		payload.SL = c.QueryParam("sl")
	}
	if payload.TL == "" {
		payload.TL = c.QueryParam("tl")
	}
	if len(payload.DT) == 0 {
		if queryValues := c.QueryParams()["dt"]; len(queryValues) > 0 {
			payload.DT = append(payload.DT, queryValues...)
		}
	}

	return payload, nil
}
