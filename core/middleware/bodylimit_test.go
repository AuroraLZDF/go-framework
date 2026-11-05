// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package middleware

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func newBodyLimitRouter(maxBytes int64) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(BodyLimit(maxBytes))
	router.POST("/echo", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	return router
}

func TestBodyLimitRejectsOversizedRequest(t *testing.T) {
	router := newBodyLimitRouter(64)

	req := httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader(strings.Repeat("x", 128)))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusRequestEntityTooLarge, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"code":413`) {
		t.Errorf("响应应包含业务码 413, body = %s", w.Body.String())
	}
}

func TestBodyLimitAllowsSmallRequest(t *testing.T) {
	router := newBodyLimitRouter(1024)

	req := httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader("hello"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestBodyLimitDisabledWhenNonPositive(t *testing.T) {
	router := newBodyLimitRouter(0)

	req := httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader(strings.Repeat("x", 4096)))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("未启用限制时 status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestIsRequestTooLarge(t *testing.T) {
	body := http.MaxBytesReader(httptest.NewRecorder(), io.NopCloser(strings.NewReader(strings.Repeat("x", 128))), 64)
	if _, err := io.ReadAll(body); !IsRequestTooLarge(err) {
		t.Fatalf("IsRequestTooLarge(%v) = false, want true", err)
	}
	if IsRequestTooLarge(errors.New("other error")) {
		t.Error("普通错误不应被识别为请求体超限")
	}
}
