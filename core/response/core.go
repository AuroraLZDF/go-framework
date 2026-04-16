// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/auroralzdf/biddingAI.

package response

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AuroraLZDF/go-framework/core/errno"
)

// Response 定义了发生错误时的返回消息.
type Response struct {
	// Code 指定了业务错误码.
	Code int `json:"code"`

	// Message 包含了可以直接对外展示的错误信息.
	Message string `json:"message"`

	// Content 包含了可以直接对外展示的数据信息
	Content interface{} `json:"content"`
}

// Success 将正确相应数据写入 HTTP 响应主体
func Success(c *gin.Context, data any, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: message,
		Content: data,
	})
}

// Error 使用 errno.Decode 方法，根据错误类型，尝试从 err 中提取业务错误码和错误信息.
func Error(c *gin.Context, err error) {
	httpCode, code, message := errno.Decode(err)
	c.JSON(httpCode, Response{
		Code:    code,
		Message: message,
	})
}
