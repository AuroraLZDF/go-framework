// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package log

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.uber.org/zap"

	ctxkey "github.com/AuroraLZDF/go-framework/core/context"
)

// jsonRotateOptions 构造输出到临时目录 json 轮转文件的调试级别日志配置
func jsonRotateOptions(basePath string) *Options {
	return &Options{
		Level:  "debug",
		Format: "json",
		Rotation: &RotateConfig{
			Interval: "daily",
			BasePath: basePath,
		},
	}
}

// readRotateLogs 读取 basePath 匹配到的所有轮转日志文件内容
func readRotateLogs(t *testing.T, basePath string) string {
	t.Helper()

	files, err := filepath.Glob(basePath + ".*.log")
	if err != nil {
		t.Fatalf("查找日志文件失败: %v", err)
	}
	if len(files) == 0 {
		t.Fatalf("未产生轮转日志文件, basePath = %s", basePath)
	}

	var sb strings.Builder
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("读取日志文件 %s 失败: %v", f, err)
		}
		sb.Write(data)
	}

	return sb.String()
}

// assertPanics 断言 fn 执行时会触发 panic
func assertPanics(t *testing.T, name string, fn func()) {
	t.Helper()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("%s 应触发 panic", name)
		}
	}()
	fn()
}

func TestNewLoggerWithRotation(t *testing.T) {
	base := filepath.Join(t.TempDir(), "app")

	logger := NewLogger(jsonRotateOptions(base))
	if logger == nil {
		t.Fatal("NewLogger() 不应返回 nil")
	}

	logger.Debug("debug message", "k", "v")
	logger.Info("info message")
	logger.Warning("warning message")
	logger.Error("error message")
	logger.Sync()

	// 顶层 logger 持有轮转文件，Close 应关闭它
	logger.Close()

	content := readRotateLogs(t, base)
	for _, want := range []string{
		`"message":"debug message"`,
		`"message":"info message"`,
		`"message":"warning message"`,
		`"message":"error message"`,
		`"level":"DEBUG"`,
		`"level":"INFO"`,
		`"level":"WARN"`,
		`"level":"ERROR"`,
		`"k":"v"`,
	} {
		if !strings.Contains(content, want) {
			t.Errorf("日志文件应包含 %q, content = %s", want, content)
		}
	}
}

// 说明：Fatal/Fatalw 内部调用 os.Exit(1)，无法在测试进程中验证，Fatal 分支跳过测试。

func TestNewLoggerPanicLevel(t *testing.T) {
	base := filepath.Join(t.TempDir(), "app")

	logger := NewLogger(jsonRotateOptions(base))
	assertPanics(t, "Panic", func() { logger.Panic("panic message") })
	logger.Close()

	content := readRotateLogs(t, base)
	if !strings.Contains(content, `"message":"panic message"`) {
		t.Errorf("日志文件应包含 panic 消息, content = %s", content)
	}
	if !strings.Contains(content, `"level":"PANIC"`) {
		t.Errorf("日志文件应包含 PANIC 级别, content = %s", content)
	}
}

func TestNewLoggerNilOptions(t *testing.T) {
	// nil 选项应回退为默认配置而不 panic
	logger := NewLogger(nil)
	if logger == nil {
		t.Fatal("NewLogger(nil) 不应返回 nil")
	}
	logger.Info("nil options works")
	logger.Sync()
}

func TestNewLoggerFallsBackWhenRotationFails(t *testing.T) {
	dir := t.TempDir()
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatalf("准备阻塞文件失败: %v", err)
	}

	// 轮转目录不可创建时应回退为 stdout logger，而不是返回 nil
	logger := NewLogger(rotationOptions(filepath.Join(blocker, "app")))
	if logger == nil {
		t.Fatal("初始化失败时 NewLogger 应回退为可用的 logger，而不是返回 nil")
	}
	logger.Info("fallback works")
}

func TestNewLoggerInvalidLevelAndEmptyOutputPaths(t *testing.T) {
	// 非法级别应回退为 info；空 OutputPaths 应默认输出到 stdout
	logger := NewLogger(&Options{Level: "not-a-level", Format: "json"})
	if logger == nil {
		t.Fatal("非法级别时 NewLogger 应回退为默认级别并返回可用 logger")
	}
	logger.Info("invalid level works")
	logger.Sync()
}

