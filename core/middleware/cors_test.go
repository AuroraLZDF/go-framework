// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// serve 使用指定的中间件处理一次请求
func serve(t *testing.T, mw gin.HandlerFunc, method, path, origin string) *httptest.ResponseRecorder {
	t.Helper()

	r := gin.New()
	r.Use(mw)
	r.Any("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(method, path, nil)
	if origin != "" {
		req.Header.Set("Origin", origin)
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	return w
}

func TestCORSDefaultAllowsAnyOriginWithoutCredentials(t *testing.T) {
	w := serve(t, CORS(), http.MethodGet, "/test", "https://example.com")

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("Access-Control-Allow-Origin = %q, want *", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "" {
		t.Errorf("默认配置不应允许携带凭证, got %q", got)
	}
	if got := w.Header().Get("Vary"); got != "Origin" {
		t.Errorf("Vary = %q, want Origin（所有分支都必须携带）", got)
	}
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestCORSExplicitWildcardAllowsAnyOrigin(t *testing.T) {
	w := serve(t, CORS("*"), http.MethodGet, "/test", "https://example.com")

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("显式 \"*\" 应允许任意源, Access-Control-Allow-Origin = %q", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "" {
		t.Errorf("通配符模式不应允许携带凭证, got %q", got)
	}
	if got := w.Header().Get("Vary"); got != "Origin" {
		t.Errorf("Vary = %q, want Origin", got)
	}
}

func TestCORSWhitelistEchoesOrigin(t *testing.T) {
	mw := CORS("https://allowed.example.com")

	w := serve(t, mw, http.MethodGet, "/test", "https://allowed.example.com")
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://allowed.example.com" {
		t.Errorf("Access-Control-Allow-Origin = %q", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("Access-Control-Allow-Credentials = %q, want true", got)
	}
	if got := w.Header().Get("Vary"); got != "Origin" {
		t.Errorf("Vary = %q, want Origin", got)
	}
}

func TestCORSWhitelistRejectsUnknownOrigin(t *testing.T) {
	mw := CORS("https://allowed.example.com")

	w := serve(t, mw, http.MethodGet, "/test", "https://evil.example.com")
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("非白名单来源不应返回 Access-Control-Allow-Origin, got %q", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "" {
		t.Errorf("非白名单来源不应允许凭证, got %q", got)
	}
	if got := w.Header().Get("Vary"); got != "Origin" {
		t.Errorf("Vary = %q, want Origin", got)
	}
}

func TestCORSPreflight(t *testing.T) {
	w := serve(t, CORS(), http.MethodOptions, "/test", "https://example.com")

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNoContent)
	}
	methods := w.Header().Get("Access-Control-Allow-Methods")
	if !strings.Contains(methods, "PATCH") {
		t.Errorf("Access-Control-Allow-Methods 应包含 PATCH, got %q", methods)
	}

	headers := w.Header().Get("Access-Control-Allow-Headers")
	for _, want := range []string{
		"Authorization",
		"Content-Type",
		"X-Request-ID",
		"X-Requested-With",
		"Accept",
		"Origin",
	} {
		if !strings.Contains(headers, want) {
			t.Errorf("Access-Control-Allow-Headers 应包含 %s, got %q", want, headers)
		}
	}
}

func TestCorsDelegatesToCORS(t *testing.T) {
	w := serve(t, Cors, http.MethodGet, "/test", "https://example.com")

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("Cors 应等价于 CORS(), Access-Control-Allow-Origin = %q", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "" {
		t.Errorf("Cors 不应允许携带凭证, got %q", got)
	}
}

func TestSecureHeadersAndNext(t *testing.T) {
	r := gin.New()
	r.Use(Secure)
	called := false
	r.GET("/test", func(c *gin.Context) {
		called = true
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))

	if !called {
		t.Error("Secure 应调用 c.Next() 继续执行后续处理")
	}
	if got := w.Header().Get("X-Frame-Options"); got != "DENY" {
		t.Errorf("X-Frame-Options = %q, want DENY", got)
	}
	if got := w.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("X-Content-Type-Options = %q, want nosniff", got)
	}
	if got := w.Header().Get("Content-Security-Policy"); got != "default-src 'self'" {
		t.Errorf("Content-Security-Policy = %q, want default-src 'self'", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Secure 不应设置 CORS 头, got %q", got)
	}
}
