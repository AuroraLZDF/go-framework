// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package token

const (
	// MinSecretLength 是 HS256 签名密钥的最小长度（32 字节 / 256 位）.
	// HMAC-SHA256 的密钥强度不应低于其输出长度，过短的密钥可被暴力破解.
	MinSecretLength = 32

	// defaultExpireHours 是未显式配置过期时间时使用的默认值（单位：小时）.
	defaultExpireHours = 24

	// defaultTokenIDClaim 是未显式配置时用于存放用户 ID 的 claim 名.
	defaultTokenIDClaim = "uid"
)

// Options 包含 JWT 签发与校验相关的配置项.
type Options struct {
	// 用于存放唯一用户的字段
	TokenId string
	// jwt 秘钥（至少 MinSecretLength 字节）
	Secret string
	// token 过期时间（单位：小时）
	Expire int
	// 黑名单在 Redis 中的 key 前缀，默认为 jwt:blacklist:
	BlacklistPrefix string
	// 黑名单校验失败（如 Redis 不可用）时是否拒绝请求.
	// 默认 false：放行请求，保证鉴权可用性；对安全性要求高的场景应设为 true.
	BlacklistFailClosed bool
}
