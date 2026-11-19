// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package server

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	dbpkg "github.com/AuroraLZDF/go-framework/core/database/db"
)

// stubConnPool 是不会真正连接数据库的连接池桩，配合 DryRun 使用，
// 由于不是 *sql.DB，(*gorm.DB).DB() 会返回错误.
type stubConnPool struct{}

var errStubConnPool = errors.New("stub conn pool")

func (stubConnPool) PrepareContext(context.Context, string) (*sql.Stmt, error) {
	return nil, errStubConnPool
}

func (stubConnPool) ExecContext(context.Context, string, ...interface{}) (sql.Result, error) {
	return nil, errStubConnPool
}

func (stubConnPool) QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error) {
	return nil, errStubConnPool
}

func (stubConnPool) QueryRowContext(context.Context, string, ...interface{}) *sql.Row {
	return &sql.Row{}
}

// stubDriverConn 是最小化的 driver.Conn 实现，获取连接成功即可让 PingContext 返回 nil.
type stubDriverConn struct{}

func (stubDriverConn) Prepare(string) (driver.Stmt, error) { return nil, errStubConnPool }
func (stubDriverConn) Close() error                        { return nil }
func (stubDriverConn) Begin() (driver.Tx, error)           { return nil, errStubConnPool }

// healthyConnector 返回可用连接，用于模拟健康的 MySQL.
type healthyConnector struct{}

func (healthyConnector) Connect(context.Context) (driver.Conn, error) { return stubDriverConn{}, nil }
func (healthyConnector) Driver() driver.Driver                        { return stubDriver{} }

// failingConnector 永远连接失败，用于模拟 MySQL 连接不可达.
type failingConnector struct{}

func (failingConnector) Connect(context.Context) (driver.Conn, error) {
	return nil, errors.New("dial 127.0.0.1:1: connection refused")
}
func (failingConnector) Driver() driver.Driver { return stubDriver{} }

type stubDriver struct{}

func (stubDriver) Open(string) (driver.Conn, error) { return nil, errStubConnPool }

// newStubGormDB 用给定的 gorm.ConnPool 构造一个不做任何真实网络请求的 *gorm.DB.
func newStubGormDB(t *testing.T, connPool gorm.ConnPool) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      connPool,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{DisableAutomaticPing: true})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}

	return db
}

func newTestApp(t *testing.T) *Application {
	t.Helper()

	app, err := NewApplication(testConfig())
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}
	t.Cleanup(func() { _ = app.Close() })

	return app
}

func TestRegisterRoutesInvokesCallback(t *testing.T) {
	app := newTestApp(t)

	var gotEngine *gin.Engine
	var gotDB *gorm.DB
	called := false

	app.RegisterRoutes(func(engine *gin.Engine, db *gorm.DB) {
		called = true
		gotEngine = engine
		gotDB = db
	})

	if !called {
		t.Fatal("RegisterRoutes 应调用传入的路由注册函数")
	}
	if gotEngine != app.GetEngine() {
		t.Error("注册回调收到的 engine 与 GetEngine() 返回的不一致")
	}
	if gotDB != app.GetDB() {
		t.Error("注册回调收到的 db 与 GetDB() 返回的不一致")
	}

	// 传入 nil 不应 panic
	app.RegisterRoutes(nil)
}

func TestGetDBNilBeforeDatabaseInit(t *testing.T) {
	app := newTestApp(t)

	if app.GetDB() != nil {
		t.Error("未初始化数据库时 GetDB() 应返回 nil")
	}
}

func TestNewApplicationFailsOnUnreachableMySQL(t *testing.T) {
	cfg := testConfig()
	cfg.Database.MySQL.Host = "127.0.0.1:1"
	cfg.Database.MySQL.Username = "root"
	cfg.Database.MySQL.Database = "testdb"

	app, err := NewApplication(cfg)
	if err == nil {
		_ = app.Close()
		t.Fatal("MySQL 不可达时 NewApplication 应返回错误")
	}
	if !strings.Contains(err.Error(), "failed to init database") {
		t.Errorf("error = %v, 应包含 failed to init database", err)
	}
	if app != nil && app.GetDB() != nil {
		t.Error("初始化失败后不应保留数据库实例")
	}
}

func TestNewApplicationDegradesWhenRedisUnreachable(t *testing.T) {
	cfg := testConfig()
	cfg.Redis.Addr = "127.0.0.1:1"
	cfg.Redis.DialTimeout = 200 * time.Millisecond

	app, err := NewApplication(cfg)
	if err != nil {
		t.Fatalf("Redis 不可达时 NewApplication 应降级运行而不是失败, error = %v", err)
	}
	t.Cleanup(func() { _ = app.Close() })

	if dbpkg.GetRedis() != nil {
		t.Error("Redis 初始化失败后不应保留 Redis 客户端")
	}
}

func TestReadyzReturns503WhenDependenciesUnavailable(t *testing.T) {
	app := newTestApp(t)

	// 配置了依赖但实际未初始化成功：mysql/redis 都应报告 unavailable
	app.config.Database.MySQL.Host = "127.0.0.1:1"
	app.config.Redis.Addr = "127.0.0.1:1"

	w := httptest.NewRecorder()
	app.GetEngine().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/readyz", http.NoBody))

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusServiceUnavailable, w.Body.String())
	}
	body := w.Body.String()
	for _, want := range []string{`"mysql":"unavailable"`, `"redis":"unavailable"`} {
		if !strings.Contains(body, want) {
			t.Errorf("响应缺少 %s, body = %s", want, body)
		}
	}
}

func TestReadyzReturns503WhenGormDBInvalid(t *testing.T) {
	app := newTestApp(t)

	// *gorm.DB 底层连接池不是 *sql.DB，GetDBConn 失败应报告 unavailable
	app.db = newStubGormDB(t, stubConnPool{})
	app.config.Database.MySQL.Host = "127.0.0.1:1"

	w := httptest.NewRecorder()
	app.GetEngine().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/readyz", http.NoBody))

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusServiceUnavailable, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"mysql":"unavailable"`) {
		t.Errorf("响应应报告 mysql unavailable, body = %s", w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"redis":"skipped"`) {
		t.Errorf("未配置 Redis 时应报告 skipped, body = %s", w.Body.String())
	}
}

func TestReadyzReturns503WhenMySQLPingFails(t *testing.T) {
	app := newTestApp(t)

	app.db = newStubGormDB(t, sql.OpenDB(failingConnector{}))
	app.config.Database.MySQL.Host = "127.0.0.1:1"

	w := httptest.NewRecorder()
	app.GetEngine().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/readyz", http.NoBody))

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusServiceUnavailable, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"mysql":"unavailable"`) {
		t.Errorf("Ping 失败时响应应报告 mysql unavailable, body = %s", w.Body.String())
	}
}

func TestReadyzReturns200WhenMySQLHealthy(t *testing.T) {
	app := newTestApp(t)

	app.db = newStubGormDB(t, sql.OpenDB(healthyConnector{}))
	app.config.Database.MySQL.Host = "127.0.0.1:3306"

	if app.GetDB() == nil {
		t.Fatal("GetDB() 不应返回 nil")
	}

	w := httptest.NewRecorder()
	app.GetEngine().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/readyz", http.NoBody))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
	}
	body := w.Body.String()
	for _, want := range []string{`"mysql":"ok"`, `"redis":"skipped"`, `"status":"ok"`} {
		if !strings.Contains(body, want) {
			t.Errorf("响应缺少 %s, body = %s", want, body)
		}
	}
}
