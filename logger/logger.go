package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

var (
	// Logger 全局日志实例
	Logger zerolog.Logger
)

// Init 初始化日志系统
// level: trace, debug, info, warn, error, fatal, panic
// pretty: 是否使用美化输出（开发环境推荐 true，生产环境推荐 false）
func Init(level string, pretty bool) {
	// 设置全局日志级别
	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(logLevel)

	// 设置时间格式
	zerolog.TimeFieldFormat = time.RFC3339

	// 设置输出
	if pretty {
		// 开发环境：使用控制台美化输出
		output := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "15:04:05",
			NoColor:    false,
		}
		Logger = zerolog.New(output).With().Timestamp().Logger()
	} else {
		// 生产环境：使用 JSON 格式输出
		Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	}
}

// Trace 记录 Trace 级别日志
func Trace(msg string) {
	Logger.Trace().Msg(msg)
}

// Tracef 记录带格式的 Trace 级别日志
func Tracef(format string, v ...interface{}) {
	Logger.Trace().Msgf(format, v...)
}

// Debug 记录 Debug 级别日志
func Debug(msg string) {
	Logger.Debug().Msg(msg)
}

// Debugf 记录带格式的 Debug 级别日志
func Debugf(format string, v ...interface{}) {
	Logger.Debug().Msgf(format, v...)
}

// Info 记录 Info 级别日志
func Info(msg string) {
	Logger.Info().Msg(msg)
}

// Infof 记录带格式的 Info 级别日志
func Infof(format string, v ...interface{}) {
	Logger.Info().Msgf(format, v...)
}

// Warn 记录 Warn 级别日志
func Warn(msg string) {
	Logger.Warn().Msg(msg)
}

// Warnf 记录带格式的 Warn 级别日志
func Warnf(format string, v ...interface{}) {
	Logger.Warn().Msgf(format, v...)
}

// Error 记录 Error 级别日志
func Error(msg string) {
	Logger.Error().Msg(msg)
}

// Errorf 记录带格式的 Error 级别日志
func Errorf(format string, v ...interface{}) {
	Logger.Error().Msgf(format, v...)
}

// Err 记录带错误的 Error 级别日志
func Err(err error) {
	Logger.Error().Err(err).Send()
}

// Fatal 记录 Fatal 级别日志并退出程序
func Fatal(msg string) {
	Logger.Fatal().Msg(msg)
}

// Fatalf 记录带格式的 Fatal 级别日志并退出程序
func Fatalf(format string, v ...interface{}) {
	Logger.Fatal().Msgf(format, v...)
}

// WithField 添加字段到日志上下文
func WithField(key string, value interface{}) zerolog.Context {
	return Logger.With().Interface(key, value)
}

// WithFields 添加多个字段到日志上下文
func WithFields(fields map[string]interface{}) zerolog.Context {
	ctx := Logger.With()
	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}
	return ctx
}

// WithError 添加错误到日志上下文
func WithError(err error) zerolog.Context {
	return Logger.With().Err(err)
}
