// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

// Package known 定义框架在请求上下文（gin.Context / context.Context）中
// 共享的键，以及框架识别的 HTTP 请求头名称.
//
// 导入路径为 core/context，但包名为 known（历史命名，避免与标准库 context
// 冲突）。这些常量的值属于稳定 API：中间件、访问日志与业务代码通过这些
// 键交换 request_id、认证用户等信息，v1.x 内不会改变。
package known
