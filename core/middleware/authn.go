// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/auroralzdf/biddingAI.

package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"

	jwtpkg "go-framework/core/auth/jwt"
	ctxkey "go-framework/core/context"
	"go-framework/core/errno"
	"go-framework/core/response"
)

// Authn 是认证中间件，用来从 gin.Context 中提取 token 并验证 token 是否合法，
// 如果合法则将 token 中的 sub 作为<用户名>存放在 gin.Context 的 XUsernameKey 键中.
func Authn() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 解析 JWT Token
		userId, err := jwtpkg.ParseRequest(c.Request)
		if err != nil {
			response.Error(c, errno.ErrTokenInvalid)
			c.Abort()

			return
		}

		userID, _ := strconv.ParseInt(userId, 10, 64)
		// 设置全局变量
		c.Set(ctxkey.XUsernameKey, userID)
		c.Next()
	}
}
