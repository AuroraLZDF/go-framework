// Copyright 2022 Innkeeper Belm(孔令飞) <nosbelm@qq.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://elgo.

package errno

var (
	// ErrUserAlreadyExist 代表用户已经存在.
	ErrUserAlreadyExist = &Errno{HTTP: 200, Code: 1, Message: "User already exist."}

	// ErrUserNotFound 表示未找到用户.
	ErrUserNotFound = &Errno{HTTP: 200, Code: 1, Message: "User was not found."}

	// ErrPasswordIncorrect 表示密码不正确.
	ErrPasswordIncorrect = &Errno{HTTP: 200, Code: 1, Message: "Password was incorrect."}

	// ErrUserStatusDisable 表示用户状态禁用.
	ErrUserStatusDisable = &Errno{HTTP: 200, Code: 1, Message: "User status is disabled."}

	// ErrMissingUsernameOrEmail 缺少用户名或邮箱
	ErrMissingUsernameOrEmail = &Errno{HTTP: 200, Code: 1, Message: "Missing username or email."}
)
