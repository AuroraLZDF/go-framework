// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/auroralzdf/biddingAI.

package token

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	ctxkey "github.com/AuroraLZDF/go-framework/core/context"
	"github.com/AuroraLZDF/go-framework/core/errno"
	"github.com/AuroraLZDF/go-framework/core/log"
	"github.com/AuroraLZDF/go-framework/core/response"
)

var (
	// ErrMissingHeader 表示 `Authorization` 请求头为空.
	//
	// Deprecated: ParseToken 现在返回带 HTTP 语义的 errno.ErrTokenMissing.
	ErrMissingHeader = errors.New("the length of the `Authorization` header is zero")

	// ErrBlacklistToken 表示存在于黑名单中的 token.
	//
	// Deprecated: ParseRequest 现在返回 errno.ErrTokenBlacklisted.
	ErrBlacklistToken = errors.New("token has in blacklist and expire")

	// ErrParseToken 表示解析 token 字符串失败.
	//
	// Deprecated: token 相关的解析错误现在尽量返回 *errno.Errno，便于 errno.Decode 统一处理.
	ErrParseToken = errors.New("parse token error")

	// ErrTokenNotInitialized 表示尚未通过 Init 配置 JWT 密钥.
	ErrTokenNotInitialized = errors.New("jwt token is not initialized")

	// ErrInvalidOptions 表示 Init 传入的配置为空.
	ErrInvalidOptions = errors.New("jwt options must not be nil")

	// ErrEmptySecret 表示 JWT 密钥为空.
	ErrEmptySecret = errors.New("jwt secret must not be empty")

	// ErrWeakSecret 表示 JWT 密钥长度不足 MinSecretLength 字节.
	ErrWeakSecret = errors.New("jwt secret is too short")
)

// config 保存当前生效的 JWT 配置.
// 使用 atomic.Pointer 保证 Init 与并发读取（Sign/ParseToken）之间没有数据竞争；
// nil 表示尚未初始化，此时拒绝签发与验签，避免空密钥被用于伪造 token.
var config atomic.Pointer[Options]

// Init 校验并保存 JWT 配置.
// 应在签发或校验 token 之前调用（server.NewApplication 会自动完成接线）.
// Secret 不能为空且长度不得小于 MinSecretLength 字节.
func Init(opts *Options) error {
	if opts == nil {
		return ErrInvalidOptions
	}
	if opts.Secret == "" {
		return ErrEmptySecret
	}
	if len(opts.Secret) < MinSecretLength {
		return fmt.Errorf("%w: got %d bytes, want at least %d", ErrWeakSecret, len(opts.Secret), MinSecretLength)
	}

	// 复制一份，避免调用方在 Init 之后修改自己的 Options 影响运行期配置.
	cfg := *opts
	if cfg.Expire <= 0 {
		cfg.Expire = defaultExpireHours
	}
	if cfg.TokenId == "" {
		cfg.TokenId = defaultTokenIDClaim
	}
	if cfg.BlacklistPrefix == "" {
		cfg.BlacklistPrefix = defaultBlacklistPrefix
	}
	config.Store(&cfg)

	return nil
}

// Reset 清空当前 JWT 配置，供测试或进程重配使用.
func Reset() {
	config.Store(nil)
}

// currentOptions 返回当前配置；未初始化时返回 nil.
func currentOptions() *Options {
	return config.Load()
}

