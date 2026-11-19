// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// newRateLimitRouter 创建测试路由，请求可通过 X-Test-IP 头指定客户端 IP
// （在 RateLimit 之前写入 RemoteAddr，确保限流器按预期 IP 计算）.
func newRateLimitRouter(rate float64, burst int) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.SetTrustedProxies(nil)
	r.Use(func(c *gin.Context) {
		if ip := c.Request.Header.Get("X-Test-IP"); ip != "" {
			c.Request.RemoteAddr = ip
		}
		c.Next()
	})
	r.Use(RateLimit(rate, burst))
	r.GET("/ok", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	return r
}

func TestRateLimitBlocksBurstExceeded(t *testing.T) {
	r := newRateLimitRouter(0.0001, 3)

	// 桶容量 3：前 3 个请求通过，第 4 个被限流
	codes := make([]int, 0, 4)
	for i := 0; i < 4; i++ {
		req := httptest.NewRequest(http.MethodGet, "/ok", nil)
		req.Header.Set("X-Test-IP", "10.0.0.1:12345")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		codes = append(codes, w.Code)
	}

	for i, code := range codes[:3] {
		if code != http.StatusOK {
			t.Errorf("第 %d 个请求 = %d, want 200", i+1, code)
		}
	}
	if codes[3] != http.StatusTooManyRequests {
		t.Errorf("第 4 个请求 = %d, want 429", codes[3])
	}
}

func TestRateLimitRefillsOverTime(t *testing.T) {
	r := newRateLimitRouter(1000, 1)

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	req.Header.Set("X-Test-IP", "10.0.0.2:12345")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("首次请求 = %d, want 200", w.Code)
	}

	// rate 足够大，等待令牌补充后应再次通过
	time.Sleep(5 * time.Millisecond)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("补充令牌后请求 = %d, want 200", w.Code)
	}
}

func TestRateLimitIsolatedByClientIP(t *testing.T) {
	r := newRateLimitRouter(0.0001, 1)

	req1 := httptest.NewRequest(http.MethodGet, "/ok", nil)
	req1.Header.Set("X-Test-IP", "10.0.1.1:12345")
	req2 := httptest.NewRequest(http.MethodGet, "/ok", nil)
	req2.Header.Set("X-Test-IP", "10.0.1.2:12345")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req1)
	if w.Code != http.StatusOK {
		t.Fatalf("IP1 首次请求 = %d, want 200", w.Code)
	}

	// 另一个 IP 不应受 IP1 限流影响
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req2)
	if w.Code != http.StatusOK {
		t.Errorf("IP2 请求 = %d, want 200", w.Code)
	}

	// IP1 第二次请求应被限流
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req1)
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("IP1 第二次请求 = %d, want 429", w.Code)
	}
}

func TestRateLimitInvalidParamsPassThrough(t *testing.T) {
	// rate/burst 非法时不限流，直接放行
	for _, params := range [][2]float64{{0, 10}, {10, 0}, {-1, -1}} {
		r := newRateLimitRouter(params[0], int(params[1]))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/ok", nil))
		if w.Code != http.StatusOK {
			t.Errorf("RateLimit(%v, %v) 应放行, got %d", params[0], params[1], w.Code)
		}
	}
}

func TestRateLimit429Body(t *testing.T) {
	r := newRateLimitRouter(0.0001, 1)

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	req.Header.Set("X-Test-IP", "10.0.2.1:12345")
	r.ServeHTTP(httptest.NewRecorder(), req) // 取走唯一令牌

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("状态码 = %d, want 429", w.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("响应体不是合法 JSON: %v\n%s", err, w.Body.String())
	}
	if code, ok := body["code"].(float64); !ok || code != 429 {
		t.Errorf("响应 code = %v, want 429", body["code"])
	}
}

func TestRateLimiterAllow(t *testing.T) {
	l := newRateLimiter(0.0001, 1)
	base := time.Now()

	if !l.allow("10.0.3.1", base) {
		t.Fatal("首次 allow 应成功")
	}
	if l.allow("10.0.3.1", base.Add(10*time.Millisecond)) {
		t.Fatal("令牌取空后 allow 应失败")
	}
	// 不同 key 互不影响
	if !l.allow("10.0.3.2", base) {
		t.Error("不同 IP 的 allow 不应受影响")
	}
}

func TestRateLimiterPurgeStaleBuckets(t *testing.T) {
	l := newRateLimiter(1, 1)
	l.maxSize = 2
	base := time.Now()

	l.allow("a", base)
	l.allow("b", base)
	// 达到阈值触发清理：a、b 均已过期，应被删除
	l.allow("c", base.Add(11*time.Minute))

	l.mu.Lock()
	_, aExists := l.buckets["a"]
	_, bExists := l.buckets["b"]
	cBucket, cExists := l.buckets["c"]
	l.mu.Unlock()

	if aExists || bExists {
		t.Errorf("过期桶应被清理: a=%v b=%v", aExists, bExists)
	}
	if !cExists || cBucket.tokens < 0 {
		t.Error("新桶应存在且未被错误清理")
	}
}
