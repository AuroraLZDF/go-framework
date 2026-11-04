// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

const (
	// minJWTSecretLength 是 HS256 密钥的最小长度（256 位）
	minJWTSecretLength = 32
)

// envKeys 列出所有支持通过环境变量覆盖的配置项.
// viper 的 Unmarshal 只会处理已知 key，因此必须显式 BindEnv，
// 否则 YAML 中未出现的配置项无法通过 FRAMEWORK_* 注入.
var envKeys = []string{
	"app.name",
	"app.version",
	"app.mode",
	"server.http-addr",
	"server.https-addr",
	"server.tls-cert",
	"server.tls-key",
	"server.shutdown-timeout",
	"server.read-header-timeout",
	"server.read-timeout",
	"server.write-timeout",
	"server.idle-timeout",
	"server.max-header-bytes",
	"database.mysql.host",
	"database.mysql.username",
	"database.mysql.password",
	"database.mysql.database",
	"database.mysql.max-idle-connections",
	"database.mysql.max-open-connections",
	"database.mysql.max-connection-life-time",
	"database.mysql.log-level",
	"redis.addr",
	"redis.port",
	"redis.password",
	"redis.db",
	"redis.pool-size",
	"redis.dial-timeout",
	"redis.read-timeout",
	"redis.write-timeout",
	"jwt.secret",
	"jwt.expire",
	"jwt.token-id",
	"jwt.blacklist-prefix",
	"jwt.blacklist-fail-closed",
	"log.disable-caller",
	"log.disable-stacktrace",
	"log.level",
	"log.format",
	"log.dir",
}

