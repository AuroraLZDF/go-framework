// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package errno

var (
	// ErrUnauthorized 表示请求没有被授权.
	// ErrUnauthorized = &Errno{HTTP: 401, Code: "AuthFailure.Unauthorized", Message: "Unauthorized."}

	// ErrTokenMissing 表示请求头中没有令牌.
	ErrTokenMissing = &Errno{HTTP: 401, Code: 401, Message: "Token missing."}

	// ErrTokenExpired 表示令牌已过期.
	ErrTokenExpired = &Errno{HTTP: 401, Code: 401, Message: "Token expired."}

	// ErrTokenBlacklisted 表示令牌已被吊销（logout 或用户级失效）.
	ErrTokenBlacklisted = &Errno{HTTP: 401, Code: 401, Message: "Token has been revoked."}

	// ErrTokenRefreshExpired 表示令牌刷新已过期.
	ErrTokenRefreshExpired = &Errno{HTTP: 401, Code: 401, Message: "Token refresh expired."}

	// ErrTokenRefreshFailed 表示令牌刷新失败.
	ErrTokenRefreshFailed = &Errno{HTTP: 401, Code: 401, Message: "Token refresh failed."}

	// ErrTokenRefreshInvalid 表示令牌刷新无效.
	ErrTokenRefreshInvalid = &Errno{HTTP: 401, Code: 401, Message: "Token refresh invalid."}
)
