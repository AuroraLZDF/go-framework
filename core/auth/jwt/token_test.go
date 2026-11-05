// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package token

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/AuroraLZDF/go-framework/core/errno"
	"github.com/AuroraLZDF/go-framework/core/util"
)

func testSecret() string {
	return strings.Repeat("a", MinSecretLength)
}

// initForTest 初始化测试配置，并在测试结束后清空，避免测试间相互影响.
func initForTest(t *testing.T, opts *Options) {
	t.Helper()

	if opts == nil {
		opts = &Options{TokenId: "uid", Secret: testSecret(), Expire: 1}
	}
	if err := Init(opts); err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	t.Cleanup(Reset)
}

func newRequest(tokenString string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	if tokenString != "" {
		req.Header.Set("Authorization", "Bearer "+tokenString)
	}

	return req
}

func newRawRequest(header string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	if header != "" {
		req.Header.Set("Authorization", header)
	}

	return req
}

func signRaw(t *testing.T, method jwt.SigningMethod, claims jwt.MapClaims, secret string) string {
	t.Helper()

	tokenString, err := jwt.NewWithClaims(method, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("签发 token 失败: %v", err)
	}

	return tokenString
}

func validClaims() jwt.MapClaims {
	now := time.Now()

	return jwt.MapClaims{
		"uid": "user-1",
		"jti": "jti-1",
		"nbf": now.Unix(),
		"iat": now.Unix(),
		"exp": now.Add(time.Hour).Unix(),
	}
}

func TestInitRejectsInvalidOptions(t *testing.T) {
	t.Cleanup(Reset)

	if err := Init(nil); !errors.Is(err, ErrInvalidOptions) {
		t.Errorf("Init(nil) error = %v, want ErrInvalidOptions", err)
	}
	if err := Init(&Options{Secret: ""}); !errors.Is(err, ErrEmptySecret) {
		t.Errorf("Init(空 Secret) error = %v, want ErrEmptySecret", err)
	}
	if err := Init(&Options{Secret: strings.Repeat("a", MinSecretLength-1)}); !errors.Is(err, ErrWeakSecret) {
		t.Errorf("Init(短 Secret) error = %v, want ErrWeakSecret", err)
	}
	// 失败时不应留下半初始化状态
	if _, err := Sign("1"); !errors.Is(err, ErrTokenNotInitialized) {
		t.Errorf("Init 失败后 Sign() error = %v, want ErrTokenNotInitialized", err)
	}
}

func TestInitAcceptsStrongSecretAndDefaults(t *testing.T) {
	t.Cleanup(Reset)

	if err := Init(&Options{Secret: testSecret()}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	cfg := currentOptions()
	if cfg == nil {
		t.Fatal("Init() 后配置不应为 nil")
	}
	if cfg.TokenId != defaultTokenIDClaim {
		t.Errorf("默认 TokenId = %q, want %q", cfg.TokenId, defaultTokenIDClaim)
	}
	if cfg.Expire != defaultExpireHours {
		t.Errorf("默认 Expire = %d, want %d", cfg.Expire, defaultExpireHours)
	}
}

func TestResetClearsConfig(t *testing.T) {
	initForTest(t, nil)

	Reset()

	if _, err := Sign("1"); !errors.Is(err, ErrTokenNotInitialized) {
		t.Errorf("Reset 后 Sign() error = %v, want ErrTokenNotInitialized", err)
	}
	if _, err := ParseToken(newRequest("a.b.c")); !errors.Is(err, ErrTokenNotInitialized) {
		t.Errorf("Reset 后 ParseToken() error = %v, want ErrTokenNotInitialized", err)
	}
}

func TestSignAndParseRequestRoundTrip(t *testing.T) {
	initForTest(t, nil)

	resp, err := Sign("user-1")
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}
	if resp.AccessToken == "" {
		t.Fatal("Sign() 返回的 token 为空")
	}

	uid, err := ParseRequest(newRequest(resp.AccessToken))
	if err != nil {
		t.Fatalf("ParseRequest() error = %v", err)
	}
	if uid != "user-1" {
		t.Errorf("ParseRequest() = %q, want %q", uid, "user-1")
	}
}

