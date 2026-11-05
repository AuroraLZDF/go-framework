// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package util

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// AddFileIfNotExists 若文件不存在则创建（包括父目录），已存在时直接返回 nil.
// 注意：检查与创建之间存在 TOCTOU 窗口，并发调用方应自行保证互斥.
func AddFileIfNotExists(path string) error {
	info, err := os.Stat(path)
	if err == nil {
		if info.IsDir() {
			return fmt.Errorf("util: %q is a directory, not a file", path)
		}

		return nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("util: stat %q: %w", path, err)
	}

	if err := AddDir(filepath.Dir(path)); err != nil {
		return err
	}

	return AddFile(path)
}

// AddDir 创建目录（含父目录）.
func AddDir(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("util: create dir %q: %w", path, err)
	}

	return nil
}

// AddFile 创建空文件；文件已存在时不覆盖、不截断，直接返回 nil.
// 创建后立即关闭文件句柄，避免句柄泄漏.
func AddFile(path string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		if errors.Is(err, fs.ErrExist) {
			return nil
		}

		return fmt.Errorf("util: create file %q: %w", path, err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("util: close file %q: %w", path, err)
	}

	return nil
}
