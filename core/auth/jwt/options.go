// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/auroralzdf/biddingAI.

package token

// Options 包含与日志相关的配置项.
type Options struct {
	// 用于存放唯一用户的字段
	TokenId string
	// jwt 秘钥
	Secret string
	// token 过期时间（单位：小时）
	Expire int
	// 黑名单存储路径
	BlackPath string
}
