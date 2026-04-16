// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/auroralzdf/biddingAI.

package token

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"

	ctxkey "github.com/AuroraLZDF/go-framework/core/context"
	"github.com/AuroraLZDF/go-framework/core/errno"
	"github.com/AuroraLZDF/go-framework/core/response"
)

var (
	// ErrMissingHeader 表示 `Authorization` 请求头为空.
	ErrMissingHeader = errors.New("the length of the `Authorization` header is zero")

	// ErrBlacklistToken 表示存在于黑名单中的 token
	ErrBlacklistToken = errors.New("token has in blacklist and expire")

	// ErrParseToken 表示解析 token 字符串失败
	ErrParseToken = errors.New("parse token error")
)

var (
	config = Options{}
	once   sync.Once
)

// Init 使用指定的选项初始化 jwt.
func Init(opts *Options) {
	once.Do(func() {
		config = *opts
	})
}

// ParseToken 解析令牌字符串为 jwt.Token 对象
func ParseToken(r *http.Request) (*jwt.Token, error) {
	// 1. 从请求头中取出 tokenString
	var tokenString string
	header := r.Header.Get(ctxkey.XAuthorization)
	if len(header) == 0 {
		return nil, ErrMissingHeader
	}

	// 从请求头中取出 token
	if _, err := fmt.Sscanf(header, ctxkey.XTokenType+" %s", &tokenString); err != nil {
		return nil, err
	}

	// 2. 从 tokenString 解析出 token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 确保 token 加密算法是预期的加密算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}

		return []byte(config.Secret), nil
	})
	// 解析失败
	if err != nil {
		return nil, ErrParseToken
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

	// 判断 token 是否存在与黑名单中
	if ok := IsBlacklisted(token); ok {
		return "", ErrBlacklistToken
	}

	var tokenId string
	// 如果解析成功，从 token 中取出 token 的主题
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		tokenId = claims[config.TokenId].(string)
	}

	return tokenId, nil
}

// Sign 使用 jwtSecret 签发 token，token 的 claims 中会存放传入的 subject.
func Sign(tokenId string) (*response.TokenResponse, error) {
	var expires = time.Now().Add(time.Duration(config.Expire) * time.Hour).Unix()

	// Token 的内容
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		config.TokenId: tokenId,
		"nbf":          time.Now().Unix(),
		"iat":          time.Now().Unix(),
		"exp":          expires,
	})

	// 签发 token
	tokenString, err := token.SignedString([]byte(config.Secret))

	var response = response.TokenResponse{
		AccessToken: tokenString,
		TokenType:   ctxkey.XTokenType,
		ExpiresIn:   expires,
	}
	return &response, err
}
