// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package token

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"

	"github.com/AuroraLZDF/go-framework/core/errno"
	"github.com/AuroraLZDF/go-framework/core/util"
)

const (
	// defaultBlacklistPrefix 黑名单在 Redis 中的默认 key 前缀
	defaultBlacklistPrefix = "jwt:blacklist:"
	// jtiClaim 是 JWT 标准中用于标识单个令牌的 claim
	jtiClaim = "jti"
	// userRevokeKeyInfix 用于区分用户级吊销 key 与会话级 jti key
	userRevokeKeyInfix = "user:"
)

// blacklistKey 生成会话级黑名单在 Redis 中的 key
func blacklistKey(jti string) string {
	return blacklistPrefix() + jti
}

// userRevokeKey 生成用户级吊销在 Redis 中的 key
func userRevokeKey(userID string) string {
	return blacklistPrefix() + userRevokeKeyInfix + userID
}

// blacklistPrefix 返回当前配置的黑名单 key 前缀.
func blacklistPrefix() string {
	cfg := currentOptions()
	if cfg == nil || cfg.BlacklistPrefix == "" {
		return defaultBlacklistPrefix
	}

	return cfg.BlacklistPrefix
}

// blacklistFailClosed 返回黑名单查询失败时是否应拒绝请求.
// 默认 false（fail-open）：Redis 不可用时放行，保证鉴权可用性；
// 可通过 Options.BlacklistFailClosed=true 切换为 fail-closed.
func blacklistFailClosed() bool {
	cfg := currentOptions()
	if cfg == nil {
		return false
	}

	return cfg.BlacklistFailClosed
}

// jtiFromClaims 从 Token 的 claims 中安全提取 jti.
func jtiFromClaims(token *jwt.Token) (string, bool) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", false
	}

	jti, ok := claims[jtiClaim].(string)

	return jti, ok && jti != ""
}

// IsBlacklisted 检查 Token 是否已被吊销.
// 第一个返回值表示是否命中黑名单；第二个返回值表示吊销状态查询是否失败
// （例如 Redis 不可用），调用方可据此记录告警.
//
// 注意：即使会话级（jti）查询失败，仍会继续检查用户级吊销时间点，
// 避免 Redis 抖动时 fail-open 分支完全跳过用户级吊销检查。
func IsBlacklisted(token *jwt.Token) (bool, error) {
	jti, ok := jtiFromClaims(token)
	if !ok {
		// 无法解析出 jti 视为不可信，按黑名单处理
		return true, nil
	}

	cache := util.NewCache()
	ctx := context.Background()
	failClosed := blacklistFailClosed()

	// 1. 会话级吊销：jti 是否在黑名单中
	sessionHit, sessionErr := cache.Exists(ctx, blacklistKey(jti))
	if sessionErr == nil && sessionHit {
		return true, nil
	}

	// 2. 用户级吊销：该用户在吊销时间点之前签发的 token 均已失效
	userID, ok := subjectFromClaims(token)
	if !ok {
		return true, nil
	}

	revokedAt, getErr := cache.Get(ctx, userRevokeKey(userID))
	if getErr != nil {
		if errors.Is(getErr, redis.Nil) {
			// 用户级吊销记录不存在；若会话级查询失败，按配置决定放行还是拒绝
			if sessionErr != nil {
				return failClosed, sessionErr
			}

			return false, nil
		}

		return failClosed, getErr
	}

	revokedAtUnix, err := strconv.ParseInt(revokedAt, 10, 64)
	if err != nil {
		// 吊销时间格式异常时按已吊销处理，避免绕过
		//nolint:nilerr // 解析失败时按已吊销处理是安全语义，不是误用
		return true, nil
	}

	issuedAt, err := token.Claims.GetIssuedAt()
	if err != nil || issuedAt == nil {
		//nolint:nilerr // 无法确认签发时间时按已吊销处理（fail-secure）
		return true, nil
	}

	if issuedAt.Unix() <= revokedAtUnix {
		return true, nil
	}

	// 用户级检查通过，但如果会话级查询失败过：fail-closed 时拒绝，fail-open 时放行并返回错误
	if sessionErr != nil {
		return failClosed, sessionErr
	}

	return false, nil
}

// AddToBlacklist 将会话级令牌（jti）加入黑名单，过期时间与 Token 一致
func AddToBlacklist(r *http.Request) error {
	token, err := ParseToken(r)
	if err != nil {
		return err
	}

	jti, ok := jtiFromClaims(token)
	if !ok {
		return errno.ErrTokenInvalid
	}

	// 获取 Token 的过期时间
	expire, err := token.Claims.GetExpirationTime()
	if err != nil || expire == nil {
		return errno.ErrTokenInvalid
	}

	// Token 已过期，无需写入黑名单
	ttl := time.Until(expire.Time)
	if ttl <= 0 {
		return nil
	}

	return util.NewCache().Set(context.Background(), blacklistKey(jti), "1", ttl)
}

// FailureAllClient 失效指定用户的所有 Token.
// 通过记录吊销时间点实现：该用户在吊销时间点之前签发的 token 都会被视为失效，
// 因此不受 config.Expire 后续调整的影响.
func FailureAllClient(tokenId string) error {
	return util.NewCache().Set(
		context.Background(),
		userRevokeKey(tokenId),
		strconv.FormatInt(time.Now().Unix(), 10),
		0,
	)
}
