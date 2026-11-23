// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

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

// ctxKey 是 context.Context 专用键类型。
// context.WithValue 的键不应使用内置 string（staticcheck SA1029），
// 因此存取 context 的地方请使用 XRequestIDCtxKey.
type ctxKey string

// XRequestIDCtxKey context.Context 中请求 ID 的键，与 Gin 上下文中的 XRequestIDKey 对应.
const XRequestIDCtxKey ctxKey = "X-Request-ID"
