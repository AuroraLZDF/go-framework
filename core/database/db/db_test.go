// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package db

import (
	"errors"
	"net/url"
	"strings"
	"testing"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

func TestMySQLDSNIncludesTimeoutsAndUTC(t *testing.T) {
	opts := &MySQLOptions{
		Username: "root",
		Password: "pwd",
		Host:     "127.0.0.1:3306",
		Database: "demo",
	}

	dsn, err := opts.DSN()
	if err != nil {
		t.Fatalf("DSN() error = %v", err)
	}

	for _, want := range []string{"loc=UTC", "timeout=", "readTimeout=", "writeTimeout=", "parseTime=true"} {
		if !strings.Contains(dsn, want) {
			t.Errorf("DSN() 应包含 %q, got %s", want, dsn)
		}
	}
}

func TestMySQLDSNEscapesDatabaseName(t *testing.T) {
	const dbName = "weird/db?name&x=1"
	opts := &MySQLOptions{Username: "root", Host: "127.0.0.1:3306", Database: dbName}

	dsn, err := opts.DSN()
	if err != nil {
		t.Fatalf("DSN() error = %v", err)
	}

	// 原始字符会破坏 DSN 结构，必须被路径转义
	if !strings.Contains(dsn, url.PathEscape(dbName)) {
		t.Fatalf("DSN() 未转义 database name, got %s", dsn)
	}

	// 驱动解析后应还原为原始库名
	parsed, err := mysql.ParseDSN(dsn)
	if err != nil {
		t.Fatalf("driver ParseDSN() error = %v", err)
	}
	if parsed.DBName != dbName {
		t.Errorf("解析后的 DBName = %q, want %q", parsed.DBName, dbName)
	}
}

func TestMySQLOptionsValidation(t *testing.T) {
	tests := []struct {
		name string
		opts *MySQLOptions
	}{
		{"nil options", nil},
		{"empty host", &MySQLOptions{}},
		{"username with colon", &MySQLOptions{Host: "127.0.0.1:3306", Username: "bad:user"}},
		{"username with control char", &MySQLOptions{Host: "127.0.0.1:3306", Username: "bad\nuser"}},
		{"password with control char", &MySQLOptions{Host: "127.0.0.1:3306", Password: "pwd\r"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := tt.opts.DSN(); err == nil {
				t.Fatalf("%s: DSN() 应返回错误", tt.name)
			}
		})
	}
}

func TestInitMySQLNilOptions(t *testing.T) {
	db, err := InitMySQL(nil)
	if err == nil {
		t.Fatal("InitMySQL(nil) 应返回错误")
	}
	if db != nil {
		t.Errorf("InitMySQL(nil) 返回的 db = %v, want nil", db)
	}
}

func TestInitMySQLAlreadyInitialized(t *testing.T) {
	// 直接放置一个占位单例，验证重复初始化会返回明确错误（避免依赖真实 MySQL）
	mysqlMu.Lock()
	prev := mysqlDB
	mysqlDB = &gorm.DB{}
	mysqlMu.Unlock()
	t.Cleanup(func() {
		mysqlMu.Lock()
		mysqlDB = prev
		mysqlMu.Unlock()
	})

	db, err := InitMySQL(&MySQLOptions{Host: "127.0.0.1:3306", Username: "root"})
	if err == nil {
		t.Fatal("重复 InitMySQL 应返回错误")
	}
	if !errors.Is(err, ErrDBAlreadyInitialized) {
		t.Errorf("error = %v, want errors.Is(ErrDBAlreadyInitialized)", err)
	}
	if db != nil {
		t.Errorf("重复 InitMySQL 返回的 db = %v, want nil", db)
	}
}

func TestCloseWithoutInitialization(t *testing.T) {
	// 未初始化时应可重复调用且不报错
	for i := 0; i < 2; i++ {
		if err := Close(); err != nil {
			t.Fatalf("第 %d 次 Close() error = %v", i+1, err)
		}
	}

	if GetDB() != nil {
		t.Error("Close() 后 GetDB() 应返回 nil")
	}
	if GetRedis() != nil {
		t.Error("Close() 后 GetRedis() 应返回 nil")
	}
}

func TestInitRedisNilOptions(t *testing.T) {
	client, err := InitRedis(nil)
	if err == nil {
		t.Fatal("InitRedis(nil) 应返回错误")
	}
	if client != nil {
		t.Errorf("InitRedis(nil) 返回的 client = %v, want nil", client)
	}
}