func TestSignExpiresInIsRemainingSeconds(t *testing.T) {
	initForTest(t, &Options{TokenId: "uid", Secret: testSecret(), Expire: 2})

	resp, err := Sign("user-1")
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	want := int64(2 * 60 * 60)
	if resp.ExpiresIn != want {
		t.Errorf("ExpiresIn = %d, want %d（剩余秒数，而不是绝对 Unix 时间戳）", resp.ExpiresIn, want)
	}
}

func TestSignEmptyTokenID(t *testing.T) {
	initForTest(t, nil)

	if _, err := Sign(""); err == nil {
		t.Error("Sign(\"\") 应返回错误")
	}
}

func TestSignRequiresInit(t *testing.T) {
	t.Cleanup(Reset)

	if _, err := Sign("user-1"); !errors.Is(err, ErrTokenNotInitialized) {
		t.Errorf("Sign() error = %v, want ErrTokenNotInitialized", err)
	}
}

func TestParseTokenMissingHeader(t *testing.T) {
	initForTest(t, nil)

	_, err := ParseToken(newRequest(""))
	if !errors.Is(err, errno.ErrTokenMissing) {
		t.Errorf("ParseToken() error = %v, want errno.ErrTokenMissing", err)
	}
}

func TestParseTokenMalformedHeader(t *testing.T) {
	initForTest(t, nil)

	testCases := []string{
		"not-a-bearer-token",
		"Basic abc",
		"Bearer",
		"Bearer ",
		"Bearer a b",
	}
	for _, header := range testCases {
		if _, err := ParseToken(newRawRequest(header)); !errors.Is(err, errno.ErrTokenInvalid) {
			t.Errorf("header %q error = %v, want errno.ErrTokenInvalid", header, err)
		}
	}
}

func TestParseTokenSchemeIsCaseInsensitive(t *testing.T) {
	initForTest(t, nil)

	resp, err := Sign("user-1")
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	for _, header := range []string{
		"bearer " + resp.AccessToken,
		"BEARER " + resp.AccessToken,
		"Bearer   " + resp.AccessToken,
	} {
		if _, err := ParseToken(newRawRequest(header)); err != nil {
			t.Errorf("header %q 应被接受, error = %v", header, err)
		}
	}
}

func TestParseTokenExpired(t *testing.T) {
	initForTest(t, nil)

	claims := validClaims()
	claims["exp"] = time.Now().Add(-time.Hour).Unix()
	tokenString := signRaw(t, jwt.SigningMethodHS256, claims, testSecret())

	_, err := ParseToken(newRequest(tokenString))
	if !errors.Is(err, errno.ErrTokenExpired) {
		t.Errorf("ParseToken() error = %v, want errno.ErrTokenExpired", err)
	}
}

func TestParseTokenInvalidSignature(t *testing.T) {
	initForTest(t, nil)

	tokenString := signRaw(t, jwt.SigningMethodHS256, validClaims(), "other-secret")

	_, err := ParseToken(newRequest(tokenString))
	if !errors.Is(err, errno.ErrTokenInvalid) {
		t.Errorf("ParseToken() error = %v, want errno.ErrTokenInvalid", err)
	}
}

func TestParseTokenRejectsWrongAlgorithm(t *testing.T) {
	initForTest(t, nil)

	// 使用相同密钥但不同算法签名，应被拒绝
	tokenString := signRaw(t, jwt.SigningMethodHS384, validClaims(), testSecret())

	_, err := ParseToken(newRequest(tokenString))
	if !errors.Is(err, errno.ErrTokenInvalid) {
		t.Errorf("ParseToken() error = %v, want errno.ErrTokenInvalid", err)
	}
}

