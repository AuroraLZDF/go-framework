// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package version

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestGitVersionMatchesRelease(t *testing.T) {
	if GitVersion != "v0.3.0" {
		t.Errorf("GitVersion = %q, want v0.3.0", GitVersion)
	}
}

func TestBuildDateIsRFC3339(t *testing.T) {
	if _, err := time.Parse(time.RFC3339, BuildDate); err != nil {
		t.Errorf("BuildDate = %q 不是合法的 RFC3339 时间: %v", BuildDate, err)
	}
	if !strings.HasSuffix(BuildDate, "Z") {
		t.Errorf("BuildDate = %q, 应为 UTC 时间（以 Z 结尾）", BuildDate)
	}
}

func TestGetContainsGitVersion(t *testing.T) {
	info := Get()
	if info.GitVersion != GitVersion {
		t.Errorf("Get().GitVersion = %q, want %q", info.GitVersion, GitVersion)
	}
	if info.Platform == "" {
		t.Error("Get().Platform 不应为空")
	}
}

func TestInfoString(t *testing.T) {
	info := Get()
	s := info.String()

	if s == "" {
		t.Fatal("String() 不应返回空字符串")
	}
	for _, want := range []string{"gitVersion:", GitVersion, "gitCommit:", "buildDate:"} {
		if !strings.Contains(s, want) {
			t.Errorf("String() 输出缺少 %q:\n%s", want, s)
		}
	}
}

func TestInfoText(t *testing.T) {
	info := Get()
	b, err := info.Text()
	if err != nil {
		t.Fatalf("Text() error = %v", err)
	}
	if len(b) == 0 {
		t.Fatal("Text() 不应返回空内容")
	}
	if !strings.Contains(string(b), "platform:") {
		t.Errorf("Text() 输出缺少 platform:\n%s", b)
	}
}

func TestInfoToJSON(t *testing.T) {
	info := Get()
	raw := info.ToJSON()

	var parsed map[string]string
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		t.Fatalf("ToJSON() 不是合法 JSON: %v\nraw = %s", err, raw)
	}
	if parsed["gitVersion"] != GitVersion {
		t.Errorf("gitVersion = %q, want %q", parsed["gitVersion"], GitVersion)
	}
	if parsed["platform"] != info.Platform {
		t.Errorf("platform = %q, want %q", parsed["platform"], info.Platform)
	}
	if parsed["goVersion"] == "" {
		t.Error("goVersion 不应为空")
	}
}
