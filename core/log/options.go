// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package log

import (
	"go.uber.org/zap/zapcore"
)

// RotateConfig 日志轮转配置
type RotateConfig struct {
	// 按时间轮转：daily(每天), hourly(每小时)
	Interval string // 默认 daily
	// 日志文件基础路径，实际文件名为 ${BasePath}.YYYYMMDD.log
	BasePath string
	// 是否同时输出到 stdout
	Stdout bool // 默认 true
}

// Options 包含与日志相关的配置项.
type Options struct {
	// 是否开启 caller，如果开启会在日志中显示调用日志所在的文件和行号
	DisableCaller bool
	// 是否禁止在 panic 及以上级别打印堆栈信息
	DisableStacktrace bool
	// 指定日志级别，可选值：debug, info, warn, error, dpanic, panic, fatal
	Level string
	// 指定日志显示格式，可选值：console, json
	Format string
	// 指定日志输出位置
	OutputPaths []string
	// 日志轮转配置，如果不为 nil 则启用日志轮转
	Rotation *RotateConfig
}

// NewOptions 创建一个带有默认参数的 Options 对象.
func NewOptions() *Options {
	return &Options{
		DisableCaller:     false,
		DisableStacktrace: false,
		Level:             zapcore.InfoLevel.String(),
		Format:            "console",
		OutputPaths:       []string{"stdout"},
	}
}