func TestWithAddsFieldsAndDerivedCloseKeepsFileOpen(t *testing.T) {
	base := filepath.Join(t.TempDir(), "app")

	l, rw, err := newLogger(jsonRotateOptions(base))
	if err != nil {
		t.Fatalf("newLogger() error = %v", err)
	}

	derived := l.With(zap.String("svc", "unit-test"))
	derived.Info("from derived")

	// 派生 logger 不拥有轮转文件，Close 不应关闭底层文件
	derived.Close()
	if rw.closed {
		t.Fatal("派生 logger 的 Close 不应关闭轮转文件")
	}

	// 派生 logger 关闭后，顶层 logger 仍应能继续写入
	l.Info("from owner")
	l.Sync()
	l.Close()

	content := readRotateLogs(t, base)
	if !strings.Contains(content, `"message":"from derived"`) {
		t.Errorf("日志文件应包含派生 logger 的消息, content = %s", content)
	}
	if !strings.Contains(content, `"svc":"unit-test"`) {
		t.Errorf("With 添加的字段应写入日志, content = %s", content)
	}
	if !strings.Contains(content, `"message":"from owner"`) {
		t.Errorf("派生 logger 关闭后顶层 logger 仍应可写入, content = %s", content)
	}
}

func TestCWithContext(t *testing.T) {
	base := filepath.Join(t.TempDir(), "app")

	l, _, err := newLogger(jsonRotateOptions(base))
	if err != nil {
		t.Fatalf("newLogger() error = %v", err)
	}

	// 携带 requestID 的 context 应把请求 ID 写入日志
	reqCtx := context.WithValue(context.Background(), ctxkey.XRequestIDCtxKey, "req-123")
	c1 := l.C(reqCtx)
	c1.Info("with request id")

	// 普通 context 与 nil context 应安全返回可用 logger
	c2 := l.C(context.Background())
	c2.Info("without request id")
	// nil context 经变量传入（SA1012 拦截字面量 nil），验证运行时 nil 安全
	var nilCtx context.Context
	c3 := l.C(nilCtx)
	c3.Info("with nil context")

	l.Sync()
	l.Close()

	content := readRotateLogs(t, base)
	for _, want := range []string{
		`"message":"with request id"`,
		`"X-Request-ID":"req-123"`,
		`"message":"without request id"`,
		`"message":"with nil context"`,
	} {
		if !strings.Contains(content, want) {
			t.Errorf("日志文件应包含 %q, content = %s", want, content)
		}
	}
}

func TestPackageLevelMethods(t *testing.T) {
	Reset()
	t.Cleanup(Reset)

	base := filepath.Join(t.TempDir(), "app")
	if err := Init(jsonRotateOptions(base)); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	Debug("global debug", "k", "v")
	Info("global info")
	Warning("global warning")
	Error("global error")
	Sync()

	// 包级 Panic 会 panic，用 recover 捕获后继续断言
	assertPanics(t, "Panic", func() { Panic("global panic") })

	content := readRotateLogs(t, base)
	for _, want := range []string{
		`"message":"global debug"`,
		`"message":"global info"`,
		`"message":"global warning"`,
		`"message":"global error"`,
		`"message":"global panic"`,
		`"level":"DEBUG"`,
		`"level":"WARN"`,
		`"level":"ERROR"`,
		`"level":"PANIC"`,
		`"k":"v"`,
	} {
		if !strings.Contains(content, want) {
			t.Errorf("日志文件应包含 %q, content = %s", want, content)
		}
	}
}

func TestRotateWriterHourlyAndSync(t *testing.T) {
	base := filepath.Join(t.TempDir(), "hourly", "app")

	w, err := NewRotateWriter(base, "hourly")
	if err != nil {
		t.Fatalf("NewRotateWriter(hourly) error = %v", err)
	}

	// hourly 模式的时间键包含小时段，文件名中应出现日期_小时格式
	key := w.getCurrentKey()
	if !strings.Contains(key, "_") {
		t.Errorf("hourly 模式的时间键应包含 '_', got %s", key)
	}
	if want := base + "." + key + ".log"; w.fileName(key) != want {
		t.Errorf("fileName() = %s, want %s", w.fileName(key), want)
	}

	if _, err := w.Write([]byte("hourly line\n")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := w.Sync(); err != nil {
		t.Fatalf("Sync() error = %v", err)
	}

	if err := w.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	// 关闭后 Sync 与重复 Close 都应安全返回
	if err := w.Sync(); err != nil {
		t.Errorf("关闭后 Sync() error = %v, want nil", err)
	}
	if err := w.Close(); err != nil {
		t.Errorf("重复 Close() error = %v, want nil", err)
	}

	if content := readRotateLogs(t, base); !strings.Contains(content, "hourly line") {
		t.Errorf("日志文件应包含写入内容, content = %s", content)
	}
}
