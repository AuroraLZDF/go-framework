// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	jwtpkg "github.com/AuroraLZDF/go-framework/core/auth/jwt"
	"github.com/AuroraLZDF/go-framework/core/config"
	ctxkey "github.com/AuroraLZDF/go-framework/core/context"
	"github.com/AuroraLZDF/go-framework/core/middleware"
)

func testConfig() *config.Config {
	return &config.Config{
		App: config.AppConfig{Name: "test-app", Mode: gin.TestMode},
		JWT: config.JWTConfig{Secret: strings.Repeat("a", 32)},
	}
}

func TestNewApplicationNilConfig(t *testing.T) {
	if _, err := NewApplication(nil); err == nil {
		t.Fatal("NewApplication(nil) 应返回错误而不是 panic")
	}
}

func TestNewApplicationRequiresJWTSecret(t *testing.T) {
	cfg := testConfig()
	cfg.JWT.Secret = ""

	if _, err := NewApplication(cfg); err == nil {
		t.Fatal("缺少 JWT Secret 时应返回错误")
	}
}

func TestNewApplicationWiresJWT(t *testing.T) {
	app, err := NewApplication(testConfig())
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}
	t.Cleanup(func() { _ = app.Close() })

	engine := app.GetEngine()
	engine.GET("/admin", middleware.Authn(), func(c *gin.Context) {
		uid, _ := c.Get(ctxkey.XUsernameKey)
		c.JSON(http.StatusOK, gin.H{"uid": uid})
	})

	// 未携带 token 时必须返回 401
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/admin", nil))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("无 token 时 status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	// 能正常签发 token，说明配置已接线到 jwt 包
	resp, err := jwtpkg.Sign("999")
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+resp.AccessToken)
	w = httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("有效 token 时 status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "999") {
		t.Errorf("响应中应包含 token 中的 uid, body = %s", w.Body.String())
	}
}

func TestNoRouteReturnsUnifiedResponse(t *testing.T) {
	app, err := NewApplication(testConfig())
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}
	t.Cleanup(func() { _ = app.Close() })

	w := httptest.NewRecorder()
	app.GetEngine().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/not-exist", nil))

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
	body := w.Body.String()
	if !strings.Contains(body, `"code":404`) || !strings.Contains(body, "Page not found") {
		t.Errorf("404 响应结构不统一, body = %s", body)
	}
	if strings.Contains(body, "content") {
		t.Errorf("错误响应不应包含 content 字段, body = %s", body)
	}
}

func TestRecoveryReturnsUnifiedResponse(t *testing.T) {
	app, err := NewApplication(testConfig())
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}
	t.Cleanup(func() { _ = app.Close() })

	app.GetEngine().GET("/panic", func(c *gin.Context) {
		panic("internal detail: password=secret")
	})

	w := httptest.NewRecorder()
	app.GetEngine().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/panic", nil))

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Internal server error") {
		t.Errorf("500 响应结构不统一, body = %s", body)
	}
	if strings.Contains(body, "password=secret") {
		t.Errorf("响应不应泄露 panic 详情, body = %s", body)
	}
}

func TestCloseIsIdempotent(t *testing.T) {
	app, err := NewApplication(testConfig())
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}

	if err := app.Close(); err != nil {
		t.Fatalf("第一次 Close() error = %v", err)
	}
	if err := app.Close(); err != nil {
		t.Fatalf("重复 Close() error = %v", err)
	}
}

func TestNewApplicationFailsOnUnwritableLogDir(t *testing.T) {
	dir := t.TempDir()
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatalf("准备阻塞文件失败: %v", err)
	}

	cfg := testConfig()
	cfg.Log.Dir = filepath.Join(blocker, "logs")

	if _, err := NewApplication(cfg); err == nil {
		t.Fatal("日志目录不可写时 NewApplication 应返回错误，而不是静默丢失文件日志")
	}

	// 创建失败时应回滚 jwt/日志状态，再次创建应用应成功
	app, err := NewApplication(testConfig())
	if err != nil {
		t.Fatalf("失败回滚后再次创建应用 error = %v", err)
	}
	_ = app.Close()
}

func TestRunGracefulShutdown(t *testing.T) {
	cfg := testConfig()
	cfg.Server.HTTPAddr = "127.0.0.1:0"
	cfg.Log = config.LogConfig{Dir: t.TempDir(), Level: "info", Format: "json"}

	app, err := NewApplication(cfg)
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() { done <- app.Run(ctx) }()

	// 给 HTTP 服务器留出启动时间，再触发优雅关闭
	time.Sleep(150 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run() 未在超时内退出")
	}

	// 退出日志必须在关闭日志文件之前写入，否则日志文件中不会出现
	files, err := filepath.Glob(filepath.Join(cfg.Log.Dir, "*.log"))
	if err != nil {
		t.Fatalf("查找日志文件失败: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("未找到轮转日志文件")
	}

	found := false
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("读取日志文件 %s 失败: %v", file, err)
		}
		if strings.Contains(string(data), "Server exiting") {
			found = true
		}
	}
	if !found {
		t.Error("日志文件中应包含退出日志 \"Server exiting\"")
	}
}
