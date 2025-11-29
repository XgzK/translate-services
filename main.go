package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/XgzK/translate-services/internal/config"
	"github.com/XgzK/translate-services/internal/logging"
	"github.com/XgzK/translate-services/internal/server"
)

// main 是服务的入口函数，参数: 无，返回: 无
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "配置验证失败: %v\n", err)
		os.Exit(1)
	}

	logger := logging.New(logging.Options{
		Debug:   cfg.Debug,
		Service: "deeplx-server",
	})

	logger.Info().
		Str("port", cfg.Port).
		Bool("debug", cfg.Debug).
		Str("service_type", cfg.Translation.ServiceType).
		Bool("has_api_key", cfg.Translation.APIKey != "").
		Bool("custom_base_url", cfg.Translation.BaseURL != "").
		Msg("配置加载成功")

	srv, err := server.New(cfg, logger, nil)
	if err != nil {
		logger.Fatal().Err(err).Msg("创建服务器失败")
	}

	addr := fmt.Sprintf(":%s", cfg.Port)
	logger.Info().Str("address", addr).Msg("服务启动中")

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- srv.Start(addr)
	}()

	select {
	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal().Err(err).Msg("服务器运行失败")
		}
	case <-ctx.Done():
		logger.Info().Msg("收到停止信号，准备优雅停机")
		shutdownTimeout := time.Duration(cfg.Server.GetShutdownTimeout()) * time.Second
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error().Err(err).Msg("优雅停机失败")
		}
		err := <-serverErr
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error().Err(err).Msg("服务器关闭时出现错误")
		} else {
			logger.Info().Msg("服务器已停机")
		}
	}
}
