// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package config

import "time"

// Config 通用配置结构
type Config struct {
	// 应用基础配置
	App AppConfig `mapstructure:"app"`

	// 服务器配置
	Server ServerConfig `mapstructure:"server"`

	// 数据库配置
	Database DatabaseConfig `mapstructure:"database"`

	// Redis 配置
	Redis RedisConfig `mapstructure:"redis"`

	// JWT 配置
	JWT JWTConfig `mapstructure:"jwt"`

	// 日志配置
	Log LogConfig `mapstructure:"log"`
}

// AppConfig 应用配置
type AppConfig struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
	Mode    string `mapstructure:"mode"` // debug/release/test
}

// ServerConfig 服务器配置
type ServerConfig struct {
	HTTPAddr  string `mapstructure:"http-addr"`
	HTTPSAddr string `mapstructure:"https-addr"`
	TLSCert   string `mapstructure:"tls-cert"`
	TLSKey    string `mapstructure:"tls-key"`

	// ShutdownTimeout 优雅关闭的最长等待时间
	ShutdownTimeout time.Duration `mapstructure:"shutdown-timeout"`

	// HTTP 服务器超时与请求头大小限制，默认 10s/30s/60s/120s/1MB
	ReadHeaderTimeout time.Duration `mapstructure:"read-header-timeout"`
	ReadTimeout       time.Duration `mapstructure:"read-timeout"`
	WriteTimeout      time.Duration `mapstructure:"write-timeout"`
	IdleTimeout       time.Duration `mapstructure:"idle-timeout"`
	MaxHeaderBytes    int           `mapstructure:"max-header-bytes"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	MySQL MySQLOptions `mapstructure:"mysql"`
}

// MySQLOptions MySQL 配置选项
type MySQLOptions struct {
	Host                  string        `mapstructure:"host"`
	Username              string        `mapstructure:"username"`
	Password              string        `mapstructure:"password"`
	Database              string        `mapstructure:"database"`
	MaxIdleConnections    int           `mapstructure:"max-idle-connections"`
	MaxOpenConnections    int           `mapstructure:"max-open-connections"`
	MaxConnectionLifeTime time.Duration `mapstructure:"max-connection-life-time"`
	LogLevel              int           `mapstructure:"log-level"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Port     string `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`

	// 连接池与超时配置，默认 10/5s/3s/3s
	PoolSize     int           `mapstructure:"pool-size"`
	DialTimeout  time.Duration `mapstructure:"dial-timeout"`
	ReadTimeout  time.Duration `mapstructure:"read-timeout"`
	WriteTimeout time.Duration `mapstructure:"write-timeout"`
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret              string `mapstructure:"secret"`
	Expire              int    `mapstructure:"expire"`
	TokenID             string `mapstructure:"token-id"`
	BlacklistPrefix     string `mapstructure:"blacklist-prefix"`
	BlacklistFailClosed bool   `mapstructure:"blacklist-fail-closed"`
}

// LogConfig 日志配置
type LogConfig struct {
	DisableCaller     bool   `mapstructure:"disable-caller"`
	DisableStacktrace bool   `mapstructure:"disable-stacktrace"`
	Level             string `mapstructure:"level"`
	Format            string `mapstructure:"format"`
	Dir               string `mapstructure:"dir"`
}
