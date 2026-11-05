// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// validSecret 是满足最小长度要求的测试密钥
const validSecret = "0123456789abcdef0123456789abcdef"

func TestValidateDefaults(t *testing.T) {
	cfg := Config{JWT: JWTConfig{Secret: validSecret}}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"App.Mode", cfg.App.Mode, "release"},
		{"Server.HTTPAddr", cfg.Server.HTTPAddr, ":8080"},
		{"Log.Level", cfg.Log.Level, "info"},
		{"Log.Format", cfg.Log.Format, "console"},
		{"JWT.TokenID", cfg.JWT.TokenID, "uid"},
		{"JWT.BlacklistPrefix", cfg.JWT.BlacklistPrefix, "jwt:blacklist:"},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
		}
	}

	if cfg.JWT.Expire != 24 {
		t.Errorf("JWT.Expire = %d, want 24", cfg.JWT.Expire)
	}
	if cfg.Server.ShutdownTimeout != 10*time.Second {
		t.Errorf("Server.ShutdownTimeout = %s, want 10s", cfg.Server.ShutdownTimeout)
	}
}

func TestValidateMySQLOptional(t *testing.T) {
	// 未配置 Host 时不要求 MySQL 参数
	cfg := Config{JWT: JWTConfig{Secret: validSecret}}
	cfg.Database.MySQL.Username = "root"
	if err := cfg.Validate(); err != nil {
		t.Fatalf("未配置 MySQL 时不应报错, got %v", err)
	}

	// 配置 Host 后必须提供用户名与数据库名
	cfg = Config{JWT: JWTConfig{Secret: validSecret}}
	cfg.Database.MySQL.Host = "127.0.0.1:3306"
	if err := cfg.Validate(); err == nil {
		t.Fatal("缺少 username 时应返回错误")
	}

	cfg.Database.MySQL.Username = "root"
	if err := cfg.Validate(); err == nil {
		t.Fatal("缺少 database 时应返回错误")
	}

	cfg.Database.MySQL.Database = "app"
	if err := cfg.Validate(); err != nil {
		t.Fatalf("MySQL 参数完整时不应报错, got %v", err)
	}
}

func TestValidateHTTPS(t *testing.T) {
	cfg := Config{Server: ServerConfig{HTTPSAddr: ":443"}, JWT: JWTConfig{Secret: validSecret}}
	if err := cfg.Validate(); err == nil {
		t.Fatal("启用 HTTPS 但未提供证书时应返回错误")
	}

	cfg.Server.TLSCert = "/tmp/cert.pem"
	cfg.Server.TLSKey = "/tmp/key.pem"
	if err := cfg.Validate(); err != nil {
		t.Fatalf("证书完整时不应报错, got %v", err)
	}
}

func TestValidateJWTSecret(t *testing.T) {
	tests := []struct {
		name    string
		secret  string
		wantErr bool
	}{
		{"missing", "", true},
		{"too short", "short-secret", true},
		{"valid", validSecret, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{JWT: JWTConfig{Secret: tt.secret}}
			err := cfg.Validate()
			if tt.wantErr && err == nil {
				t.Fatal("应返回错误")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("不应返回错误, got %v", err)
			}
		})
	}
}

func TestValidateEnums(t *testing.T) {
	cfg := Config{App: AppConfig{Mode: "production"}, JWT: JWTConfig{Secret: validSecret}}
	if err := cfg.Validate(); err == nil {
		t.Fatal("非法 app.mode 应返回错误")
	}

	cfg = Config{Log: LogConfig{Format: "xml"}, JWT: JWTConfig{Secret: validSecret}}
	if err := cfg.Validate(); err == nil {
		t.Fatal("非法 log.format 应返回错误")
	}
}

func TestValidateNil(t *testing.T) {
	var cfg *Config
	if err := cfg.Validate(); err == nil {
		t.Fatal("nil 配置应返回错误而不是 panic")
	}
}

func TestLoadEnvOverridesMissingKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	// YAML 中刻意不包含 log 与 database 段
	if err := os.WriteFile(path, []byte("app:\n  name: demo\n"), 0o600); err != nil {
		t.Fatalf("写入配置失败: %v", err)
	}

	t.Setenv("FRAMEWORK_LOG_LEVEL", "debug")
	t.Setenv("FRAMEWORK_DATABASE_MYSQL_HOST", "127.0.0.1:3307")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Log.Level != "debug" {
		t.Errorf("Log.Level = %q, want %q", cfg.Log.Level, "debug")
	}
	if cfg.Database.MySQL.Host != "127.0.0.1:3307" {
		t.Errorf("Database.MySQL.Host = %q, want %q", cfg.Database.MySQL.Host, "127.0.0.1:3307")
	}
}

