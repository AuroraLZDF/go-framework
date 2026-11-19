// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package verflag

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/AuroraLZDF/go-framework/core/version"
)

// TestVersionValueExported 确认 VersionValue 类型对外可用（v1.0 API 冻结的一部分）.
// 在旧版本中该类型未导出，外部无法调用 Version/VersionVar，属于无效 API.
func TestVersionValueExported(t *testing.T) {
	v := VersionFalse
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

// verflagExitEnv 标记测试二进制以子进程模式运行 PrintAndExitIfRequested，
// 因为该函数会调用 os.Exit，必须在子进程中验证.
const verflagExitEnv = "GO_FRAMEWORK_TEST_VERFLAG_EXIT"

func TestPrintAndExitIfRequestedNoopWhenFalse(t *testing.T) {
	*versionFlag = VersionFalse
	t.Cleanup(func() { *versionFlag = VersionFalse })

	// 未请求 --version 时不应退出进程；若函数误调用 os.Exit，测试会直接失败
	PrintAndExitIfRequested()
}

func TestPrintAndExitIfRequestedExitsInSubprocess(t *testing.T) {
	*versionFlag = VersionFalse
	t.Cleanup(func() { *versionFlag = VersionFalse })

	tests := []struct {
		name string
		mode string
	}{
		{"version=true", "true"},
		{"version=raw", "raw"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestPrintAndExitIfRequestedSubprocess$", "-test.v")
			cmd.Env = append(os.Environ(),
				verflagExitEnv+"=1",
				"GO_FRAMEWORK_TEST_VERFLAG_MODE="+tt.mode,
			)
			out, err := cmd.CombinedOutput()

			// os.Exit(0) 属于正常退出（err = nil）；若函数未退出，子进程的
			// t.Fatal 会以 exit code 1 结束（err 非 nil）
			if err != nil {
				t.Fatalf("子进程应以 exit code 0 结束, err = %v, output = %s", err, out)
			}
			if strings.Contains(string(out), "未按预期退出进程") {
				t.Errorf("PrintAndExitIfRequested 应直接退出进程, output =\n%s", out)
			}
			// 子进程在 os.Exit 前打印版本信息，输出应包含版本号
			if !strings.Contains(string(out), version.GitVersion) {
				t.Errorf("输出应包含版本号 %s, output =\n%s", version.GitVersion, out)
			}
		})
	}
}

// TestPrintAndExitIfRequestedSubprocess 仅作为子进程运行，
// 由 TestPrintAndExitIfRequestedExitsInSubprocess 通过环境变量触发.
func TestPrintAndExitIfRequestedSubprocess(t *testing.T) {
	if os.Getenv(verflagExitEnv) != "1" {
		t.Skip("仅在子进程模式下运行")
	}

	switch os.Getenv("GO_FRAMEWORK_TEST_VERFLAG_MODE") {
	case "raw":
		*versionFlag = VersionRaw
	case "true":
		*versionFlag = VersionTrue
	default:
		return
	}

	PrintAndExitIfRequested()

	// VersionTrue/VersionRaw 必须直接退出进程，走到这里说明行为不符
	t.Fatal("PrintAndExitIfRequested 未按预期退出进程")
}
