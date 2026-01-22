package log

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/my-chat/common/pkg/config"
	"github.com/natefinch/lumberjack"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// InitLog 初始化日志
func InitLog(cfg config.LoggerConfiguration) {
	// 创建日志目录
	if cfg.Filename != "" {
		dir := filepath.Dir(cfg.Filename)
		if err := os.MkdirAll(dir, 0755); err != nil {
			panic(err)
		}
	}

	// 设置日志级别
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.TimeFieldFormat = time.RFC3339

	// 控制台输出
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05",
	}

	var writers []io.Writer
	writers = append(writers, consoleWriter)

	// 文件输出
	if cfg.Filename != "" {
		fileWriter := &lumberjack.Logger{
			Filename:   cfg.Filename,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}
		writers = append(writers, fileWriter)
	}

	multi := io.MultiWriter(writers...)
	log.Logger = zerolog.New(multi).With().Timestamp().Caller().Logger()
}

// Debug 调试日志
func Debug() *zerolog.Event {
	return log.Debug()
}

// Info 信息日志
func Info() *zerolog.Event {
	return log.Info()
}

// Warn 警告日志
func Warn() *zerolog.Event {
	return log.Warn()
}

// Error 错误日志
func Error() *zerolog.Event {
	return log.Error()
}

// Fatal 致命日志
func Fatal() *zerolog.Event {
	return log.Fatal()
}
