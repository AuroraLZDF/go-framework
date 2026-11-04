// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package server

import (
	"time"

	"github.com/gin-gonic/gin"

	ctxkey "github.com/AuroraLZDF/go-framework/core/context"
	"github.com/AuroraLZDF/go-framework/core/log"
)

// accessLog 返回基于 zap 的访问日志中间件.
// 与 gin.Logger() 不同，访问日志会走统一的日志系统，配置了日志目录时也会写入轮转文件.
// 该中间件依赖 RequestID 中间件，必须在 RequestID 之后注册，才能记录 request_id.
func accessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		status := c.Writer.Status()
		fields := []any{
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", status,
			"latency_ms", float64(time.Since(start)) / float64(time.Millisecond),
			"client_ip", c.ClientIP(),
			"size", c.Writer.Size(),
		}

		if query := c.Request.URL.RawQuery; query != "" {
			fields = append(fields, "query", query)
		}
		if requestID := c.GetString(ctxkey.XRequestIDKey); requestID != "" {
			fields = append(fields, "request_id", requestID)
		}

		if status >= 500 {
			log.Error("HTTP request failed", fields...)
			return
		}
		log.Info("HTTP request", fields...)
	}
}
