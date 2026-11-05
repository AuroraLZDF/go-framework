// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package log

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	ctxkey "github.com/AuroraLZDF/go-framework/core/context"
)

// Logger 定义了 AuroraLZDF/go-framework 项目的日志接口. 该接口只包含了支持的日志记录方法.
type Logger interface {
	Debug(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Warning(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
	Panic(msg string, keysAndValues ...interface{})
	Fatal(msg string, keysAndValues ...interface{})
	C(ctx context.Context) Logger
	Sync()
	Close()
}

// zapLogger 是 Logger 接口的具体实现. 它底层封装了 zap.Logger.
type zapLogger struct {
	z *zap.Logger
	// rotate 仅在启用日志轮转时非空，用于退出前关闭文件.
	// 只有 Init/NewLogger 返回的顶层 logger 才持有文件；从 C/With 派生的 logger
	// 不拥有文件，避免关闭单个派生 logger 时误关全局日志文件.
	rotate     *RotateWriter
	ownsRotate bool
}

// 确保 zapLogger 实现了 Logger 接口. 以下变量赋值，可以使错误在编译期被发现.
var _ Logger = &zapLogger{}

var (
	mu sync.Mutex
	// refs 记录 Init 的次数，Close 时递减；仅当归零时才关闭底层轮转文件，
	// 避免一个 Application.Close 关掉另一个 Application 仍在使用中的日志文件.
	refs      int
	rotateRef *RotateWriter

	// stdPtr 持有当前全局 Logger；使用原子指针保证读写并发安全.
	stdPtr atomic.Pointer[zapLogger]
)

// nopLogger 在日志系统关闭后作为兜底，保证关闭后继续调用日志函数不会 panic.
var nopLogger = &zapLogger{z: zap.NewNop()}

func init() {
	logger, _, err := newLogger(NewOptions())
	if err != nil {
		// 默认配置（console + stdout）理论上不会失败，兜底为 nop logger.
		logger = nopLogger
	}
	stdPtr.Store(logger)
}

// current 返回当前全局 Logger，永不为 nil.
func current() *zapLogger {
	if l := stdPtr.Load(); l != nil {
		return l
	}
	return nopLogger
}

// Init 使用指定的选项初始化 Logger.
// 轮转目录/文件打不开等初始化错误会返回给调用方（fail-loud），而不是静默丢失日志.
func Init(opts *Options) error {
	logger, rw, err := newLogger(opts)
	if err != nil {
		return err
	}

	mu.Lock()
	oldRotate := rotateRef
	rotateRef = rw
	refs++
	stdPtr.Store(logger)
	mu.Unlock()

	// 全局 Logger 已被替换，旧写入器不再被使用，立即关闭避免文件句柄泄漏.
	if oldRotate != nil {
		_ = oldRotate.Close()
	}

	return nil
}

// Reset 重置全局日志状态，仅用于测试.
// 会关闭当前轮转文件并把 Logger 置为 nop，之后可以重新 Init.
func Reset() {
	mu.Lock()
	rw := rotateRef
	rotateRef = nil
	refs = 0
	stdPtr.Store(nopLogger)
	mu.Unlock()

	if rw != nil {
		_ = rw.Close()
	}
}

// NewLogger 根据传入的 opts 创建 Logger.
// 为了保持向后兼容，该接口不返回错误：初始化失败时会输出到 stderr 并回退为 stdout logger.
// 需要感知初始化失败（如日志目录不可写）请使用 Init.
// 返回的 Logger 不会注册为全局 logger；如需全局生效请使用 Init.
func NewLogger(opts *Options) Logger {
	logger, _, err := newLogger(opts)
	if err == nil {
		return logger
	}

	fmt.Fprintf(os.Stderr, "failed to build logger: %v; falling back to stdout logger\n", err)

	fallback, _, ferr := newLogger(NewOptions())
	if ferr != nil {
		return nopLogger
	}
	return fallback
}

// newLogger 构建 Logger 与可选的轮转写入器.
func newLogger(opts *Options) (*zapLogger, *RotateWriter, error) {
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

	// 如果有轮转配置，使用自定义写入器
	if opts.Rotation != nil {
		rw, err := NewRotateWriter(opts.Rotation.BasePath, opts.Rotation.Interval)
		if err != nil {
			return nil, nil, err
		}

		core := createRotateCore(opts, encoderConfig, zapLevel, rw)

		zapOpts := []zap.Option{zap.AddCallerSkip(1)}
		if !opts.DisableStacktrace {
			zapOpts = append(zapOpts, zap.AddStacktrace(zapcore.PanicLevel))
		}
		if !opts.DisableCaller {
			zapOpts = append(zapOpts, zap.AddCaller())
		}

		z := zap.New(core, zapOpts...)
		logger := &zapLogger{z: z, rotate: rw, ownsRotate: true}
		// 把标准库的 log.Logger 的 info 级别的输出重定向到 zap.Logger
		zap.RedirectStdLog(z)
		return logger, rw, nil
	}

	// 创建构建 zap.Logger 需要的配置
	cfg := &zap.Config{
		// 是否在日志中显示调用日志所在的文件和行号，例如：`"caller":"go-framework/log.go:75"`
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
	if len(cfg.OutputPaths) == 0 {
		cfg.OutputPaths = []string{"stdout"}
	}

	// 使用 cfg 创建 *zap.Logger 对象
	zapOpts := []zap.Option{zap.AddCallerSkip(1)}
	if !opts.DisableStacktrace {
		zapOpts = append(zapOpts, zap.AddStacktrace(zapcore.PanicLevel))
	}

	z, err := cfg.Build(zapOpts...)
	if err != nil {
		return nil, nil, err
	}
	logger := &zapLogger{z: z}

	// 把标准库的 log.Logger 的 info 级别的输出重定向到 zap.Logger
	zap.RedirectStdLog(z)

	return logger, nil, nil
}

// createRotateCore 创建支持日志轮转的 zapcore.Core
func createRotateCore(opts *Options, encoderConfig zapcore.EncoderConfig, level zapcore.Level, rw *RotateWriter) zapcore.Core {
	var encoder zapcore.Encoder
	if opts.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 创建多输出 core
	cores := []zapcore.Core{
		zapcore.NewCore(encoder, zapcore.AddSync(rw), level),
	}

	if opts.Rotation.Stdout {
		cores = append(cores, zapcore.NewCore(
			encoder,
			zapcore.Lock(os.Stdout),
			level,
		))
	}

	return zapcore.NewTee(cores...)
}

// Sync 调用底层 zap.Logger 的 Sync 方法，将缓存中的日志刷新到磁盘文件中. 主程序需要在退出前调用 Sync.
func Sync() { current().Sync() }

func (l *zapLogger) Sync() {
	_ = l.z.Sync()
}

// Close 关闭日志文件.
// 全局 Close 是引用计数的：每次 Init 计一次，Close 递减，归零时才真正关闭轮转文件.
func Close() error {
	mu.Lock()
	if refs == 0 {
		mu.Unlock()
		return nil
	}
	refs--
	if refs > 0 {
		mu.Unlock()
		return nil
	}

	rw := rotateRef
	rotateRef = nil
	// 关闭后仍可能有日志调用，切换到 nop logger，避免向已关闭的文件写入.
	stdPtr.Store(nopLogger)
	mu.Unlock()

	if rw != nil {
		return rw.Close()
	}
	return nil
}

// Close 关闭派生 logger 持有的文件.
// 只有 Init/NewLogger 创建的顶层 logger 拥有文件；C/With 派生的 logger 关闭时不会影响全局日志.
func (l *zapLogger) Close() {
	if l.ownsRotate && l.rotate != nil {
		_ = l.rotate.Close()
	}
}

// Debug 输出 debug 级别的日志.
func Debug(msg string, keysAndValues ...interface{}) {
	current().Debug(msg, keysAndValues...)
}

func (l *zapLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.z.Sugar().Debugw(msg, keysAndValues...)
}

// Info 输出 info 级别的日志.
func Info(msg string, keysAndValues ...interface{}) {
	current().Info(msg, keysAndValues...)
}

func (l *zapLogger) Info(msg string, keysAndValues ...interface{}) {
	l.z.Sugar().Infow(msg, keysAndValues...)
}

// Warning 输出 warning 级别的日志.
func Warning(msg string, keysAndValues ...interface{}) {
	current().Warning(msg, keysAndValues...)
}

func (l *zapLogger) Warning(msg string, keysAndValues ...interface{}) {
	l.z.Sugar().Warnw(msg, keysAndValues...)
}

// Error 输出 error 级别的日志.
func Error(msg string, keysAndValues ...interface{}) {
	current().Error(msg, keysAndValues...)
}

func (l *zapLogger) Error(msg string, keysAndValues ...interface{}) {
	l.z.Sugar().Errorw(msg, keysAndValues...)
}

// Panic 输出 panic 级别的日志.
func Panic(msg string, keysAndValues ...interface{}) {
	current().Panic(msg, keysAndValues...)
}

func (l *zapLogger) Panic(msg string, keysAndValues ...interface{}) {
	l.z.Sugar().Panicw(msg, keysAndValues...)
}

// Fatal 输出 fatal 级别的日志.
func Fatal(msg string, keysAndValues ...interface{}) {
	current().Fatal(msg, keysAndValues...)
}

func (l *zapLogger) Fatal(msg string, keysAndValues ...interface{}) {
	l.z.Sugar().Fatalw(msg, keysAndValues...)
}

// With 返回携带额外字段的派生 logger.
func (l *zapLogger) With(fields ...zap.Field) *zapLogger {
	lc := l.clone()
	lc.z = lc.z.With(fields...)
	return lc
}

// C 解析传入的 context，尝试提取关注的键值，并添加到 zap.Logger 结构化日志中.
func C(ctx context.Context) Logger {
	return current().C(ctx)
}

func (l *zapLogger) C(ctx context.Context) Logger {
	lc := l.clone()

	if ctx == nil {
		return lc
	}

	if requestID := ctx.Value(ctxkey.XRequestIDKey); requestID != nil {
		lc.z = lc.z.With(zap.Any(ctxkey.XRequestIDKey, requestID))
	}

	return lc
}

// clone 拷贝 zapLogger，派生 logger 不拥有轮转文件.
func (l *zapLogger) clone() *zapLogger {
	lc := *l
	lc.ownsRotate = false
	return &lc
}
