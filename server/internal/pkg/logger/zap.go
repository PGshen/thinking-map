package logger

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	// Log 全局日志实例
	Log *zap.Logger
)

// Config 日志配置
type Config struct {
	Level      string `yaml:"level"`      // 日志级别
	Filename   string `yaml:"filename"`   // 日志文件路径
	MaxSize    int    `yaml:"maxSize"`    // 单个文件最大尺寸，单位是MB
	MaxBackups int    `yaml:"maxBackups"` // 最大保留文件数
	MaxAge     int    `yaml:"maxAge"`     // 最大保留天数
	Compress   bool   `yaml:"compress"`   // 是否压缩
	Console    bool   `yaml:"console"`    // 是否输出到控制台
}

// Init 初始化日志
func Init(cfg *Config) error {
	// 设置日志级别
	level := zap.InfoLevel
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		return err
	}

	// 创建编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     timeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 创建输出
	var cores []zapcore.Core

	// 控制台输出
	if cfg.Console {
		consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
		consoleCore := zapcore.NewCore(
			consoleEncoder,
			zapcore.AddSync(os.Stdout),
			level,
		)
		cores = append(cores, consoleCore)
	}

	// 文件输出
	if cfg.Filename != "" {
		// 创建日志目录
		if err := os.MkdirAll(cfg.Filename, 0755); err != nil {
			return err
		}

		// 创建文件输出
		fileEncoder := zapcore.NewJSONEncoder(encoderConfig)
		fileCore := zapcore.NewCore(
			fileEncoder,
			zapcore.AddSync(&lumberjack.Logger{
				Filename:   cfg.Filename,
				MaxSize:    cfg.MaxSize,
				MaxBackups: cfg.MaxBackups,
				MaxAge:     cfg.MaxAge,
				Compress:   cfg.Compress,
			}),
			level,
		)
		cores = append(cores, fileCore)
	}

	// 创建日志实例
	core := zapcore.NewTee(cores...)
	Log = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return nil
}

// timeEncoder 自定义时间编码器
func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

// Debug 输出Debug级别日志
func Debug(msg string, fields ...zap.Field) {
	Log.Debug(msg, fields...)
}

// Info 输出Info级别日志
func Info(msg string, fields ...zap.Field) {
	Log.Info(msg, fields...)
}

// Warn 输出Warn级别日志
func Warn(msg string, fields ...zap.Field) {
	Log.Warn(msg, fields...)
}

// Error 输出Error级别日志
func Error(msg string, fields ...zap.Field) {
	Log.Error(msg, fields...)
}

// Fatal 输出Fatal级别日志
func Fatal(msg string, fields ...zap.Field) {
	Log.Fatal(msg, fields...)
}

// With 创建子日志实例
func With(fields ...zap.Field) *zap.Logger {
	return Log.With(fields...)
}

// Sync 同步日志
func Sync() error {
	return Log.Sync()
}
