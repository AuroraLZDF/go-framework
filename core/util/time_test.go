// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package util

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTimeMarshalJSON(t *testing.T) {
	tm := FromTime(time.Date(2026, 11, 4, 12, 0, 0, 0, time.Local))

	data, err := json.Marshal(tm)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if got := string(data); got != `"2026-11-04 12:00:00"` {
		t.Errorf("MarshalJSON() = %s, want \"2026-11-04 12:00:00\"", got)
	}
}

func TestTimeJSONRoundTrip(t *testing.T) {
	var back Time
	if err := json.Unmarshal([]byte(`"2026-11-04 12:00:00"`), &back); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	data, err := json.Marshal(back)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if got := string(data); got != `"2026-11-04 12:00:00"` {
		t.Errorf("往返后 MarshalJSON() = %s, want \"2026-11-04 12:00:00\"", got)
	}
}

func TestTimeScan(t *testing.T) {
	var tm Time

	if err := tm.Scan(nil); err != nil {
		t.Fatalf("Scan(nil) error = %v", err)
	}
	if !tm.ToTime().IsZero() {
		t.Errorf("Scan(nil) 后应为零值, got %s", tm.String())
	}

	want := time.Date(2026, 11, 4, 12, 0, 0, 0, time.Local)
	if err := tm.Scan(want); err != nil {
		t.Fatalf("Scan(time.Time) error = %v", err)
	}
	if !tm.ToTime().Equal(want) {
		t.Errorf("Scan(time.Time) = %s, want %s", tm.String(), want.Format(DefaultTimeFormat))
	}

	if err := tm.Scan([]byte("2026-11-04 12:00:00")); err != nil {
		t.Fatalf("Scan([]byte) error = %v", err)
	}
	if got := tm.String(); got != "2026-11-04 12:00:00" {
		t.Errorf("Scan([]byte) = %s, want 2026-11-04 12:00:00", got)
	}

	if err := tm.Scan([]byte("bad-time")); err == nil {
		t.Error("Scan(非法 []byte) 应返回错误")
	}
	if err := tm.Scan(42); err == nil {
		t.Error("Scan(不支持的类型) 应返回错误")
	}
}

func TestTimeValue(t *testing.T) {
	want := time.Date(2026, 11, 4, 12, 0, 0, 0, time.Local)
	tm := FromTime(want)

	v, err := tm.Value()
	if err != nil {
		t.Fatalf("Value() error = %v", err)
	}
	got, ok := v.(time.Time)
	if !ok {
		t.Fatalf("Value() 类型 = %T, want time.Time", v)
	}
	if !got.Equal(want.UTC()) {
		t.Errorf("Value() = %s, want %s（应转换为 UTC）", got.Format(time.RFC3339), want.UTC().Format(time.RFC3339))
	}
	if got.Location() != time.UTC {
		t.Error("Value() 返回的时间应为 UTC")
	}
}

func TestNowAndTimeConversions(t *testing.T) {
	now := Now()
	if now.ToTime().IsZero() {
		t.Fatal("Now() 不应返回零值")
	}
	if diff := time.Since(now.ToTime()); diff < 0 || diff > time.Minute {
		t.Errorf("Now() 与 time.Now() 偏差 %v 超出预期", diff)
	}

	std := time.Date(2026, 11, 4, 12, 0, 0, 0, time.Local)
	got := FromTime(std)
	if !got.ToTime().Equal(std) {
		t.Errorf("FromTime(%s) = %s, 应保持同一时刻", std, got.String())
	}

	var zero Time
	if !zero.ToTime().IsZero() {
		t.Error("零值 Time 转换后仍应为零值")
	}
}
