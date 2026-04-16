// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/auroralzdf/biddingAI.

package known

const (
	// XRequestIDKey 用来定义 Gin 上下文中的键，代表请求的 uuid.
	XRequestIDKey = "X-Request-ID"

	// XUsernameKey 用来定义 Token-ID，代表当前登录用户的唯一ID
	XUsernameKey = "X-Username"

	// XAuthorization 用来定义 JWT Token 存储在 Header 中对应的 Key
	XAuthorization = "Authorization"

	// XTokenType 用来定义 Token type
	XTokenType = "Bearer"
)
