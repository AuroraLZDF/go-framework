// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AuroraLZDF/go-framework/core/errno"
	"github.com/AuroraLZDF/go-framework/core/response"
)

// IsRequestTooLarge 判断错误是否由 BodyLimit 的 http.MaxBytesReader 触发.
// 业务绑定（c.ShouldBindJSON 等）收到超限错误时，可据此映射为 413：
//
//	if err := c.ShouldBindJSON(&req); err != nil {
//	    if middleware.IsRequestTooLarge(err) {
//	        response.Error(c, errno.ErrRequestEntityTooLarge)
//	        return
//	    }
//	    response.Error(c, errno.ErrBind)
//	    return
//	}
func IsRequestTooLarge(err error) bool {
	var maxBytesErr *http.MaxBytesError
	return errors.As(err, &maxBytesErr)
}

// BodyLimit 是一个 Gin 中间件，用来限制请求体大小（单位：字节）.
// 当请求声明了 Content-Length 且超过限制时，直接返回统一错误响应 HTTP 413；
// 未声明长度（chunked）的请求体由 MaxBytesReader 限制，超限读取会返回错误，
// 业务绑定通常表现为 400，但内存占用仍被限制在 maxBytes 以内。
// maxBytes <= 0 时不做限制，直接放行。
func BodyLimit(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if maxBytes <= 0 || c.Request == nil {
			c.Next()
			return
		}

		// 已知长度时直接拒绝，避免继续读取请求体
		if c.Request.ContentLength > maxBytes {
			response.Error(c, errno.ErrRequestEntityTooLarge)
			c.Abort()
			return
		}

		// 未知长度（如 chunked）时通过 MaxBytesReader 兜底：
		// 超限读取返回 *http.MaxBytesError，由业务绑定错误路径处理
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		}

		c.Next()
	}
}
