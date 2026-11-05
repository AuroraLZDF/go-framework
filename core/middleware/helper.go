// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// NoCache 是一个 Gin 中间件，用来禁止客户端缓存 HTTP 请求的返回结果.
func NoCache(c *gin.Context) {
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "Thu, 01 Jan 1970 00:00:00 GMT")
	c.Header("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
	c.Next()
}

// defaultCORS 是默认的跨域中间件实例
var defaultCORS = CORS()

// Cors 是 CORS 的兼容入口，等价于 CORS().
//
// Deprecated: 请直接使用 CORS() 创建中间件（Cors 将在 v2 中移除）.
func Cors(c *gin.Context) {
	defaultCORS(c)
}

// Secure 是一个 Gin 中间件，用来添加一些安全和资源访问相关的 HTTP 头.
func Secure(c *gin.Context) {
	c.Header("X-Frame-Options", "DENY")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("X-XSS-Protection", "1; mode=block")
	// 默认只允许同源资源，业务如引入 CDN 等可在更外层覆盖该头
	c.Header("Content-Security-Policy", "default-src 'self'")
	if c.Request.TLS != nil {
		c.Header("Strict-Transport-Security", "max-age=31536000")
	}
	c.Next()
}
