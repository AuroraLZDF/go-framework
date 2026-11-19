// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package util

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestParamToInt64(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/users/:id", func(c *gin.Context) {
		id, err := ParamToInt64(c, "id")
		if err != nil {
			t.Errorf("ParamToInt64() error = %v", err)
		}
		if id != 42 {
			t.Errorf("ParamToInt64() = %d, want 42", id)
		}
		c.Status(http.StatusOK)
	})
	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/users/42", http.NoBody))
}

func TestParamToInt64Invalid(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/users/:id", func(c *gin.Context) {
		if _, err := ParamToInt64(c, "id"); err == nil {
			t.Error("非法参数应返回错误")
		}
		c.Status(http.StatusOK)
	})
	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/users/abc", http.NoBody))
}

func TestParamToInt64IgnoresQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/users", func(c *gin.Context) {
		if _, err := ParamToInt64(c, "id"); err == nil {
			t.Error("不应从 query 参数中读取路径参数")
		}
		c.Status(http.StatusOK)
	})
	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/users?id=7", http.NoBody))
}

func TestAssignIfNotNil(t *testing.T) {
	dst := "old"
	src := "new"
	AssignIfNotNil(&dst, &src)
	if dst != "new" {
		t.Errorf("src 非 nil 时应完成赋值, dst = %q, want %q", dst, "new")
	}

	num := 1
	AssignIfNotNil(&num, (*int)(nil))
	if num != 1 {
		t.Errorf("src 为 nil 时不应修改目标值, num = %d, want 1", num)
	}

	type point struct{ X, Y int }
	pdst := point{X: 0, Y: 0}
	psrc := point{X: 3, Y: 4}
	AssignIfNotNil(&pdst, &psrc)
	if pdst.X != 3 || pdst.Y != 4 {
		t.Errorf("结构体类型应完成赋值, pdst = %v", pdst)
	}
}

func TestGenerateRandomString(t *testing.T) {
	if got := GenerateRandomString(0); got != "" {
		t.Errorf("长度为 0 时应返回空字符串, got %q", got)
	}
	if got := GenerateRandomString(-1); got != "" {
		t.Errorf("负长度时应返回空字符串, got %q", got)
	}

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	first := GenerateRandomString(32)
	if len(first) != 32 {
		t.Fatalf("长度 = %d, want 32", len(first))
	}
	for _, ch := range first {
		if !containsRune(charset, ch) {
			t.Fatalf("生成结果包含非法字符 %q", ch)
		}
	}

	seen := map[string]struct{}{first: {}}
	for i := 0; i < 10; i++ {
		s := GenerateRandomString(32)
		if _, ok := seen[s]; ok {
			t.Fatalf("生成了重复的随机字符串: %q", s)
		}
		seen[s] = struct{}{}
	}
}

func containsRune(s string, r rune) bool {
	for _, c := range s {
		if c == r {
			return true
		}
	}

	return false
}
