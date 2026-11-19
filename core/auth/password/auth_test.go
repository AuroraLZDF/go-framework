// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package auth

import (
	"errors"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestEncryptAndCompare(t *testing.T) {
	const password = "s3cr3t-Passw0rd!"

	hashed, err := Encrypt(password)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}
	if hashed == "" {
		t.Fatal("Encrypt() 返回的密文不应为空")
	}

	// 正确密码应校验通过
	if err := Compare(hashed, password); err != nil {
		t.Errorf("Compare(正确密码) error = %v, want nil", err)
	}

	// 错误密码应校验失败
	if err := Compare(hashed, "wrong-password"); err == nil {
		t.Error("Compare(错误密码) 应返回错误")
	} else if !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		t.Errorf("Compare(错误密码) error = %v, want ErrMismatchedHashAndPassword", err)
	}

	// 相同明文每次加密应产生不同盐值
	hashed2, err := Encrypt(password)
	if err != nil {
		t.Fatalf("第二次 Encrypt() error = %v", err)
	}
	if hashed == hashed2 {
		t.Error("两次 Encrypt 相同明文应产生不同密文（随机盐）")
	}
}

func TestEncryptUsesDefaultCost(t *testing.T) {
	hashed, err := Encrypt("cost-check")
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	if !strings.HasPrefix(hashed, "$2a$") && !strings.HasPrefix(hashed, "$2b$") {
		t.Fatalf("密文应携带 bcrypt 版本前缀, got %s", hashed)
	}

	cost, err := bcrypt.Cost([]byte(hashed))
	if err != nil {
		t.Fatalf("bcrypt.Cost() error = %v", err)
	}
	if cost != bcrypt.DefaultCost {
		t.Errorf("加密成本 = %d, want bcrypt.DefaultCost(%d)", cost, bcrypt.DefaultCost)
	}
}

func TestEncryptLengthBoundaries(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		wantErr bool
	}{
		{"空字符串", "", false},
		{"72 字节边界", strings.Repeat("a", 72), false},
		// bcrypt 最大支持 72 字节，超过时应返回错误而不是静默截断
		{"73 字节超长", strings.Repeat("a", 73), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashed, err := Encrypt(tt.source)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Encrypt(超长密码) 应返回错误")
				}
				if !errors.Is(err, bcrypt.ErrPasswordTooLong) {
					t.Errorf("Encrypt(超长密码) error = %v, want ErrPasswordTooLong", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("Encrypt() error = %v", err)
			}

			// 边界内的密码应能通过 Compare 校验
			if err := Compare(hashed, tt.source); err != nil {
				t.Errorf("Compare(原密码) error = %v", err)
			}
		})
	}
}

func TestCompareWithInvalidHash(t *testing.T) {
	tests := []struct {
		name   string
		hashed string
	}{
		{"空密文", ""},
		{"非 bcrypt 格式", "not-a-bcrypt-hash"},
		{"错误前缀", "$3a$10$someinvalidhashpayload00000000000000000000000000000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Compare(tt.hashed, "any-password"); err == nil {
				t.Errorf("%s: Compare 应返回错误", tt.name)
			}
		})
	}
}
