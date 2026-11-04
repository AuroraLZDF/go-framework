// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package log

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// rotationOptions 构造启用轮转的测试选项
func rotationOptions(basePath string) *Options {
	return &Options{
		Level:  "info",
		Format: "json",
		Rotation: &RotateConfig{
			Interval: "daily",
			BasePath: basePath,
		},
	}
}

func TestNewRotateWriterFailsWhenBasePathUnavailable(t *testing.T) {
	dir := t.TempDir()
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatalf("准备阻塞文件失败: %v", err)
	}

	if _, err := NewRotateWriter(filepath.Join(blocker, "app"), "daily"); err == nil {
		t.Fatal("日志目录不可创建时应返回错误")
	}
}

func TestInitFailsWhenRotateDirUnavailable(t *testing.T) {
	Reset()
	t.Cleanup(Reset)

	dir := t.TempDir()
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatalf("准备阻塞文件失败: %v", err)
	}

	err := Init(rotationOptions(filepath.Join(blocker, "app")))
	if err == nil {
		t.Fatal("轮转目录不可用时 Init 应返回错误，而不是静默丢失文件日志")
	}
}

func TestInitInvalidFormatReturnsError(t *testing.T) {
	Reset()
	t.Cleanup(Reset)

	if err := Init(&Options{Level: "info", Format: "yaml"}); err == nil {
		t.Fatal("非法日志格式应返回错误，而不是 panic")
	}

	// 初始化失败后全局日志仍应可用（保持上一次的 logger）
	Info("still works after failed init")
}

func TestCloseIsRefCounted(t *testing.T) {
	Reset()
	t.Cleanup(Reset)

	dir := t.TempDir()
	secondBase := filepath.Join(dir, "second", "app")

	if err := Init(rotationOptions(filepath.Join(dir, "first", "app"))); err != nil {
		t.Fatalf("第一次 Init() error = %v", err)
	}
	if err := Init(rotationOptions(secondBase)); err != nil {
		t.Fatalf("第二次 Init() error = %v", err)
	}

	Info("before first close")
	if err := Close(); err != nil {
		t.Fatalf("第一次 Close() error = %v", err)
	}

	// 引用计数尚未归零，后续日志应继续写入第二个文件
	Info("after first close")
	if err := Close(); err != nil {
		t.Fatalf("第二次 Close() error = %v", err)
	}
	Info("after second close")

	// 归零后重复 Close 应安全返回
	if err := Close(); err != nil {
		t.Fatalf("引用计数归零后 Close() error = %v", err)
	}

	files, err := filepath.Glob(secondBase + ".*.log")
	if err != nil {
		t.Fatalf("查找日志文件失败: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("期望 1 个日志文件, got %v", files)
	}
	data, err := os.ReadFile(files[0])
	if err != nil {
		t.Fatalf("读取日志文件失败: %v", err)
	}
	if !strings.Contains(string(data), "after first close") {
		t.Errorf("引用计数未归零时日志不应被关闭, file content = %s", data)
	}
}

func TestRotateWriterWriteAfterClose(t *testing.T) {
	base := filepath.Join(t.TempDir(), "sub", "app")

	w, err := NewRotateWriter(base, "daily")
	if err != nil {
		t.Fatalf("NewRotateWriter() error = %v", err)
	}
	if _, err := w.Write([]byte("hello\n")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	if _, err := w.Write([]byte("after close\n")); !errors.Is(err, ErrRotateWriterClosed) {
		t.Fatalf("关闭后写入应返回 ErrRotateWriterClosed, got %v", err)
	}
}
