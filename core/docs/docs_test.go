// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package docs

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

const testSpec = `openapi: 3.0.3
info:
  title: test-api
  version: 1.0.0
paths:
  /health:
    get:
      summary: health check
      responses:
        '200':
          description: ok
`

func newTestEngine(t *testing.T, specPath string, opts Options) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	opts.SpecPath = specPath
	if err := Register(engine, opts); err != nil {
		t.Fatalf("Register: %v", err)
	}

	return engine
}

func doRequest(t *testing.T, engine *gin.Engine, path string) *httptest.ResponseRecorder {
	t.Helper()

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, http.NoBody))

	return w
}

func TestRegisterAndServe(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "openapi.yaml")
	if err := os.WriteFile(specPath, []byte(testSpec), 0o600); err != nil {
		t.Fatalf("write spec: %v", err)
	}

	engine := newTestEngine(t, specPath, Options{})

	// Swagger UI 页面
	w := doRequest(t, engine, "/docs")
	if w.Code != http.StatusOK {
		t.Errorf("GET /docs = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "/docs/openapi") {
		t.Error("UI page should reference the spec URL")
	}
	if got := w.Header().Get("Content-Type"); !strings.Contains(got, "text/html") {
		t.Errorf("UI content-type = %q, want text/html", got)
	}

	// 规范文件
	w = doRequest(t, engine, "/docs/openapi")
	if w.Code != http.StatusOK {
		t.Errorf("GET /docs/openapi = %d, want 200", w.Code)
	}
	if got := w.Header().Get("Content-Type"); got != "application/yaml" {
		t.Errorf("spec content-type = %q, want application/yaml", got)
	}
	if !strings.Contains(w.Body.String(), "openapi: 3.0.3") {
		t.Errorf("spec body missing content, got: %s", w.Body.String())
	}
}

func TestRegisterJSONSpec(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "swagger.json")
	if err := os.WriteFile(specPath, []byte(`{"openapi":"3.0.3"}`), 0o600); err != nil {
		t.Fatalf("write spec: %v", err)
	}

	engine := newTestEngine(t, specPath, Options{RoutePrefix: "/api-docs"})

	w := doRequest(t, engine, "/api-docs")
	if w.Code != http.StatusOK {
		t.Errorf("custom prefix UI = %d, want 200", w.Code)
	}

	w = doRequest(t, engine, "/api-docs/openapi")
	if got := w.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("json spec content-type = %q, want application/json", got)
	}
}

func TestRegisterErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 缺少 SpecPath
	if err := Register(gin.New(), Options{}); err == nil {
		t.Error("Register without SpecPath should fail")
	}

	// 文件不存在
	err := Register(gin.New(), Options{SpecPath: filepath.Join(t.TempDir(), "missing.yaml")})
	if err == nil || !strings.Contains(err.Error(), "read spec") {
		t.Errorf("Register with missing spec should fail with read error, got %v", err)
	}

	// 不支持的扩展名
	dir := t.TempDir()
	specPath := filepath.Join(dir, "spec.txt")
	if err := os.WriteFile(specPath, []byte("x"), 0o600); err != nil {
		t.Fatalf("write spec: %v", err)
	}
	if err := Register(gin.New(), Options{SpecPath: specPath}); err == nil {
		t.Error("Register with .txt spec should fail")
	}
}
