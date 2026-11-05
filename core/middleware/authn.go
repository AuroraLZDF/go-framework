// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package middleware

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	jwtpkg "github.com/AuroraLZDF/go-framework/core/auth/jwt"
	ctxkey "github.com/AuroraLZDF/go-framework/core/context"
	"github.com/AuroraLZDF/go-framework/core/errno"
	"github.com/AuroraLZDF/go-framework/core/response"
)

// Authn 是认证中间件，用来从 gin.Context 中提取 token 并验证 token 是否合法，
// 如果合法则将 token 中的 sub 作为<用户名>存放在 gin.Context 的 XUsernameKey 键中.
func Authn() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 解析 JWT Token
		userId, err := jwtpkg.ParseRequest(c.Request)
		if err != nil {
			// 区分过期、缺失、被吊销等错误，便于客户端决定是否刷新 token
			response.Error(c, authnErrno(err))
			c.Abort()

			return
		}

		userID, err := strconv.ParseInt(userId, 10, 64)
		if err != nil {
			response.Error(c, errno.ErrTokenInvalid)
			c.Abort()

			return
		}

		// 设置全局变量
		c.Set(ctxkey.XUsernameKey, userID)
		c.Next()
	}
}

// authnErrno 把 ParseRequest 返回的错误映射为可直接输出的 *errno.Errno.
// 未携带 errno 信息的错误（例如 JWT 未初始化）按非法 token 处理，避免泄漏内部细节.
func authnErrno(err error) *errno.Errno {
	var e *errno.Errno
	if errors.As(err, &e) {
		return e
	}

	return errno.ErrTokenInvalid
}
