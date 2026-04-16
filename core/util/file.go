// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/auroralzdf/biddingAI.

package util

import (
	"os"
	"path/filepath"
)

// AddFileIfNotExists add file if not exists
func AddFileIfNotExists(path string) error {
	// 检查文件或目录是否存在
	if _, err := os.Stat(path); err == nil {
		return nil
	}

	// 创建目录
	if err := AddDir(path); err != nil {
		return err
	}

	// 创建文件
	if err := AddFile(path); err != nil {
		return err
	}
	return nil
}

// AddDir add dir
func AddDir(path string) error {
	// 创建目录
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return nil
}

// AddFile add file
func AddFile(path string) error {
	// 创建文件
	if _, err := os.Create(path); err != nil {
		return err
	}
	return nil
}
