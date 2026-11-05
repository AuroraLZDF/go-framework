// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package verflag

import "testing"

// TestVersionValueExported 确认 VersionValue 类型对外可用（v1.0 API 冻结的一部分）.
// 在旧版本中该类型未导出，外部无法调用 Version/VersionVar，属于无效 API.
func TestVersionValueExported(t *testing.T) {
	var v VersionValue = VersionFalse
	if v.String() != "false" {
		t.Fatalf("VersionFalse.String() = %q, want %q", v.String(), "false")
	}
	raw := VersionRaw
	if raw.String() != "raw" {
		t.Fatalf("VersionRaw.String() = %q, want %q", raw.String(), "raw")
	}
	if !v.IsBoolFlag() {
		t.Fatal("VersionValue should implement pflag bool flag")
	}
	if err := v.Set("true"); err != nil {
		t.Fatalf("VersionValue.Set(true) returned error: %v", err)
	}
	if v != VersionTrue {
		t.Fatalf("after Set(true), v = %v, want VersionTrue", v)
	}
	if got := v.Get(); got != &v {
		t.Fatalf("Get() = %v, want pointer to v", got)
	}
	if v.Type() != "version" {
		t.Fatalf("Type() = %q, want %q", v.Type(), "version")
	}
}