// Load 加载配置文件
func Load(configPath string) (*Config, error) {
	// 使用局部 viper，避免全局状态在多次加载之间相互污染
	v := viper.New()

	if configPath != "" {
		// 从命令行选项指定的配置文件中读取
		v.SetConfigFile(configPath)
	} else {
		// 查找用户主目录
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home dir: %w", err)
		}

		// 将用 `$HOME/.framework` 目录加入到配置文件的搜索路径中
		v.AddConfigPath(filepath.Join(home, ".framework"))

		// 把当前目录加入到配置文件的搜索路径中
		v.AddConfigPath(".")

		// 设置配置文件格式为 YAML
		v.SetConfigType("yaml")

		// 配置文件名称（没有文件扩展名）
		v.SetConfigName("config")
	}

	// 读取环境变量的前缀为 FRAMEWORK
	v.SetEnvPrefix("FRAMEWORK")

	// 将 key 字符串中 '.' 和 '-' 替换为 '_'
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// 读取匹配的环境变量
	v.AutomaticEnv()

	// 显式绑定每个配置项，使 YAML 中不存在的 key 也能被环境变量覆盖
	for _, key := range envKeys {
		if err := v.BindEnv(key); err != nil {
			return nil, fmt.Errorf("failed to bind env for %s: %w", key, err)
		}
	}

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		// 未显式指定配置文件时，允许完全依赖环境变量与默认值启动
		if configPath != "" || !errors.As(err, &notFound) {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// applyDefaults 填充合理的默认值
func (c *Config) applyDefaults() {
	if c.App.Mode == "" {
		c.App.Mode = "release"
	}
	if c.Server.HTTPAddr == "" {
		c.Server.HTTPAddr = ":8080"
	}
	if c.Server.ShutdownTimeout <= 0 {
		c.Server.ShutdownTimeout = 10 * time.Second
	}
	if c.Server.ReadHeaderTimeout == 0 {
		c.Server.ReadHeaderTimeout = 10 * time.Second
	}
	if c.Server.ReadTimeout == 0 {
		c.Server.ReadTimeout = 30 * time.Second
	}
	if c.Server.WriteTimeout == 0 {
		c.Server.WriteTimeout = 60 * time.Second
	}
	if c.Server.IdleTimeout == 0 {
		c.Server.IdleTimeout = 120 * time.Second
	}
	if c.Server.MaxHeaderBytes == 0 {
		c.Server.MaxHeaderBytes = 1 << 20 // 1MB
	}
	if c.Redis.PoolSize == 0 {
		c.Redis.PoolSize = 10
	}
	if c.Redis.DialTimeout == 0 {
		c.Redis.DialTimeout = 5 * time.Second
	}
	if c.Redis.ReadTimeout == 0 {
		c.Redis.ReadTimeout = 3 * time.Second
	}
	if c.Redis.WriteTimeout == 0 {
		c.Redis.WriteTimeout = 3 * time.Second
	}
	if c.Log.Level == "" {
		c.Log.Level = "info"
	}
	if c.Log.Format == "" {
		c.Log.Format = "console"
	}
	if c.JWT.Expire <= 0 {
		c.JWT.Expire = 24
	}
	if c.JWT.TokenID == "" {
		c.JWT.TokenID = "uid"
	}
	if c.JWT.BlacklistPrefix == "" {
		c.JWT.BlacklistPrefix = "jwt:blacklist:"
	}
}

// Validate 验证配置的有效性，并填充默认值
func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("config is nil")
	}

	c.applyDefaults()

	// 枚举校验，避免非法值在运行时触发 panic
	switch c.App.Mode {
	case "debug", "release", "test":
	default:
		return fmt.Errorf("app.mode must be one of debug, release, test, got %q", c.App.Mode)
	}
	switch c.Log.Format {
	case "console", "json":
	default:
		return fmt.Errorf("log.format must be one of console, json, got %q", c.Log.Format)
	}

	// JWT 密钥必须显式配置，且长度满足 HS256 的最低要求
	if c.JWT.Secret == "" {
		return fmt.Errorf("jwt secret is required")
	}
	if len(c.JWT.Secret) < minJWTSecretLength {
		return fmt.Errorf("jwt secret must be at least %d bytes, got %d", minJWTSecretLength, len(c.JWT.Secret))
	}

	// 仅在配置了 MySQL 地址时校验必需的连接参数
	if c.Database.MySQL.Host != "" {
		if c.Database.MySQL.Username == "" {
			return fmt.Errorf("mysql username is required")
		}
		if c.Database.MySQL.Database == "" {
			return fmt.Errorf("mysql database is required")
		}
	}

	// 启用 HTTPS 时必须提供证书与私钥
	if c.Server.HTTPSAddr != "" && (c.Server.TLSCert == "" || c.Server.TLSKey == "") {
		return fmt.Errorf("tls cert and key are required when https-addr is set")
	}

	// HTTP 服务器超时必须为正数，避免出现无限等待或立即超时
	if c.Server.ReadHeaderTimeout <= 0 {
		return fmt.Errorf("server.read-header-timeout must be positive, got %s", c.Server.ReadHeaderTimeout)
	}
	if c.Server.ReadTimeout <= 0 {
		return fmt.Errorf("server.read-timeout must be positive, got %s", c.Server.ReadTimeout)
	}
	if c.Server.WriteTimeout <= 0 {
		return fmt.Errorf("server.write-timeout must be positive, got %s", c.Server.WriteTimeout)
	}
	if c.Server.IdleTimeout <= 0 {
		return fmt.Errorf("server.idle-timeout must be positive, got %s", c.Server.IdleTimeout)
	}
	if c.Server.MaxHeaderBytes <= 0 {
		return fmt.Errorf("server.max-header-bytes must be positive, got %d", c.Server.MaxHeaderBytes)
	}

	// Redis 连接池与超时不允许为负数（0 已在 applyDefaults 中填默认值）
	if c.Redis.PoolSize < 0 {
		return fmt.Errorf("redis.pool-size must not be negative, got %d", c.Redis.PoolSize)
	}
	if c.Redis.DialTimeout < 0 || c.Redis.ReadTimeout < 0 || c.Redis.WriteTimeout < 0 {
		return fmt.Errorf("redis timeouts must not be negative")
	}

	return nil
}
