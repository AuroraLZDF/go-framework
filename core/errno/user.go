// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package errno

var (
	// ErrUserAlreadyExist 代表用户已经存在.
	ErrUserAlreadyExist = &Errno{HTTP: 409, Code: 1, Message: "User already exist."}

	// ErrUserNotFound 表示未找到用户.
	ErrUserNotFound = &Errno{HTTP: 404, Code: 1, Message: "User was not found."}

	// ErrPasswordIncorrect 表示密码不正确.
	ErrPasswordIncorrect = &Errno{HTTP: 401, Code: 1, Message: "Password was incorrect."}

	// ErrUserStatusDisable 表示用户状态禁用.
	ErrUserStatusDisable = &Errno{HTTP: 403, Code: 1, Message: "User status is disabled."}

	// ErrMissingUsernameOrEmail 缺少用户名或邮箱
	ErrMissingUsernameOrEmail = &Errno{HTTP: 400, Code: 1, Message: "Missing username or email."}
)
