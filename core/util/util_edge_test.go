// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package util

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestStructToMapSkipsUnexportedFields(t *testing.T) {
	type sample struct {
		Name   string
		Age    int
		hidden string // 未导出字段：不能 panic，也不应出现在结果中
	}

	got := StructToMap(sample{Name: "alice", Age: 18, hidden: "secret"})
	if got["name"] != "alice" {
		t.Errorf("name = %q, want alice", got["name"])
	}
	if got["age"] != "18" {
		t.Errorf("age = %q, want 18", got["age"])
	}
	if _, ok := got["hidden"]; ok {
		t.Error("未导出字段不应出现在结果中")
	}
}

func TestStructToMapNilAndNonStruct(t *testing.T) {
	if got := StructToMap(nil); got != nil {
		t.Errorf("StructToMap(nil) = %v, want nil", got)
	}
	if got := StructToMap(42); got != nil {
		t.Errorf("StructToMap(42) = %v, want nil", got)
	}

	var p *struct{ Name string }
	if got := StructToMap(p); got != nil {
		t.Errorf("StructToMap(nil pointer) = %v, want nil", got)
	}

	type sample struct{ Name string }
	if got := StructToMap(&sample{Name: "bob"}); got["name"] != "bob" {
		t.Errorf("StructToMap(*struct) name = %q, want bob", got["name"])
	}
}

func TestSnakeCaseAcronyms(t *testing.T) {
	type sample struct {
		UserID     int
		HTTPServer string
		ID         int
	}

	got := StructToMap(sample{})
	for _, want := range []string{"user_id", "http_server", "id"} {
		if _, ok := got[want]; !ok {
			t.Errorf("StructToMap 缺少 key %q, got %v", want, got)
		}
	}
}

