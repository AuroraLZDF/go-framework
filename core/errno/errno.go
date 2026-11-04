// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/auroralzdf/biddingAI.

package errno

import (
	"errors"
	"fmt"
)

// Errno 定义了 biddingAI 使用的错误类型.
type Errno struct {
	HTTP    int
	Code    int
	Message string
}

// Error 实现 error 接口中的 `Error` 方法.
func (err *Errno) Error() string {
	return err.Message
}

// SetMessage 返回一个带有自定义 Message 的 Errno 副本，避免修改全局错误码定义.
func (err *Errno) SetMessage(format string, args ...interface{}) *Errno {
	clone := *err
	clone.Message = fmt.Sprintf(format, args...)
	return &clone
}

// Decode 尝试从 err 中解析出业务错误码和错误信息.
func Decode(err error) (int, int, string) {
	if err == nil {
		return OK.HTTP, OK.Code, OK.Message
	}

	// 使用 errors.As，保证被 %w 包装后的业务错误也能正确解析
	var typed *Errno
	if errors.As(err, &typed) {
		return typed.HTTP, typed.Code, typed.Message
	}

	// 未知错误统一按服务端错误处理，不向客户端回显原始错误信息，避免泄露内部细节
	return InternalServerError.HTTP, InternalServerError.Code, InternalServerError.Message
}
