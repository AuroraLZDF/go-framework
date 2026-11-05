// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package version

import (
	"strings"
	"testing"
	"time"
)

func TestGitVersionIsV013(t *testing.T) {
	if GitVersion != "v0.2.0" {
		t.Errorf("GitVersion = %q, want v0.2.0", GitVersion)
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
