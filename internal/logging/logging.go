package logging

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// Options 定义日志器配置，参数: 无，返回: 无
type Options struct {
	Debug   bool
	Service string
	Writer  io.Writer
}

// New 创建带有统一字段的结构化日志器，参数: Options 配置，返回: 初始化好的 zerolog.Logger 指针
func New(opts Options) *zerolog.Logger {
	writer := opts.Writer
	if writer == nil {
		writer = os.Stdout
	}

	consoleWriter := zerolog.ConsoleWriter{
		Out:        writer,
		TimeFormat: time.RFC3339,
		FormatLevel: func(i interface{}) string {
			if level, ok := i.(string); ok {
				return strings.ToUpper(level)
			}
			return "INFO"
		},
	}

	level := zerolog.InfoLevel
	if opts.Debug {
		level = zerolog.DebugLevel
	}

	contextBuilder := zerolog.New(consoleWriter).With().Timestamp()
	if opts.Service != "" {
		contextBuilder = contextBuilder.Str("service", opts.Service)
	}
	contextBuilder = contextBuilder.Time("started_at", time.Now())

	logger := contextBuilder.Logger().Level(level)
	return &logger
}
