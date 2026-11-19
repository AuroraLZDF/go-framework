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

func setupMetricsRouter(path string, handler gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Metrics())
	r.GET(path, handler)
	r.Handle("GET", "/metrics", MetricsHandler())
	return r
}

func TestMetricsRecordsRequest(t *testing.T) {
	r := setupMetricsRouter("/api/ping", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/ping", nil))

	body := fetchMetrics(t, r)
	for _, want := range []string{
		`http_requests_total{method="GET",route="/api/ping",status="200"}`,
		`http_request_duration_seconds_count{method="GET",route="/api/ping"}`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("metrics 输出缺少 %q:\n%s", want, body)
		}
	}
}

func TestMetricsUnknownRoute(t *testing.T) {
	r := setupMetricsRouter("/registered", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// 请求未注册的路由，FullPath 为空应记为 UNKNOWN
	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/unregistered", nil))

	body := fetchMetrics(t, r)
	if !strings.Contains(body, `route="UNKNOWN"`) {
		t.Errorf("未注册路由应记为 UNKNOWN:\n%s", body)
	}
}

func TestMetricsSkipsMetricsPath(t *testing.T) {
	r := setupMetricsRouter("/registered", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/metrics", nil))

	body := fetchMetrics(t, r)
	if strings.Contains(body, `route="/metrics"`) {
		t.Errorf("/metrics 自身不应被统计:\n%s", body)
	}
}

func TestMetricsNilRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Metrics())
	r.GET("/ok", func(c *gin.Context) { c.Status(http.StatusOK) })

	// c.Request 为 nil 时应安全放行（不会 panic）
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = nil
	Metrics()(c)
}

// fetchMetrics 请求 /metrics 并返回文本输出.
func fetchMetrics(t *testing.T, r *gin.Engine) string {
	t.Helper()

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("GET /metrics = %d, want 200", w.Code)
	}
	return w.Body.String()
}
