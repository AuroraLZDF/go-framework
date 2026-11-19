// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/AuroraLZDF/go-framework/core/errno"
	"github.com/AuroraLZDF/go-framework/core/response"
)

// rateLimiterTTL 桶空闲多久后可被清理，避免长期累积导致内存膨胀.
const rateLimiterTTL = 10 * time.Minute

// rateLimiterMaxEntries 触发主动清理的桶数量阈值.
const rateLimiterMaxEntries = 100000

// tokenBucket 是按客户端维护的令牌桶.
type tokenBucket struct {
	tokens float64
	last   time.Time
}

// rateLimiter 是基于内存令牌桶的按 IP 限流器.
type rateLimiter struct {
	rate    float64 // 每秒补充的令牌数
	burst   int     // 桶容量
	maxAge  time.Duration
	maxSize int

	mu      sync.Mutex
	buckets map[string]*tokenBucket
}

// newRateLimiter 创建限流器；rate 为每秒允许请求数，burst 为桶容量.
func newRateLimiter(rate float64, burst int) *rateLimiter {
	return &rateLimiter{
		rate:    rate,
		burst:   burst,
		maxAge:  rateLimiterTTL,
		maxSize: rateLimiterMaxEntries,
		buckets: make(map[string]*tokenBucket),
	}
}

// allow 判断 key 是否可以取到一个令牌，并做惰性补桶与清理.
func (l *rateLimiter) allow(key string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.buckets) >= l.maxSize {
		for k, b := range l.buckets {
			if now.Sub(b.last) > l.maxAge {
				delete(l.buckets, k)
			}
		}
	}

	b, ok := l.buckets[key]
	if !ok {
		b = &tokenBucket{tokens: float64(l.burst), last: now}
		l.buckets[key] = b
	}

	// 按距上次访问的时间补充令牌，上限为桶容量
	if elapsed := now.Sub(b.last).Seconds(); elapsed > 0 {
		b.tokens = min(float64(l.burst), b.tokens+elapsed*l.rate)
	}
	b.last = now

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// RateLimit 是一个 Gin 中间件，基于内存令牌桶按客户端 IP 限流.
// rate 为每秒允许的请求数，burst 为突发容量；超限返回统一错误响应 HTTP 429.
// rate <= 0 或 burst <= 0 时不做限制，直接放行.
func RateLimit(rate float64, burst int) gin.HandlerFunc {
	if rate <= 0 || burst <= 0 {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	l := newRateLimiter(rate, burst)
	return func(c *gin.Context) {
		if c.Request == nil {
			c.Next()
			return
		}
		if !l.allow(c.ClientIP(), time.Now()) {
			response.Error(c, errno.ErrTooManyRequests)
			c.Abort()
			return
		}
		c.Next()
	}
}