func TestAddFileIfNotExistsDoesNotTruncate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "a.txt")

	if err := AddFileIfNotExists(path); err != nil {
		t.Fatalf("首次 AddFileIfNotExists() error = %v", err)
	}
	if err := os.WriteFile(path, []byte("keep-me"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := AddFileIfNotExists(path); err != nil {
		t.Fatalf("再次 AddFileIfNotExists() error = %v", err)
	}
	if err := AddFile(path); err != nil {
		t.Fatalf("AddFile(已存在) error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "keep-me" {
		t.Errorf("已有文件被截断/覆盖, content = %q", data)
	}
}

func TestAddDirCreatesDirectoryItself(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "x", "y")
	if err := AddDir(dir); err != nil {
		t.Fatalf("AddDir() error = %v", err)
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("stat %q: %v", dir, err)
	}
	if !info.IsDir() {
		t.Fatalf("%q 不是目录", dir)
	}
}

func TestAddFileIfNotExistsOnDirectory(t *testing.T) {
	dir := t.TempDir()
	if err := AddFileIfNotExists(dir); err == nil {
		t.Error("path 是目录时应返回错误")
	}
}

func TestGetRealClientIPFallsBackWhenHeaderInvalid(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "203.0.113.9:5555"
	r.Header.Set("X-Real-IP", "not-an-ip")

	if got := GetRealClientIP(r); got != "203.0.113.9" {
		t.Errorf("GetRealClientIP() = %q, want fallback 203.0.113.9", got)
	}
}

func TestGetRealClientIPSkipsInvalidEntries(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "203.0.113.9:5555"
	r.Header.Set("X-Forwarded-For", "invalid, 8.8.8.8")

	if got := GetRealClientIP(r); got != "8.8.8.8" {
		t.Errorf("GetRealClientIP() = %q, want 8.8.8.8", got)
	}
}

func TestGetRealClientIPPrefersHeader(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "203.0.113.9:5555"
	r.Header.Set("X-Real-IP", "198.51.100.7, 10.0.0.1")

	if got := GetRealClientIP(r); got != "198.51.100.7" {
		t.Errorf("GetRealClientIP() = %q, want 198.51.100.7", got)
	}
}

func TestGetRealClientIPRemoteAddrWithoutPort(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "192.0.2.5"

	if got := GetRealClientIP(r); got != "192.0.2.5" {
		t.Errorf("GetRealClientIP() = %q, want 192.0.2.5", got)
	}

	if got := remoteIP("[2001:db8::1]:443"); got != "2001:db8::1" {
		t.Errorf("remoteIP(ipv6) = %q, want 2001:db8::1", got)
	}
}

func TestTimeUnmarshalJSONReturnsError(t *testing.T) {
	var tm Time

	for _, invalid := range []string{"123", `"bad-time"`, `"2026/11/04"`} {
		if err := tm.UnmarshalJSON([]byte(invalid)); err == nil {
			t.Errorf("UnmarshalJSON(%s) 应返回错误", invalid)
		}
	}
	if err := tm.UnmarshalJSON(nil); err == nil {
		t.Error("UnmarshalJSON(nil) 应返回错误")
	}

	if err := tm.UnmarshalJSON([]byte(`"2026-11-04 12:00:00"`)); err != nil {
		t.Fatalf("合法时间解析失败: %v", err)
	}
	if got := tm.String(); got != "2026-11-04 12:00:00" {
		t.Errorf("String() = %q, want 2026-11-04 12:00:00", got)
	}
}

func TestTreeServiceRepeatedCallsAndNoMutation(t *testing.T) {
	data := []map[string]interface{}{
		{"id": 1, "parent_id": 0, "name": "root"},
		{"id": 2, "parent_id": 1, "name": "a"},
		{"id": 3, "parent_id": 2, "name": "a1"},
		{"id": 4, "parent_id": 1, "name": "b"},
	}

	ts := NewTreeService(TreeService{PrimaryKey: "id", ParentKey: "parent_id"})

	result := ts.MakeTree(data)
	if len(result) != 1 {
		t.Fatalf("MakeTree() 根节点数 = %d, want 1", len(result))
	}

	children, ok := result[0]["children"].([]map[string]interface{})
	if !ok || len(children) != 2 {
		t.Fatalf("MakeTree() children = %v, want 2 个子节点", result[0]["children"])
	}

	// 入参 map 不应被注入 children/leaf/expanded 等字段
	for _, item := range data {
		for _, key := range []string{"children", "leaf", "expanded", "level"} {
			if _, exists := item[key]; exists {
				t.Errorf("输入数据被修改，item %v 出现 key %q", item["id"], key)
			}
		}
	}

	// 重复调用不重复追加
	again := ts.MakeTree(data)
	if len(again) != 1 {
		t.Errorf("重复 MakeTree() 根节点数 = %d, want 1", len(again))
	}

	// 线性结构：重复调用结果数量稳定、且不修改入参
	linear := ts.MakeTreeForHTML(data)
	if len(linear) != len(data) {
		t.Fatalf("MakeTreeForHTML() 节点数 = %d, want %d", len(linear), len(data))
	}
	linearAgain := ts.MakeTreeForHTML(data)
	if len(linearAgain) != len(data) {
		t.Errorf("重复 MakeTreeForHTML() 节点数 = %d, want %d", len(linearAgain), len(data))
	}
	for _, item := range data {
		if _, exists := item["level"]; exists {
			t.Errorf("线性结构修改了输入数据: item %v", item["id"])
		}
	}
}

func TestTreeServiceUsesConfiguredKeys(t *testing.T) {
	ts := NewTreeService(TreeService{
		PrimaryKey:  "id",
		ParentKey:   "pid",
		ExpandedKey: "open",
		ChildrenKey: "items",
		LeafKey:     "is_leaf",
		Expanded:    true,
	})

	data := []map[string]interface{}{
		{"id": 1, "pid": 0},
		{"id": 2, "pid": 1},
	}

	result := ts.MakeTree(data)
	root := result[0]
	if root["open"] != true {
		t.Errorf("root[open] = %v, want true", root["open"])
	}
	if _, ok := root["items"]; !ok {
		t.Error("应使用配置的 ChildrenKey")
	}

	leaf := root["items"].([]map[string]interface{})[0]
	if leaf["is_leaf"] != true {
		t.Errorf("leaf[is_leaf] = %v, want true", leaf["is_leaf"])
	}
}
