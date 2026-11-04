// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package validator

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestFieldNamePattern(t *testing.T) {
	valid := []string{"id", "user_id", "_private", "created_at", "Name1"}
	for _, field := range valid {
		if !fieldNamePattern.MatchString(field) {
			t.Errorf("字段名 %q 应被允许", field)
		}
	}

	invalid := []string{
		"",
		"1abc",
		"user-id",
		"users.id",
		"id; DROP TABLE users",
		"id) OR 1=1",
		"name`",
		"name space",
	}
	for _, field := range invalid {
		if fieldNamePattern.MatchString(field) {
			t.Errorf("字段名 %q 应被拒绝", field)
		}
	}
}

func TestEscapeLikeValue(t *testing.T) {
	testCases := map[string]string{
		"normal":     "normal",
		"50%":        `50\%`,
		"a_b":        `a\_b`,
		`back\slash`: `back\\slash`,
	}
	for input, want := range testCases {
		if got := escapeLikeValue(input); got != want {
			t.Errorf("escapeLikeValue(%q) = %q, want %q", input, got, want)
		}
	}
}

// noopConnPool 是一个不会真正连接数据库的连接池桩，配合 DryRun 使用.
type noopConnPool struct{}

var errNoopConnPool = errors.New("noop conn pool")

func (noopConnPool) PrepareContext(context.Context, string) (*sql.Stmt, error) {
	return nil, errNoopConnPool
}

func (noopConnPool) ExecContext(context.Context, string, ...interface{}) (sql.Result, error) {
	return nil, errNoopConnPool
}

func (noopConnPool) QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error) {
	return nil, errNoopConnPool
}

func (noopConnPool) QueryRowContext(context.Context, string, ...interface{}) *sql.Row {
	return &sql.Row{}
}

// newDryRunDB 构造一个不实际连接数据库的 *gorm.DB，用于检查生成的 SQL/错误.
func newDryRunDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      noopConnPool{},
		SkipInitializeWithVersion: true,
	}), &gorm.Config{DryRun: true})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}

	return db
}

func TestWithFiltersEscapesLikeWildcards(t *testing.T) {
	db := newDryRunDB(t)
	tx := db.Session(&gorm.Session{DryRun: true}).
		Scopes(WithFilters(Filter{Field: "name", Value: "50%_off", Operator: OpLike})).
		Table("users")

	var rows []map[string]interface{}
	if err := tx.Find(&rows).Error; err != nil {
		t.Fatalf("Find() error = %v", err)
	}

	if len(tx.Statement.Vars) != 1 {
		t.Fatalf("Vars = %v, want 1 个 LIKE 参数", tx.Statement.Vars)
	}
	pattern, ok := tx.Statement.Vars[0].(string)
	if !ok {
		t.Fatalf("LIKE 参数类型 = %T, want string", tx.Statement.Vars[0])
	}
	want := `%50\%\_off%`
	if pattern != want {
		t.Errorf("LIKE 参数 = %q, want %q（用户输入的 %%/_ 应被转义）", pattern, want)
	}
}

func TestWithFiltersRejectsQualifiedField(t *testing.T) {
	db := newDryRunDB(t)
	tx := db.Session(&gorm.Session{DryRun: true}).
		Scopes(WithFilters(Filter{Field: "users.id", Value: int64(1), Operator: OpEqual})).
		Table("users")

	var rows []map[string]interface{}
	if err := tx.Find(&rows).Error; err != nil {
		t.Fatalf("Find() error = %v", err)
	}
	if sql := tx.Statement.SQL.String(); !strings.Contains(sql, "1 = 0") {
		t.Errorf("限定字段名应 fail-closed，SQL = %q", sql)
	}
}

func TestWithFiltersUnknownOperatorReturnsError(t *testing.T) {
	db := newDryRunDB(t)
	tx := db.Session(&gorm.Session{DryRun: true}).
		Scopes(WithFilters(Filter{Field: "id", Value: int64(1), Operator: FilterOperator("ILLEGAL")})).
		Table("users")

	var rows []map[string]interface{}
	err := tx.Find(&rows).Error
	if err == nil {
		t.Fatal("未知操作符应返回错误，而不是静默按 = 处理")
	}
	if !strings.Contains(fmt.Sprint(err), "unsupported filter operator") {
		t.Errorf("错误信息 = %v, 应包含 unsupported filter operator", err)
	}
	if sql := tx.Statement.SQL.String(); sql != "" && !strings.Contains(sql, "1 = 0") {
		t.Errorf("未知操作符应 fail-closed，SQL = %q", sql)
	}
}
