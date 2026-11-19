// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package util

import (
	"testing"
	"time"
)

func TestIsWithinDuration(t *testing.T) {
	if IsWithinDuration(time.Time{}, time.Minute) {
		t.Error("零值时间不应视为在时间段内")
	}
	if !IsWithinDuration(time.Now().Add(-30*time.Second), time.Minute) {
		t.Error("30 秒前的时间在 1 分钟内应返回 true")
	}
	if IsWithinDuration(time.Now().Add(-2*time.Minute), time.Minute) {
		t.Error("2 分钟前的时间在 1 分钟内应返回 false")
	}
}

func TestIsExpired(t *testing.T) {
	if !IsExpired(time.Time{}, time.Minute) {
		t.Error("零值时间应视为已过期")
	}
	if IsExpired(time.Now().Add(-30*time.Second), time.Minute) {
		t.Error("30 秒前的时间在 1 分钟内不应过期")
	}
	if !IsExpired(time.Now().Add(-2*time.Minute), time.Minute) {
		t.Error("2 分钟前的时间在 1 分钟内应过期")
	}
}

func TestJoinInt64s(t *testing.T) {
	tests := []struct {
		name string
		nums []int64
		want string
	}{
		{"normal", []int64{1, 2, 3}, "1,2,3"},
		{"empty", []int64{}, ""},
		{"nil", nil, ""},
		{"single", []int64{42}, "42"},
		{"negative", []int64{-1, 0}, "-1,0"},
		{"max", []int64{9223372036854775807}, "9223372036854775807"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := JoinInt64s(tt.nums); got != tt.want {
				t.Errorf("JoinInt64s(%v) = %q, want %q", tt.nums, got, tt.want)
			}
		})
	}
}
