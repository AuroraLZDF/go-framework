// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	jwtpkg "github.com/AuroraLZDF/go-framework/core/auth/jwt"
)

func TestAuthnReturnsUnauthorizedForMissingToken(t *testing.T) {
	w := serve(t, Authn(), http.MethodGet, "/test", "")

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	var body struct {
		Code int `json:"code"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
	if body.Code != http.StatusUnauthorized {
		t.Errorf("body.code = %d, want %d", body.Code, http.StatusUnauthorized)
	}
}

func TestAuthnAllowsValidToken(t *testing.T) {
	if err := jwtpkg.Init(&jwtpkg.Options{
		TokenId: "uid",
		Secret:  strings.Repeat("a", jwtpkg.MinSecretLength),
		Expire:  1,
	}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	t.Cleanup(jwtpkg.Reset)

	resp, err := jwtpkg.Sign("42")
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	r := gin.New()
	r.Use(Authn())
	r.GET("/test", func(c *gin.Context) {
		uid, _ := c.Get("X-Username")
		c.JSON(http.StatusOK, gin.H{"uid": uid})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+resp.AccessToken)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestAuthnMapsExpiredToken(t *testing.T) {
	secret := strings.Repeat("b", jwtpkg.MinSecretLength)
	if err := jwtpkg.Init(&jwtpkg.Options{TokenId: "uid", Secret: secret, Expire: 1}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	t.Cleanup(jwtpkg.Reset)

	now := time.Now()
	claims := jwt.MapClaims{
		"uid": "42",
		"jti": "jti-1",
		"nbf": now.Add(-2 * time.Hour).Unix(),
		"iat": now.Add(-2 * time.Hour).Unix(),
		"exp": now.Add(-time.Hour).Unix(),
	}
	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("签发过期 token 失败: %v", err)
	}

	r := gin.New()
	r.Use(Authn())
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	var body struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
	if body.Code != http.StatusUnauthorized {
		t.Errorf("body.code = %d, want %d", body.Code, http.StatusUnauthorized)
	}
	// 过期错误应被单独映射，客户端据此触发 refresh 而不是重新登录
	if body.Message != "Token expired." {
		t.Errorf("body.message = %q, want %q", body.Message, "Token expired.")
	}
}
