// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/AuroraLZDF/go-framework/core/config"
)

func newHardeningApp(t *testing.T) *Application {
	t.Helper()

	app, err := NewApplication(testConfig())
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}
	t.Cleanup(func() { _ = app.Close() })

	return app
}

func TestHealthzIsLivenessOnly(t *testing.T) {
	app := newHardeningApp(t)

	w := httptest.NewRecorder()
	app.GetEngine().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if !strings.Contains(w.Body.String(), `"status":"ok"`) {
		t.Errorf("响应应包含 status=ok, body = %s", w.Body.String())
	}
}

func TestReadyzWithNoDependencies(t *testing.T) {
	app := newHardeningApp(t)

	w := httptest.NewRecorder()
	app.GetEngine().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
	}
	body := w.Body.String()
	for _, want := range []string{`"status":"ok"`, `"mysql":"skipped"`, `"redis":"skipped"`} {
		if !strings.Contains(body, want) {
			t.Errorf("响应缺少 %s, body = %s", want, body)
		}
	}
}

func TestBodyLimitRejectsOversizedRequest(t *testing.T) {
	cfg := testConfig()
	cfg.Server.MaxBodyBytes = 64

	app, err := NewApplication(cfg)
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}
	t.Cleanup(func() { _ = app.Close() })

	req := httptest.NewRequest(http.MethodPost, "/no-route", strings.NewReader(strings.Repeat("x", 128)))
	if req.ContentLength <= 64 {
		t.Fatalf("测试请求 ContentLength = %d, 应大于限制", req.ContentLength)
	}

	w := httptest.NewRecorder()
	app.GetEngine().ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusRequestEntityTooLarge, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"code":413`) {
		t.Errorf("响应应包含业务码 413, body = %s", w.Body.String())
	}
}

func TestBodyLimitAllowsNormalRequest(t *testing.T) {
	cfg := testConfig()
	cfg.Server.MaxBodyBytes = 1024

	app, err := NewApplication(cfg)
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}
	t.Cleanup(func() { _ = app.Close() })

	app.GetEngine().POST("/echo", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	app.GetEngine().ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader("hello")))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestTrustedProxiesDisabledByDefault(t *testing.T) {
	app := newHardeningApp(t)
	app.GetEngine().GET("/ip", func(c *gin.Context) {
		c.String(http.StatusOK, c.ClientIP())
	})

	req := httptest.NewRequest(http.MethodGet, "/ip", nil)
	req.RemoteAddr = "192.0.2.10:12345"
	req.Header.Set("X-Forwarded-For", "203.0.113.9")

	w := httptest.NewRecorder()
	app.GetEngine().ServeHTTP(w, req)

	if got := w.Body.String(); got != "192.0.2.10" {
		t.Errorf("默认不信任代理时 ClientIP() = %q, want %q", got, "192.0.2.10")
	}
}

func TestTrustedProxiesConfigured(t *testing.T) {
	cfg := testConfig()
	cfg.Server.TrustedProxies = []string{"127.0.0.1"}

	app, err := NewApplication(cfg)
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}
	t.Cleanup(func() { _ = app.Close() })

	app.GetEngine().GET("/ip", func(c *gin.Context) {
		c.String(http.StatusOK, c.ClientIP())
	})

	req := httptest.NewRequest(http.MethodGet, "/ip", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set("X-Forwarded-For", "203.0.113.9")

	w := httptest.NewRecorder()
	app.GetEngine().ServeHTTP(w, req)

	if got := w.Body.String(); got != "203.0.113.9" {
		t.Errorf("可信代理下 ClientIP() = %q, want %q", got, "203.0.113.9")
	}
}

func TestHTTPSRedirectHandler(t *testing.T) {
	cfg := testConfig()
	cfg.Server.HTTPAddr = ":8080"
	cfg.Server.HTTPSAddr = ":8443"
	cfg.Server.TLSCert = "/tmp/cert.pem"
	cfg.Server.TLSKey = "/tmp/key.pem"

	app, err := NewApplication(cfg)
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}
	t.Cleanup(func() { _ = app.Close() })

	req := httptest.NewRequest(http.MethodGet, "http://example.com/x?y=1", nil)
	req.Host = "example.com:8080"

	w := httptest.NewRecorder()
	app.httpsRedirectHandler().ServeHTTP(w, req)

	if w.Code != http.StatusPermanentRedirect {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusPermanentRedirect)
	}
	if loc := w.Header().Get("Location"); loc != "https://example.com:8443/x?y=1" {
		t.Errorf("Location = %q, want %q", loc, "https://example.com:8443/x?y=1")
	}
}

func TestHTTPSRedirectHandlerOmitsDefaultPort(t *testing.T) {
	cfg := testConfig()
	cfg.Server.HTTPAddr = ":80"
	cfg.Server.HTTPSAddr = ":443"
	cfg.Server.TLSCert = "/tmp/cert.pem"
	cfg.Server.TLSKey = "/tmp/key.pem"

	app, err := NewApplication(cfg)
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}
	t.Cleanup(func() { _ = app.Close() })

	req := httptest.NewRequest(http.MethodGet, "http://example.com/x", nil)
	req.Host = "example.com"

	w := httptest.NewRecorder()
	app.httpsRedirectHandler().ServeHTTP(w, req)

	if loc := w.Header().Get("Location"); loc != "https://example.com/x" {
		t.Errorf("Location = %q, want %q", loc, "https://example.com/x")
	}
}

func TestRedisServerAddr(t *testing.T) {
	tests := []struct {
		name string
		addr string
		port string
		want string
	}{
		{"host only", "127.0.0.1", "6379", "127.0.0.1:6379"},
		{"host and port", "127.0.0.1:6380", "6379", "127.0.0.1:6380"},
		{"default port", "redis.internal", "", "redis.internal:6379"},
		{"ipv6 without port", "::1", "6379", "[::1]:6379"},
		{"empty", "", "6379", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := redisServerAddr(config.RedisConfig{Addr: tt.addr, Port: tt.port})
			if got != tt.want {
				t.Errorf("redisServerAddr() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHTTPSRedirectIPv6(t *testing.T) {
	tests := []struct {
		name      string
		host      string
		httpsAddr string
		want      string
	}{
		{"ipv6 default port", "[::1]", ":443", "https://[::1]/x?y=1"},
		{"ipv6 non-default port", "[::1]", ":8443", "https://[::1]:8443/x?y=1"},
		{"ipv6 with request port", "[::1]:8080", ":8443", "https://[::1]:8443/x?y=1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := testConfig()
			cfg.Server.HTTPAddr = ":8080"
			cfg.Server.HTTPSAddr = tt.httpsAddr
			cfg.Server.TLSCert = "/tmp/cert.pem"
			cfg.Server.TLSKey = "/tmp/key.pem"

			app, err := NewApplication(cfg)
			if err != nil {
				t.Fatalf("NewApplication() error = %v", err)
			}
			t.Cleanup(func() { _ = app.Close() })

			req := httptest.NewRequest(http.MethodGet, "http://example.com/x?y=1", nil)
			req.Host = tt.host

			w := httptest.NewRecorder()
			app.httpsRedirectHandler().ServeHTTP(w, req)

			if w.Code != http.StatusPermanentRedirect {
				t.Fatalf("status = %d, want %d", w.Code, http.StatusPermanentRedirect)
			}
			if loc := w.Header().Get("Location"); loc != tt.want {
				t.Errorf("Location = %q, want %q", loc, tt.want)
			}
			if cc := w.Header().Get("Cache-Control"); cc != "no-store" {
				t.Errorf("Cache-Control = %q, want no-store", cc)
			}
		})
	}
}

func TestBodyLimit413KeepsCORSHeaders(t *testing.T) {
	cfg := testConfig()
	cfg.Server.MaxBodyBytes = 64

	app, err := NewApplication(cfg)
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}
	t.Cleanup(func() { _ = app.Close() })

	req := httptest.NewRequest(http.MethodPost, "/no-route", strings.NewReader(strings.Repeat("x", 128)))
	req.Header.Set("Origin", "https://example.com")

	w := httptest.NewRecorder()
	app.GetEngine().ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusRequestEntityTooLarge)
	}
	if acao := w.Header().Get("Access-Control-Allow-Origin"); acao != "*" {
		t.Errorf("413 响应缺少 CORS 头, Access-Control-Allow-Origin = %q", acao)
	}
}
