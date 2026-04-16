// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/auroralzdf/biddingAI.

package middleware

import (
	"github.com/gin-gonic/gin"
)

// CORS 是跨域资源共享中间件，用来处理跨域请求
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置允许的源
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		// 设置允许的方法
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		// 设置允许的头
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		// 设置是否允许携带凭证
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		// 设置暴露的头
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		// 设置预检请求的缓存时间
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
