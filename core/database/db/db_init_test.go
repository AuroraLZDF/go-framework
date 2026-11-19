// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package db

import (
	"strings"
	"testing"
	"time"

	"database/sql"

	"gorm.io/gorm"
)

// resetMySQLSingleton 将 MySQL 单例置为 nil 并在测试结束后恢复原值，避免用例间全局状态污染.
func resetMySQLSingleton(t *testing.T) {
	t.Helper()

	mysqlMu.Lock()
	prev := mysqlDB
	mysqlDB = nil
	mysqlMu.Unlock()

	t.Cleanup(func() {
		mysqlMu.Lock()
		mysqlDB = prev
		mysqlMu.Unlock()
	})
}

func TestInitMySQLInvalidOptions(t *testing.T) {
	resetMySQLSingleton(t)

	opts := &MySQLOptions{Host: "bad\nhost", Username: "root", Database: "demo"}

	got, err := InitMySQL(opts)
	if err == nil {
		t.Fatal("非法配置的 InitMySQL 应返回错误")
	}
	if got != nil {
		t.Errorf("InitMySQL 失败时返回的 db = %v, want nil", got)
	}
	if !strings.Contains(err.Error(), "invalid mysql options") {
		t.Errorf("error = %v, 应包装 DSN 校验错误", err)
	}

	// 失败不应缓存单例，允许调用方修正配置后重试
	if db := GetDB(); db != nil {
		t.Error("初始化失败不应缓存单例")
	}
}

func TestInitMySQLUnreachableHost(t *testing.T) {
	resetMySQLSingleton(t)

	// 指向本机未监听端口，连接会被立即拒绝，用于触发 gorm.Open 的错误分支（无需真实 MySQL）
	opts := &MySQLOptions{
		Host:                  "127.0.0.1:1",
		Username:              "root",
		Password:              "pwd",
		Database:              "demo",
		MaxConnectionLifeTime: 10 * time.Minute,
	}

	got, err := InitMySQL(opts)
	if err == nil {
		t.Fatal("连接不可达的 MySQL 应返回错误")
	}
	if got != nil {
		t.Errorf("InitMySQL 失败时返回的 db = %v, want nil", got)
	}
	if db := GetDB(); db != nil {
		t.Error("初始化失败不应缓存单例")
	}
}

func TestInitMySQLPlaceholderSingleton(t *testing.T) {
	// 用无效配置 + 占位单例覆盖连接池默认值的填充逻辑（在 DSN 校验前执行），
	// 实际连接通过非法 host 快速失败，不依赖真实 MySQL.
	resetMySQLSingleton(t)

	opts := &MySQLOptions{Host: "bad\nhost"} // MaxOpen/MaxIdle/LifeTime 均为零值
	if _, err := InitMySQL(opts); err == nil {
		t.Fatal("非法配置的 InitMySQL 应返回错误")
	}
}

func TestClosePreservesSingletonWhenCloseFails(t *testing.T) {
	// 连接池为空的 gorm 实例调用 DB() 会返回 ErrInvalidDB，closeMySQL 应保留单例供重试。
	// 注意：零值 &gorm.DB{} 会因 nil *Config 而直接 panic，这里显式提供空 Config.
	mysqlMu.Lock()
	prev := mysqlDB
	mysqlDB = &gorm.DB{Config: &gorm.Config{}}
	mysqlMu.Unlock()
	t.Cleanup(func() {
		mysqlMu.Lock()
		mysqlDB = prev
		mysqlMu.Unlock()
	})

	err := Close()
	if err == nil {
		t.Fatal("无效的 gorm 实例应导致 Close 返回错误")
	}
	if !strings.Contains(err.Error(), "get sql.DB") {
		t.Errorf("error = %v, 应包含获取 sql.DB 失败的原因", err)
	}
	if GetDB() == nil {
		t.Error("关闭失败时应保留单例，便于调用方重试")
	}
}

func TestCloseMySQLSuccessWithRealConnPool(t *testing.T) {
	// 构造带真实 *sql.DB 连接池的 gorm 实例（懒连接，不会真正建连），覆盖 closeMySQL 成功路径
	sqlDB, err := sql.Open("mysql", "root:pwd@tcp(127.0.0.1:1)/demo")
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}

	mysqlMu.Lock()
	prev := mysqlDB
	mysqlDB = &gorm.DB{Config: &gorm.Config{}, Statement: &gorm.Statement{ConnPool: sqlDB}}
	mysqlMu.Unlock()
	t.Cleanup(func() {
		mysqlMu.Lock()
		mysqlDB = prev
		mysqlMu.Unlock()
	})

	if err := closeMySQL(); err != nil {
		t.Fatalf("closeMySQL() error = %v", err)
	}
	if GetDB() != nil {
		t.Error("closeMySQL 成功后单例应被重置")
	}
	if GetRedis() != nil {
		t.Error("未初始化 Redis 时 GetRedis 应返回 nil")
	}
}
