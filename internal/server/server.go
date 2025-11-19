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

	"translate-services/internal/config"
	"translate-services/internal/translation"
	"translate-services/internal/translator/deeplx"
)

// Server 服务器结构 (封装翻译服务喵～)
type Server struct {
	echo               *echo.Echo
	translationService deeplx.TranslationService
	config             *config.Config
	logger             *zerolog.Logger
	startedAt          time.Time
}

type Dependencies struct {
	TranslationService deeplx.TranslationService
}

type translateRequest struct {
	Q  string   `json:"q"`
	SL string   `json:"sl"`
	TL string   `json:"tl"`
	DT []string `json:"dt"`
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

	e := echo.New()

	s := &Server{
		echo:               e,
		translationService: service,
		config:             cfg,
		logger:             logger,
		startedAt:          time.Now(),
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
	return s.echo.Shutdown(ctx)
}

// translateHandler 处理翻译请求，参数: Echo 上下文，返回: 处理结果的错误
func (s *Server) translateHandler(c echo.Context) error {
	clientIP := c.RealIP()
	payload, err := s.decodeTranslateRequest(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":  "invalid request payload",
			"detail": err.Error(),
		})
	}

	// 获取必需参数
	q := payload.Q
	if strings.TrimSpace(q) == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "missing required parameter: q",
		})
	}

	sl := payload.SL
	tl := payload.TL
	dt := payload.DT

	if strings.TrimSpace(tl) == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "missing required parameter: tl",
		})
	}

	if len(dt) == 0 {
		// 默认只返回翻译文本
		dt = []string{"t"}
	}

	// 调试日志：记录请求参数
	s.logger.Debug().
		Str("handler", "translate_single").
		Str("ip", clientIP).
		Str("sl", sl).
		Str("tl", tl).
		Int("dt_count", len(dt)).
		Msg("收到翻译请求")

	// 调用真实的翻译服务 (浮浮酱的核心改进喵～)，为外部调用增加超时
	ctx, cancel := context.WithTimeout(c.Request().Context(), 8*time.Second)
	defer cancel()

	resp, err := s.translationService.Translate(ctx, q, sl, tl, dt)
	if err != nil {
		s.logger.Warn().
			Err(err).
			Str("handler", "translate_single").
			Str("ip", clientIP).
			Msg("翻译失败，返回上游错误")
		return c.JSON(http.StatusBadGateway, map[string]interface{}{
			"error":  "translation service unavailable",
			"detail": err.Error(),
		})
	}

	if resp == nil {
		s.logger.Error().
			Str("handler", "translate_single").
			Str("ip", clientIP).
			Msg("翻译返回为空")
		return c.JSON(http.StatusBadGateway, map[string]interface{}{
			"error":  "translation service unavailable",
			"detail": "empty response from translation provider",
		})
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
	requiredQueryParams := []string{"client", "sl", "tl", "format", "tk"}
	var missing []string
	for _, key := range requiredQueryParams {
		if strings.TrimSpace(c.QueryParam(key)) == "" {
			missing = append(missing, key)
		}
	}
	if format := c.QueryParam("format"); strings.ToLower(format) != "html" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":  "unsupported format",
			"format": format,
		})
	}
	q := c.FormValue("q")
	if strings.TrimSpace(q) == "" {
		missing = append(missing, "q")
	}
	if len(missing) > 0 {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":          "missing required parameters",
			"missing_fields": missing,
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
		Timeout: 12 * time.Second,
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