func TestParseTokenRequiresExpiration(t *testing.T) {
	initForTest(t, nil)

	claims := validClaims()
	delete(claims, "exp")
	tokenString := signRaw(t, jwt.SigningMethodHS256, claims, testSecret())

	_, err := ParseToken(newRequest(tokenString))
	if !errors.Is(err, errno.ErrTokenInvalid) {
		t.Errorf("ParseToken() error = %v, want errno.ErrTokenInvalid", err)
	}
}

func TestParseRequestRequiresTokenID(t *testing.T) {
	initForTest(t, nil)

	claims := validClaims()
	delete(claims, "uid")
	tokenString := signRaw(t, jwt.SigningMethodHS256, claims, testSecret())

	_, err := ParseRequest(newRequest(tokenString))
	if !errors.Is(err, errno.ErrTokenInvalid) {
		t.Errorf("ParseRequest() error = %v, want errno.ErrTokenInvalid", err)
	}
}

func TestParseRequestRequiresJTI(t *testing.T) {
	initForTest(t, nil)

	claims := validClaims()
	delete(claims, "jti")
	tokenString := signRaw(t, jwt.SigningMethodHS256, claims, testSecret())

	_, err := ParseRequest(newRequest(tokenString))
	if !errors.Is(err, errno.ErrTokenInvalid) {
		t.Errorf("ParseRequest() error = %v, want errno.ErrTokenInvalid", err)
	}
}

func TestIsBlacklistedWithoutJTI(t *testing.T) {
	initForTest(t, nil)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uid": "user-1",
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	blacklisted, err := IsBlacklisted(token)
	if err != nil {
		t.Fatalf("IsBlacklisted() error = %v", err)
	}
	if !blacklisted {
		t.Error("缺少 jti 时应按黑名单处理")
	}
}

func TestIsBlacklistedWithoutRedis(t *testing.T) {
	// 默认 fail-open：Redis 不可用时放行，但把错误返回给调用方记录告警
	initForTest(t, nil)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uid": "user-1",
		"jti": "jti-1",
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	blacklisted, err := IsBlacklisted(token)
	if blacklisted {
		t.Error("fail-open 模式下 Redis 不可用不应判定为黑名单")
	}
	if !errors.Is(err, util.ErrCacheUnavailable) {
		t.Errorf("IsBlacklisted() error = %v, want ErrCacheUnavailable", err)
	}

	// fail-closed：Redis 不可用时拒绝请求
	if err := Init(&Options{
		TokenId:             "uid",
		Secret:              testSecret(),
		Expire:              1,
		BlacklistFailClosed: true,
	}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	blacklisted, err = IsBlacklisted(token)
	if !blacklisted {
		t.Error("fail-closed 模式下 Redis 不可用应判定为黑名单")
	}
	if !errors.Is(err, util.ErrCacheUnavailable) {
		t.Errorf("IsBlacklisted() error = %v, want ErrCacheUnavailable", err)
	}
}

func TestSignIncludesJTI(t *testing.T) {
	initForTest(t, nil)

	resp, err := Sign("user-1")
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	token, err := ParseToken(newRequest(resp.AccessToken))
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}

	jti, ok := jtiFromClaims(token)
	if !ok || jti == "" {
		t.Error("Sign() 签发的 token 应包含非空 jti")
	}
}

func TestBlacklistKeysUseConfiguredPrefix(t *testing.T) {
	t.Cleanup(Reset)

	if err := Init(&Options{
		TokenId:         "uid",
		Secret:          testSecret(),
		Expire:          1,
		BlacklistPrefix: "custom:blacklist:",
	}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	if got := blacklistKey("jti-1"); got != "custom:blacklist:jti-1" {
		t.Errorf("blacklistKey() = %q", got)
	}
	if got := userRevokeKey("user-1"); got != "custom:blacklist:user:user-1" {
		t.Errorf("userRevokeKey() = %q", got)
	}

	Reset()
	if got := blacklistKey("jti-1"); got != defaultBlacklistPrefix+"jti-1" {
		t.Errorf("未初始化时 blacklistKey() = %q", got)
	}
}
