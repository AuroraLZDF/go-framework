// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	ctxkey "github.com/AuroraLZDF/go-framework/core/context"
)

func TestRequestIDGeneratesAndPropagates(t *testing.T) {
	var ginValue, requestValue any

	r := gin.New()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		ginValue, _ = c.Get(ctxkey.XRequestIDKey)
		requestValue = c.Request.Context().Value(ctxkey.XRequestIDKey)
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))

	header := w.Header().Get(ctxkey.XRequestIDKey)
	if header == "" {
		t.Fatal("响应头应包含 X-Request-ID")
	}
	if ginValue != header {
		t.Errorf("gin.Context 中的 RequestID = %v, want %q", ginValue, header)
	}
	if requestValue != header {
		t.Errorf("request context 中的 RequestID = %v, want %q", requestValue, header)
	}
}

func TestRequestIDRejectsInvalidClientValue(t *testing.T) {
	invalid := strings.Repeat("x", 129)

	r := gin.New()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(ctxkey.XRequestIDKey, invalid)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if got := w.Header().Get(ctxkey.XRequestIDKey); got == invalid {
		t.Error("非法 Request ID 应被替换为服务端生成的值")
	}
}

func TestRequestIDKeepsValidClientValue(t *testing.T) {
	const valid = "client-id_123.abc"

	r := gin.New()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(ctxkey.XRequestIDKey, valid)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if got := w.Header().Get(ctxkey.XRequestIDKey); got != valid {
		t.Errorf("合法 Request ID 应被复用, got %q, want %q", got, valid)
	}
}
