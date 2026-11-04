// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package generator

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateProjectName(t *testing.T) {
	valid := []string{"my-app", "app_1", "a", "MyApp"}
	for _, name := range valid {
		if err := validateProjectName(name); err != nil {
			t.Errorf("validateProjectName(%q) = %v, want nil", name, err)
		}
	}

	invalid := []string{"", "-app", "1app", "my app", "app!", "a/b", "../x", "app.dir"}
	for _, name := range invalid {
		if err := validateProjectName(name); err == nil {
			t.Errorf("validateProjectName(%q) = nil, want error", name)
		}
	}
}

func TestNewProjectGeneratesFiles(t *testing.T) {
	t.Chdir(t.TempDir())

	if err := NewProject("demo-app"); err != nil {
		t.Fatalf("NewProject() error = %v", err)
	}

	gomod, err := os.ReadFile(filepath.Join("demo-app", "go.mod"))
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	for _, want := range []string{
		"github.com/AuroraLZDF/go-framework v0.1.3",
		"gorm.io/gorm",
		"go 1.25.1",
	} {
		if !strings.Contains(string(gomod), want) {
			t.Errorf("go.mod 缺少 %q:\n%s", want, gomod)
		}
	}

	mainFile := filepath.Join("demo-app", "cmd", "app", "main.go")
	if _, err := os.Stat(mainFile); err != nil {
		t.Errorf("main.go 未生成: %v", err)
	}

	for _, name := range []string{"config.yaml", "config.example.yaml"} {
		path := filepath.Join("demo-app", name)
		content, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("读取 %s 失败: %v", name, err)
			continue
		}
		for _, want := range []string{"shutdown-timeout", "pool-size", "至少 32 字节"} {
			if !strings.Contains(string(content), want) {
				t.Errorf("%s 缺少 %q", name, want)
			}
		}
	}

	// config.yaml 与 config.example.yaml 内容应一致
	actual, err := os.ReadFile(filepath.Join("demo-app", "config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	example, err := os.ReadFile(filepath.Join("demo-app", "config.example.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if string(actual) != string(example) {
		t.Error("config.yaml 与 config.example.yaml 内容不一致")
	}
}

func TestNewProjectRejectsExistingDirectory(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.Mkdir("exists-app", 0755); err != nil {
		t.Fatal(err)
	}

	if err := NewProject("exists-app"); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("NewProject(已存在目录) error = %v, want 'already exists'", err)
	}
}

func TestNewProjectRollbackOnFailure(t *testing.T) {
	t.Chdir(t.TempDir())

	prevSteps := projectSteps
	projectSteps = append(append([]generateStep{}, prevSteps...), generateStep{
		name: "always-fail",
		run:  func(string) error { return errors.New("injected failure") },
	})
	t.Cleanup(func() { projectSteps = prevSteps })

	if err := NewProject("rollback-app"); err == nil {
		t.Fatal("NewProject() 应返回错误")
	}

	if _, err := os.Stat("rollback-app"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("失败回滚后目录仍存在, stat err = %v", err)
	}
}