func TestLoadEnvOnly(t *testing.T) {
	// 隔离 $HOME 与当前目录，确保没有可读取的配置文件
	t.Setenv("HOME", t.TempDir())
	t.Chdir(t.TempDir())

	t.Setenv("FRAMEWORK_APP_NAME", "env-app")
	t.Setenv("FRAMEWORK_JWT_SECRET", validSecret)

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("无配置文件时应能通过环境变量启动, got %v", err)
	}
	if cfg.App.Name != "env-app" {
		t.Errorf("App.Name = %q, want %q", cfg.App.Name, "env-app")
	}
	if cfg.JWT.Secret != validSecret {
		t.Errorf("JWT.Secret 未被环境变量覆盖")
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("Validate() error = %v", err)
	}
}

func TestLoadDoesNotShareGlobalState(t *testing.T) {
	dir := t.TempDir()

	writeConfig := func(name, appName, secret string) string {
		path := filepath.Join(dir, name)
		content := "app:\n  name: " + appName + "\njwt:\n  secret: " + secret + "\n"
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("写入配置失败: %v", err)
		}
		return path
	}

	first := writeConfig("first.yaml", "first-app", "first-secret")
	second := writeConfig("second.yaml", "second-app", "second-secret")

	cfg1, err := Load(first)
	if err != nil {
		t.Fatalf("Load(first) error = %v", err)
	}
	if cfg1.App.Name != "first-app" || cfg1.JWT.Secret != "first-secret" {
		t.Fatalf("第一次加载结果错误: %+v", cfg1)
	}

	cfg2, err := Load(second)
	if err != nil {
		t.Fatalf("Load(second) error = %v", err)
	}
	if cfg2.App.Name != "second-app" || cfg2.JWT.Secret != "second-secret" {
		t.Fatalf("第二次加载受到上次配置影响: %+v", cfg2)
	}
}

func TestValidateHTTPAndRedisDefaults(t *testing.T) {
	cfg := Config{JWT: JWTConfig{Secret: validSecret}}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if cfg.Server.ReadHeaderTimeout != 10*time.Second {
		t.Errorf("Server.ReadHeaderTimeout = %s, want 10s", cfg.Server.ReadHeaderTimeout)
	}
	if cfg.Server.ReadTimeout != 30*time.Second {
		t.Errorf("Server.ReadTimeout = %s, want 30s", cfg.Server.ReadTimeout)
	}
	if cfg.Server.WriteTimeout != 60*time.Second {
		t.Errorf("Server.WriteTimeout = %s, want 60s", cfg.Server.WriteTimeout)
	}
	if cfg.Server.IdleTimeout != 120*time.Second {
		t.Errorf("Server.IdleTimeout = %s, want 120s", cfg.Server.IdleTimeout)
	}
	if cfg.Server.MaxHeaderBytes != 1<<20 {
		t.Errorf("Server.MaxHeaderBytes = %d, want %d", cfg.Server.MaxHeaderBytes, 1<<20)
	}

	if cfg.Redis.PoolSize != 10 {
		t.Errorf("Redis.PoolSize = %d, want 10", cfg.Redis.PoolSize)
	}
	if cfg.Redis.DialTimeout != 5*time.Second {
		t.Errorf("Redis.DialTimeout = %s, want 5s", cfg.Redis.DialTimeout)
	}
	if cfg.Redis.ReadTimeout != 3*time.Second {
		t.Errorf("Redis.ReadTimeout = %s, want 3s", cfg.Redis.ReadTimeout)
	}
	if cfg.Redis.WriteTimeout != 3*time.Second {
		t.Errorf("Redis.WriteTimeout = %s, want 3s", cfg.Redis.WriteTimeout)
	}
}

func TestValidateRejectsInvalidServerAndRedisTimeouts(t *testing.T) {
	cfg := Config{JWT: JWTConfig{Secret: validSecret}}
	cfg.Server.ReadTimeout = -time.Second
	if err := cfg.Validate(); err == nil {
		t.Fatal("负数 server.read-timeout 应返回错误")
	}

	cfg = Config{JWT: JWTConfig{Secret: validSecret}}
	cfg.Server.MaxHeaderBytes = -1
	if err := cfg.Validate(); err == nil {
		t.Fatal("负数 server.max-header-bytes 应返回错误")
	}

	cfg = Config{JWT: JWTConfig{Secret: validSecret}}
	cfg.Redis.PoolSize = -1
	if err := cfg.Validate(); err == nil {
		t.Fatal("负数 redis.pool-size 应返回错误")
	}

	cfg = Config{JWT: JWTConfig{Secret: validSecret}}
	cfg.Redis.DialTimeout = -time.Second
	if err := cfg.Validate(); err == nil {
		t.Fatal("负数 redis.dial-timeout 应返回错误")
	}
}

