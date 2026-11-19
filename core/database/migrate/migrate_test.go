// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package migrate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCreate 生成迁移文件的基本行为：文件存在、命名正确、内容含注释头.
func TestCreate(t *testing.T) {
	dir := t.TempDir()

	up, err := Create(dir, "create_users_table")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if !strings.HasSuffix(up, ".up.sql") {
		t.Errorf("up path = %s, want suffix .up.sql", up)
	}

	matches, err := filepath.Glob(filepath.Join(dir, "*_create_users_table.*.sql"))
	if err != nil {
		t.Fatalf("Glob failed: %v", err)
	}
	if len(matches) != 2 {
		t.Fatalf("got %d migration files, want 2 (up+down): %v", len(matches), matches)
	}

	// up/down 应共享同一个时间戳版本前缀
	baseUp := filepath.Base(up)
	version := strings.SplitN(baseUp, "_", 2)[0]
	if len(version) != 14 {
		t.Errorf("version %q length = %d, want 14", version, len(version))
	}

	downPath := filepath.Join(dir, version+"_create_users_table.down.sql")
	if _, err := os.Stat(downPath); err != nil {
		t.Errorf("down migration %s missing: %v", downPath, err)
	}

	content, err := os.ReadFile(up)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if !strings.HasPrefix(string(content), "-- Migration: ") {
		t.Errorf("migration content should start with a comment header, got: %q", string(content))
	}
}

// TestCreateInvalidName 非法名称必须被拒绝.
func TestCreateInvalidName(t *testing.T) {
	for _, name := range []string{"", "_bad", "has space", "1+1", "drop;table"} {
		if _, err := Create(t.TempDir(), name); err == nil {
			t.Errorf("Create with name %q should fail", name)
		}
	}
}

// TestCreateDuplicateDirDefaultDir 空目录名回退到默认值，重复创建同名迁移必须报错.
func TestCreateDuplicateDirDefaultDir(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	if _, err := Create("", "seed_roles"); err != nil {
		t.Fatalf("Create with empty dir failed: %v", err)
	}

	// 时间戳精确到秒，同秒内重复创建必然同名，应报错
	if _, err := Create("", "seed_roles"); err == nil {
		t.Error("Create duplicate migration should fail")
	} else if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestRunValidation 未配置 DSN 或迁移目录缺失时，执行类操作应快速失败.
func TestRunValidation(t *testing.T) {
	opts := &Options{Dir: t.TempDir(), DSN: ""}
	if err := Up(opts); err == nil {
		t.Error("Up without DSN should fail")
	}
	if _, _, err := Version(opts); err == nil {
		t.Error("Version without DSN should fail")
	}

	opts = &Options{Dir: filepath.Join(t.TempDir(), "missing"), DSN: "user:pass@tcp(127.0.0.1:3306)/db"}
	if err := Up(opts); err == nil {
		t.Error("Up with missing migrations dir should fail")
	}
	if !strings.Contains(err2String(opts), "does not exist") {
		// Down 走同样的校验路径
		t.Error("Down with missing migrations dir should fail with 'does not exist'")
	}
}

func err2String(opts *Options) string {
	err := Down(opts, 1)
	if err == nil {
		return ""
	}

	return err.Error()
}
