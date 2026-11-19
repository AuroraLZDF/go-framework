// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// newProjectRoot 在临时目录中构造一个最小可用的生成项目根（go.mod + 目录结构）.
func newProjectRoot(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	mod := "example.com/demo-app"

	dirs := []string{
		filepath.Join(root, "internal", "model"),
		filepath.Join(root, "internal", "store"),
		filepath.Join(root, "internal", "controller"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o750); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
	}

	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module "+mod+"\n\ngo 1.26.0\n"), 0o600); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	// GenerateModel/GenerateCRUD 使用 os.Getwd 定位项目根
	t.Chdir(root)

	return root
}

func TestStructName(t *testing.T) {
	cases := map[string]string{
		"user":         "User",
		"user_profile": "UserProfile",
		"":             "",
	}

	for name, want := range cases {
		if got := StructName(name); got != want {
			t.Errorf("StructName(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestParseFields(t *testing.T) {
	fields, err := ParseFields("name:string, user_name: text , age:int8, birth:time, price:float")
	if err != nil {
		t.Fatalf("ParseFields: %v", err)
	}

	want := []Field{
		{Name: "Name", Column: "name", Type: "string"},
		{Name: "UserName", Column: "user_name", Type: "string", GormTag: `gorm:"type:text"`},
		{Name: "Age", Column: "age", Type: "int8"},
		{Name: "Birth", Column: "birth", Type: "time.Time"},
		{Name: "Price", Column: "price", Type: "float64"},
	}
	if len(fields) != len(want) {
		t.Fatalf("got %d fields, want %d", len(fields), len(want))
	}
	for i := range want {
		if fields[i] != want[i] {
			t.Errorf("field[%d] = %+v, want %+v", i, fields[i], want[i])
		}
	}
}

func TestParseFieldsInvalid(t *testing.T) {
	cases := []string{
		"",
		",",
		"name",
		"name:",
		":string",
		"1bad:string",
		"name:unknown_type",
		"name:string,name:string",
		"Name:string",
	}

	for _, spec := range cases {
		if _, err := ParseFields(spec); err == nil {
			t.Errorf("ParseFields(%q) should fail", spec)
		}
	}
}

func TestGenerateModel(t *testing.T) {
	root := newProjectRoot(t)

	fields, err := ParseFields("name:string,birth:time")
	if err != nil {
		t.Fatalf("ParseFields: %v", err)
	}

	files, err := GenerateModel(GenerateOptions{Name: "user_profile", Fields: fields})
	if err != nil {
		t.Fatalf("GenerateModel: %v", err)
	}
	if len(files) != 1 || !strings.HasSuffix(files[0], "user_profile.go") {
		t.Errorf("unexpected file list: %v", files)
	}

	content, err := os.ReadFile(filepath.Join(root, "internal", "model", "user_profile.go"))
	if err != nil {
		t.Fatalf("read model: %v", err)
	}

	text := string(content)
	for _, want := range []string{
		"package model",
		`"gorm.io/gorm"`,
		`"time"`,
		"type UserProfile struct {",
		"gorm.Model",
		"Name string `json:\"name\"`",
		"Birth time.Time `json:\"birth\"`",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("model file missing %q:\n%s", want, text)
		}
	}
}

func TestGenerateModelDuplicate(t *testing.T) {
	newProjectRoot(t)

	fields, _ := ParseFields("name:string")
	if _, err := GenerateModel(GenerateOptions{Name: "user", Fields: fields}); err != nil {
		t.Fatalf("first GenerateModel: %v", err)
	}

	// 重复生成默认报错，--force 允许覆盖
	if _, err := GenerateModel(GenerateOptions{Name: "user", Fields: fields}); err == nil {
		t.Error("duplicate GenerateModel should fail without --force")
	}
	if _, err := GenerateModel(GenerateOptions{Name: "user", Fields: fields, Force: true}); err != nil {
		t.Fatalf("GenerateModel with --force: %v", err)
	}
}

func TestGenerateCRUD(t *testing.T) {
	root := newProjectRoot(t)

	fields, err := ParseFields("title:string,stock:int")
	if err != nil {
		t.Fatalf("ParseFields: %v", err)
	}

	files, err := GenerateCRUD(GenerateOptions{Name: "product", Fields: fields})
	if err != nil {
		t.Fatalf("GenerateCRUD: %v", err)
	}
	if len(files) != 3 {
		t.Fatalf("got %d files, want 3: %v", len(files), files)
	}

	// store：引用模块路径与框架仓储基类
	storeContent, err := os.ReadFile(filepath.Join(root, "internal", "store", "product_store.go"))
	if err != nil {
		t.Fatalf("read store: %v", err)
	}
	for _, want := range []string{
		"package store",
		"example.com/demo-app/internal/model",
		"github.com/AuroraLZDF/go-framework/core/database/store",
		"func NewProductStore(db *gorm.DB) (*ProductStore, error)",
	} {
		if !strings.Contains(string(storeContent), want) {
			t.Errorf("store file missing %q:\n%s", want, storeContent)
		}
	}

	// controller：请求体 required 约定只作用于 string 字段
	controllerContent, err := os.ReadFile(filepath.Join(root, "internal", "controller", "product_controller.go"))
	if err != nil {
		t.Fatalf("read controller: %v", err)
	}
	for _, want := range []string{
		"package controller",
		`Title string ` + "`json:\"title\" binding:\"required\"`",
		`Stock int ` + "`json:\"stock\"`",
		"func (h *ProductController) RegisterRoutes(rg *gin.RouterGroup)",
		"errno.ErrBind",
		"fwstore.DefaultPageSize",
	} {
		if !strings.Contains(string(controllerContent), want) {
			t.Errorf("controller file missing %q:\n%s", want, controllerContent)
		}
	}
}

func TestGenerateOutsideProject(t *testing.T) {
	// 没有 go.mod 的目录中执行必须报错
	newProjectRoot(t)
	if err := os.Remove(filepath.Join(".", "go.mod")); err != nil {
		t.Fatalf("remove go.mod: %v", err)
	}

	fields, _ := ParseFields("name:string")
	if _, err := GenerateCRUD(GenerateOptions{Name: "user", Fields: fields}); err == nil {
		t.Error("GenerateCRUD without go.mod should fail")
	}
}

func TestGenerateInvalidOptions(t *testing.T) {
	newProjectRoot(t)

	fields, _ := ParseFields("name:string")

	for _, opts := range []GenerateOptions{
		{Name: "Bad-Name", Fields: fields},
		{Name: "1user", Fields: fields},
		{Name: "user", Fields: nil},
	} {
		if _, err := GenerateModel(opts); err == nil {
			t.Errorf("GenerateModel %+v should fail", opts)
		}
	}
}
