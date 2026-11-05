// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package errno

import (
	"errors"
	"fmt"
	"testing"
)

func TestSetMessageReturnsCopy(t *testing.T) {
	original := ErrTokenInvalid.Message

	custom := ErrTokenInvalid.SetMessage("custom %s", "message")
	if custom == ErrTokenInvalid {
		t.Fatal("SetMessage 应返回副本，而不是修改原对象")
	}
	if custom.Message != "custom message" {
		t.Errorf("Message = %q, want %q", custom.Message, "custom message")
	}
	if custom.Code != ErrTokenInvalid.Code || custom.HTTP != ErrTokenInvalid.HTTP {
		t.Errorf("副本的 Code/HTTP 应保持不变, got (%d, %d)", custom.HTTP, custom.Code)
	}
	if ErrTokenInvalid.Message != original {
		t.Fatalf("全局错误码被修改: %q", ErrTokenInvalid.Message)
	}
}

func TestDecode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantHTTP int
		wantCode int
		wantMsg  string
	}{
		{
			name:     "nil error",
			err:      nil,
			wantHTTP: OK.HTTP,
			wantCode: OK.Code,
			wantMsg:  OK.Message,
		},
		{
			name:     "errno",
			err:      ErrTokenInvalid,
			wantHTTP: ErrTokenInvalid.HTTP,
			wantCode: ErrTokenInvalid.Code,
			wantMsg:  ErrTokenInvalid.Message,
		},
		{
			name:     "wrapped errno",
			err:      fmt.Errorf("wrap: %w", ErrTokenExpired),
			wantHTTP: ErrTokenExpired.HTTP,
			wantCode: ErrTokenExpired.Code,
			wantMsg:  ErrTokenExpired.Message,
		},
		{
			name:     "unknown error is not leaked",
			err:      errors.New("boom: password=secret"),
			wantHTTP: InternalServerError.HTTP,
			wantCode: InternalServerError.Code,
			wantMsg:  InternalServerError.Message,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpCode, code, msg := Decode(tt.err)
			if httpCode != tt.wantHTTP {
				t.Errorf("HTTP = %d, want %d", httpCode, tt.wantHTTP)
			}
			if code != tt.wantCode {
				t.Errorf("Code = %d, want %d", code, tt.wantCode)
			}
			if msg != tt.wantMsg {
				t.Errorf("Message = %q, want %q", msg, tt.wantMsg)
			}
		})
	}
}

func TestOKCode(t *testing.T) {
	if OK.Code != 200 {
		t.Errorf("OK.Code = %d, want 200", OK.Code)
	}
}
