// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

// Package token 负责 JWT 的签发、解析与吊销（目录名为 jwt，包名为 token）.
//
// 目录名表达机制（JWT），包名表达职责（签发/校验 token）。二者不一致是
// 有意为之：golang-jwt 的公开类型（*jwt.Token）出现在本包的 ParseToken/
// IsBlacklisted 签名中，若本包也叫 jwt，同时使用两者时必须给其中一个起别名。
//
// 推荐用法：
//
//	import "github.com/AuroraLZDF/go-framework/core/auth/jwt" // 包标识符为 token
//
//	// 业务路由显式挂载认证中间件
//	r.Use(middleware.Authn())
//
// 通常情况下不需要手动调用 Init：server.NewApplication 会根据 config.Config
// 自动完成接线。只有直接使用本包（不经过 Application）时才需要调用 Init，
// 且 Secret 必须至少 token.MinSecretLength 字节。
//
// 全局状态：本包使用进程级配置（atomic.Pointer），重复调用 Init 会让最新配置
// 生效（last-wins）。测试中可用 Reset 清理；一个进程只应运行一个
// server.Application，详见 docs/api-stability.md。
package token
