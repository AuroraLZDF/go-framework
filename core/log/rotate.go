// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package log

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ErrRotateWriterClosed 表示向已关闭的轮转写入器写入日志.
var ErrRotateWriterClosed = errors.New("rotate writer is closed")

// RotateWriter 按时间轮转的日志写入器
type RotateWriter struct {
	mu       sync.Mutex
	basePath string
	interval string // daily 或 hourly
	file     *os.File
	dateStr  string
	closed   bool
}

// NewRotateWriter 创建一个新的轮转写入器.
// 首次打开日志文件失败时返回错误，调用方（log.Init）应向上传播该错误，做到 fail-loud.
func NewRotateWriter(basePath, interval string) (*RotateWriter, error) {
	if basePath == "" {
		return nil, errors.New("log rotation base path is empty")
	}

	w := &RotateWriter{
		basePath: basePath,
		interval: interval,
	}
	if w.interval == "" {
		w.interval = "daily"
	}

	w.mu.Lock()
	defer w.mu.Unlock()
	if err := w.rotateLocked(); err != nil {
		return nil, err
	}

	return w, nil
}

// Write 实现 io.Writer 接口.
// 轮转失败时返回错误而不是静默丢弃日志；dateStr 只在打开成功后更新，
// 因此打开失败后的下一次写入仍会重试轮转.
func (w *RotateWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return 0, ErrRotateWriterClosed
	}

	current := w.getCurrentKey()
	if w.file == nil || current != w.dateStr {
		if err := w.rotateLocked(); err != nil {
			return 0, err
		}
	}

	return w.file.Write(p)
}

// Sync 同步日志到磁盘
func (w *RotateWriter) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed || w.file == nil {
		return nil
	}
	return w.file.Sync()
}

// Close 关闭当前日志文件，可重复调用
func (w *RotateWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return nil
	}
	w.closed = true
	w.dateStr = ""

	if w.file == nil {
		return nil
	}

	syncErr := w.file.Sync()
	closeErr := w.file.Close()
	w.file = nil

	if syncErr != nil {
		return syncErr
	}
	return closeErr
}

// getCurrentKey 获取当前时间键
func (w *RotateWriter) getCurrentKey() string {
	if w.interval == "hourly" {
		return time.Now().Format("2006-01-02_15")
	}
	return time.Now().Format("2006-01-02")
}

// fileName 返回指定时间键对应的日志文件名
func (w *RotateWriter) fileName(key string) string {
	return fmt.Sprintf("%s.%s.log", w.basePath, key)
}

// rotateLocked 执行日志轮转（调用方必须持有 w.mu）.
func (w *RotateWriter) rotateLocked() error {
	if w.closed {
		return ErrRotateWriterClosed
	}

	key := w.getCurrentKey()
	filename := w.fileName(key)

	if dir := filepath.Dir(filename); dir != "" && dir != "." {
		// 目录不可创建时返回错误，避免后续以静默丢日志的方式失败
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create log directory %q: %w", dir, err)
		}
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file %q: %w", filename, err)
	}

	if w.file != nil {
		_ = w.file.Close()
	}
	w.file = file
	// 只有打开成功后才推进 dateStr，失败后下次写入会重试
	w.dateStr = key

	return nil
}