// ParseToken 解析令牌字符串为 jwt.Token 对象.
// 返回的错误尽量为 *errno.Errno，便于 errno.Decode 与 Authn 中间件做错误分类.
func ParseToken(r *http.Request) (*jwt.Token, error) {
	cfg := currentOptions()
	if cfg == nil {
		return nil, ErrTokenNotInitialized
	}

	// 1. 手工解析 Authorization 头：Bearer 大小写不敏感，允许多个空格，拒绝空 token
	header := strings.TrimSpace(r.Header.Get(ctxkey.XAuthorization))
	if header == "" {
		return nil, errno.ErrTokenMissing
	}

	fields := strings.Fields(header)
	if len(fields) != 2 || !strings.EqualFold(fields[0], ctxkey.XTokenType) {
		return nil, errno.ErrTokenInvalid
	}
	tokenString := fields[1]
	if tokenString == "" {
		return nil, errno.ErrTokenInvalid
	}

	// 2. 从 tokenString 解析出 token
	token, err := jwt.Parse(
		tokenString,
		func(token *jwt.Token) (interface{}, error) {
			// 确保 token 加密算法是预期的加密算法
			if token.Method != jwt.SigningMethodHS256 {
				return nil, jwt.ErrSignatureInvalid
			}

			return []byte(cfg.Secret), nil
		},
		// 固定签名算法，避免 HS256/HS384/HS512 混用
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		// 必须携带过期时间
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		// 过期单独映射，便于客户端触发 refresh；其余（格式、签名、缺少 exp 等）统一视为非法 token
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errno.ErrTokenExpired
		}

		return nil, errno.ErrTokenInvalid
	}
	// 检查 Token 是否有效
	if !token.Valid {
		return nil, errno.ErrTokenInvalid
	}

	return token, nil
}

// ParseRequest 从请求头中获取令牌，并将其传递给 ParseToken 函数以解析令牌.
func ParseRequest(r *http.Request) (string, error) {
	token, err := ParseToken(r)
	// 解析失败
	if err != nil {
		return "", err
	}

	// 先提取用户 ID，claims 不完整时直接判定为非法 token
	tokenID, ok := subjectFromClaims(token)
	if !ok {
		return "", errno.ErrTokenInvalid
	}

	// 会话级吊销依赖 jti，缺失 jti 的 token 一律视为非法
	if _, ok := jtiFromClaims(token); !ok {
		return "", errno.ErrTokenInvalid
	}

	// 判断 token 是否已被吊销
	blacklisted, err := IsBlacklisted(token)
	if err != nil {
		// 查询失败时由 BlacklistFailClosed 决定是否放行，这里只记录告警
		log.Warning("Failed to check token blacklist", "err", err)
	}
	if blacklisted {
		return "", errno.ErrTokenBlacklisted
	}

	return tokenID, nil
}

// subjectFromClaims 从 Token 的 claims 中安全提取用户 ID.
func subjectFromClaims(token *jwt.Token) (string, bool) {
	cfg := currentOptions()
	if cfg == nil {
		return "", false
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", false
	}

	tokenID, ok := claims[cfg.TokenId].(string)

	return tokenID, ok && tokenID != ""
}

// Sign 使用配置中的密钥签发 token，token 的 claims 中会存放传入的 subject.
func Sign(tokenId string) (*response.TokenResponse, error) {
	cfg := currentOptions()
	if cfg == nil {
		return nil, ErrTokenNotInitialized
	}
	if tokenId == "" {
		return nil, fmt.Errorf("jwt token id must not be empty")
	}

	now := time.Now()
	// Expire 单位为小时；过期时间 = 当前时间 + Expire
	expires := now.Add(time.Duration(cfg.Expire) * time.Hour)

	// Token 的内容，jti 用于支持按会话粒度吊销
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		cfg.TokenId: tokenId,
		jtiClaim:    uuid.NewString(),
		"nbf":       now.Unix(),
		"iat":       now.Unix(),
		"exp":       expires.Unix(),
	})

	// 签发 token
	tokenString, err := token.SignedString([]byte(cfg.Secret))
	if err != nil {
		return nil, err
	}

	return &response.TokenResponse{
		AccessToken: tokenString,
		TokenType:   ctxkey.XTokenType,
		// ExpiresIn 是有效期剩余秒数（不是绝对 Unix 时间戳）
		ExpiresIn: int64(time.Duration(cfg.Expire) * time.Hour / time.Second),
	}, nil
}
