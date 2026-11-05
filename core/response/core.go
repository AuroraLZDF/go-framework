// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package response

import (
	"net/http"

	"github.com/gin-gonic/gin"

	ctxkey "github.com/AuroraLZDF/go-framework/core/context"
	"github.com/AuroraLZDF/go-framework/core/errno"
	"github.com/AuroraLZDF/go-framework/core/log"
)

// Response 定义了发生错误时的返回消息.
type Response struct {
	// Code 指定了业务错误码.
	Code int `json:"code"`

	// Message 包含了可以直接对外展示的错误信息.
	Message string `json:"message"`

	// Content 包含了可以直接对外展示的数据信息，错误响应中省略该字段
	Content interface{} `json:"content,omitempty"`
}

// Success 将正确相应数据写入 HTTP 响应主体
func Success(c *gin.Context, data any, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    errno.OK.Code,
		Message: message,
		Content: data,
	})
}

// Error 使用 errno.Decode 方法，根据错误类型，尝试从 err 中提取业务错误码和错误信息.
func Error(c *gin.Context, err error) {
	httpCode, code, message := errno.Decode(err)

	// 5xx 说明服务端出错，记录原始错误（其中可能包含敏感信息），但不返回给客户端
	if err != nil && httpCode >= http.StatusInternalServerError {
		log.Error("Request failed with server error",
			"err", err,
			"path", c.Request.URL.Path,
			"request_id", c.GetString(ctxkey.XRequestIDKey),
		)
	}

	c.JSON(httpCode, Response{
		Code:    code,
		Message: message,
	})
}