func TestValidateTrustedProxiesAndBodyLimit(t *testing.T) {
	cfg := Config{JWT: JWTConfig{Secret: validSecret}}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if cfg.Server.MaxBodyBytes != 4<<20 {
		t.Errorf("Server.MaxBodyBytes = %d, want %d", cfg.Server.MaxBodyBytes, 4<<20)
	}
	if len(cfg.Server.TrustedProxies) != 0 {
		t.Errorf("默认不应信任任何代理, got %v", cfg.Server.TrustedProxies)
	}

	// 合法 IP / CIDR 应通过，并就地去除空白与空项
	cfg = Config{JWT: JWTConfig{Secret: validSecret}}
	cfg.Server.TrustedProxies = []string{" 127.0.0.1 ", "10.0.0.0/8", ""}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("合法 trusted-proxies 不应报错, got %v", err)
	}
	if len(cfg.Server.TrustedProxies) != 2 || cfg.Server.TrustedProxies[0] != "127.0.0.1" || cfg.Server.TrustedProxies[1] != "10.0.0.0/8" {
		t.Errorf("Validate 后 TrustedProxies 应被规范化, got %v", cfg.Server.TrustedProxies)
	}

	// 非法条目应返回错误
	cfg = Config{JWT: JWTConfig{Secret: validSecret}}
	cfg.Server.TrustedProxies = []string{"not-an-ip"}
	if err := cfg.Validate(); err == nil {
		t.Fatal("非法 trusted-proxies 应返回错误")
	}

	// 负数请求体上限应返回错误
	cfg = Config{JWT: JWTConfig{Secret: validSecret}}
	cfg.Server.MaxBodyBytes = -1
	if err := cfg.Validate(); err == nil {
		t.Fatal("负数 server.max-body-bytes 应返回错误")
	}
}

func TestValidateHTTPSDoesNotDefaultHTTPAddr(t *testing.T) {
	// 仅启用 HTTPS 时不应隐式监听明文 HTTP 端口
	cfg := Config{
		Server: ServerConfig{HTTPSAddr: ":8443", TLSCert: "/tmp/cert.pem", TLSKey: "/tmp/key.pem"},
		JWT:    JWTConfig{Secret: validSecret},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if cfg.Server.HTTPAddr != "" {
		t.Errorf("启用 HTTPS 且未显式配置 http-addr 时 Server.HTTPAddr = %q, want empty", cfg.Server.HTTPAddr)
	}
}

func TestLoadEnvTrustedProxiesAndBodyLimit(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Chdir(t.TempDir())

	t.Setenv("FRAMEWORK_JWT_SECRET", validSecret)
	t.Setenv("FRAMEWORK_SERVER_TRUSTED_PROXIES", "127.0.0.1, 10.0.0.0/8")
	t.Setenv("FRAMEWORK_SERVER_MAX_BODY_BYTES", "2048")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if len(cfg.Server.TrustedProxies) != 2 {
		t.Fatalf("TrustedProxies = %v, want 2 个元素", cfg.Server.TrustedProxies)
	}
	if cfg.Server.TrustedProxies[0] != "127.0.0.1" || cfg.Server.TrustedProxies[1] != "10.0.0.0/8" {
		t.Errorf("TrustedProxies = %v", cfg.Server.TrustedProxies)
	}
	if cfg.Server.MaxBodyBytes != 2048 {
		t.Errorf("MaxBodyBytes = %d, want 2048", cfg.Server.MaxBodyBytes)
	}
}

func TestValidateRedisAddr(t *testing.T) {
	invalid := []string{
		"redis://localhost:6379",
		"host:6379:extra",
		"host:abc",
		"[::1]:70000",
	}
	for _, addr := range invalid {
		cfg := Config{JWT: JWTConfig{Secret: validSecret}, Redis: RedisConfig{Addr: addr}}
		if err := cfg.Validate(); err == nil {
			t.Errorf("redis.addr = %q 应校验失败", addr)
		}
	}

	valid := []string{
		"",
		"localhost",
		"127.0.0.1:6379",
		"::1",
		"[::1]",
		"[::1]:6379",
	}
	for _, addr := range valid {
		cfg := Config{JWT: JWTConfig{Secret: validSecret}, Redis: RedisConfig{Addr: addr}}
		if err := cfg.Validate(); err != nil {
			t.Errorf("redis.addr = %q 应校验通过, got %v", addr, err)
		}
	}
}
