// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/auroralzdf/biddingAI.

package log

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	ctxkey "go-framework/core/context"
)

// Logger 定义了 biddingAI 项目的日志接口. 该接口只包含了支持的日志记录方法.
type Logger interface {
	Debug(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Warning(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
	Panic(msg string, keysAndValues ...interface{})
	Fatal(msg string, keysAndValues ...interface{})
	C(ctx context.Context) *zapLogger
	Sync()
}

// zapLogger 是 Logger 接口的具体实现. 它底层封装了 zap.Logger.
type zapLogger struct {
	z *zap.Logger
}

// 确保 zapLogger 实现了 Logger 接口. 以下变量赋值，可以使错误在编译期被发现.
var _ Logger = &zapLogger{}

var (
	mu sync.Mutex

	// std 定义了默认的全局 Logger.
	std = NewLogger(NewOptions())
)

// Init 使用指定的选项初始化 Logger.
func Init(opts *Options) {
	mu.Lock()
	defer mu.Unlock()

	std = NewLogger(opts)
}

// NewLogger 根据传入的 opts 创建 Logger.
func NewLogger(opts *Options) *zapLogger {
	if opts == nil {
		opts = NewOptions()
	}

	// 将文本格式的日志级别，例如 info 转换为 zapcore.Level 类型以供后面使用
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(opts.Level)); err != nil {
		// 如果指定了非法的日志级别，则默认使用 info 级别
		zapLevel = zapcore.InfoLevel
	}

	// 创建一个默认的 encoder 配置
	encoderConfig := zap.NewProductionEncoderConfig()
	// 自定义 MessageKey 为 message，message 语义更明确
	encoderConfig.MessageKey = "message"
	// 自定义 TimeKey 为 timestamp，timestamp 语义更明确
	encoderConfig.TimeKey = "timestamp"
	// 在日志文件中使用大写字母记录日志级别
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	// 配置控制台编码器使用的字段分隔符
	encoderConfig.ConsoleSeparator = "    "
	// 指定时间序列化函数，将时间序列化为 `2006-01-02 15:04:05.000` 格式，更易读
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
	}
	// 指定 time.Duration 序列化函数，将 time.Duration 序列化为经过的毫秒数的浮点数
	// 毫秒数比默认的秒数更精确
	encoderConfig.EncodeDuration = func(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendFloat64(float64(d) / float64(time.Millisecond))
	}

	// 创建构建 zap.Logger 需要的配置
	cfg := &zap.Config{
		// 是否在日志中显示调用日志所在的文件和行号，例如：`"caller":"biddingAI/biddingAI.go:75"`
		DisableCaller: opts.DisableCaller,
		// 是否禁止在 panic 及以上级别打印堆栈信息
		DisableStacktrace: opts.DisableStacktrace,
		// 指定日志级别
		Level: zap.NewAtomicLevelAt(zapLevel),
		// 指定日志显示格式，可选值：console, json
		Encoding:      opts.Format,
		EncoderConfig: encoderConfig,
		// 指定日志输出位置
		OutputPaths: opts.OutputPaths,
		// 设置 zap 内部错误输出位置
		ErrorOutputPaths: []string{"stderr"},
	}

	// 使用 cfg 创建 *zap.Logger 对象
	z, err := cfg.Build(zap.AddStacktrace(zapcore.PanicLevel), zap.AddCallerSkip(1))
	if err != nil {
		panic(err)
	}
	logger := &zapLogger{z: z}

	// 把标准库的 log.Logger 的 info 级别的输出重定向到 zap.Logger
	zap.RedirectStdLog(z)

	return logger
}

// Sync 调用底层 zap.Logger 的 Sync 方法，将缓存中的日志刷新到磁盘文件中. 主程序需要在退出前调用 Sync.
func Sync() { std.Sync() }

func (l *zapLogger) Sync() {
	_ = l.z.Sync()
}

// Debug 输出 debug 级别的日志.
func Debug(msg string, keysAndValues ...interface{}) {
	std.z.Sugar().Debugw(msg, keysAndValues...)
}

func (l *zapLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.z.Sugar().Debugw(msg, keysAndValues...)
}

// Info 输出 info 级别的日志.
func Info(msg string, keysAndValues ...interface{}) {
	std.z.Sugar().Infow(msg, keysAndValues...)
}

func (l *zapLogger) Info(msg string, keysAndValues ...interface{}) {
	l.z.Sugar().Infow(msg, keysAndValues...)
}

// Warning 输出 warning 级别的日志.
func Warning(msg string, keysAndValues ...interface{}) {
	std.z.Sugar().Infow(msg, keysAndValues...)
}

func (l *zapLogger) Warning(msg string, keysAndValues ...interface{}) {
	l.z.Sugar().Warnw(msg, keysAndValues...)
}

// Error 输出 error 级别的日志.
func Error(msg string, keysAndValues ...interface{}) {
	std.z.Sugar().Errorw(msg, keysAndValues...)
}

func (l *zapLogger) Error(msg string, keysAndValues ...interface{}) {
	l.z.Sugar().Errorw(msg, keysAndValues...)
}

// Panic 输出 panic 级别的日志.
func Panic(msg string, keysAndValues ...interface{}) {
	std.z.Sugar().Panicw(msg, keysAndValues...)
}

func (l *zapLogger) Panic(msg string, keysAndValues ...interface{}) {
	l.z.Sugar().Panicw(msg, keysAndValues...)
}

// Fatal 输出 fatal 级别的日志.
func Fatal(msg string, keysAndValues ...interface{}) {
	std.z.Sugar().Fatalw(msg, keysAndValues...)
}

func (l *zapLogger) Fatal(msg string, keysAndValues ...interface{}) {
	l.z.Sugar().Fatalw(msg, keysAndValues...)
}

func (l *zapLogger) With(fields ...zap.Field) *zapLogger {
	lc := l.clone()
	lc.z = lc.z.With(fields...)
	return lc
}

// C 解析传入的 context，尝试提取关注的键值，并添加到 zap.Logger 结构化日志中.
func C(ctx context.Context) *zapLogger {
	return std.C(ctx)
}

func (l *zapLogger) C(ctx context.Context) *zapLogger {
	lc := l.clone()

	if requestID := ctx.Value(ctxkey.XRequestIDKey); requestID != nil {
		lc.z = lc.z.With(zap.Any(ctxkey.XRequestIDKey, requestID))
	}

	return lc
}

// clone 深度拷贝 zapLogger.
func (l *zapLogger) clone() *zapLogger {
	lc := *l
	return &lc
}
