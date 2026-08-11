// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/auroralzdf/biddingAI.

package log

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// RotateWriter 按时间轮转的日志写入器
type RotateWriter struct {
	mu       sync.Mutex
	basePath string
	interval string // daily 或 hourly
	file     *os.File
	dateStr  string
}

// NewRotateWriter 创建一个新的轮转写入器
func NewRotateWriter(basePath, interval string) *RotateWriter {
	w := &RotateWriter{
		basePath: basePath,
		interval: interval,
	}
	if w.interval == "" {
		w.interval = "daily"
	}
	_ = w.rotate()
	return w
}

// Write 实现 io.Writer 接口
func (w *RotateWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 检查是否需要轮转
	current := w.getCurrentKey()
	if current != w.dateStr {
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}

	return w.file.Write(p)
}

// Sync 同步日志到磁盘
func (w *RotateWriter) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file != nil {
		return w.file.Sync()
	}
	return nil
}

// Close 关闭当前日志文件
func (w *RotateWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

// getCurrentKey 获取当前时间键
func (w *RotateWriter) getCurrentKey() string {
	if w.interval == "hourly" {
		return time.Now().Format("2006-01-02_15")
	}
	return time.Now().Format("2006-01-02")
}

// rotate 执行日志轮转
func (w *RotateWriter) rotate() error {
	if w.file != nil {
		w.file.Close()
	}

	w.dateStr = w.getCurrentKey()
	filename := fmt.Sprintf("%s.%s.log", w.basePath, w.dateStr)

	dir := filepath.Dir(filename)
	if dir != "" {
		_ = os.MkdirAll(dir, 0755)
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	w.file = file
	return nil
}
