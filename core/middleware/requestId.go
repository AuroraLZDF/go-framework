// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package middleware

import (
	"context"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	ctxkey "github.com/AuroraLZDF/go-framework/core/context"
)

// requestIDPattern 限制客户端传入的 Request ID 格式，避免超长或特殊字符进入日志.
var requestIDPattern = regexp.MustCompile(`^[A-Za-z0-9._-]{1,128}$`)

// RequestID 是一个 Gin 中间件，用来在每一个 HTTP 请求的 context, response 中注入 `X-Request-ID` 键值对.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 仅复用格式合法的 `X-Request-ID`，否则重新生成
		requestID := c.Request.Header.Get(ctxkey.XRequestIDKey)
		if !requestIDPattern.MatchString(requestID) {
			requestID = uuid.New().String()
		}

		// 将 RequestID 保存在 gin.Context 中，方便后边程序使用
		c.Set(ctxkey.XRequestIDKey, requestID)

		// 同时写入 request context，便于下游通过 log.C(ctx) 关联日志
		ctx := context.WithValue(c.Request.Context(), ctxkey.XRequestIDKey, requestID)
		c.Request = c.Request.WithContext(ctx)

		// 将 RequestID 保存在 HTTP 返回头中，Header 的键为 `X-Request-ID`
		c.Writer.Header().Set(ctxkey.XRequestIDKey, requestID)

		c.Next()
	}
}
