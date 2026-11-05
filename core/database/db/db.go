// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package db

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// MySQLOptions 定义 MySQL 数据库的选项.
type MySQLOptions struct {
	Host                  string
	Username              string
	Password              string
	Database              string
	MaxIdleConnections    int
	MaxOpenConnections    int
	MaxConnectionLifeTime time.Duration
	LogLevel              int
}

const (
	// defaultDialTimeout 建立连接的超时时间
	defaultDialTimeout = 5 * time.Second
	// defaultReadTimeout 读取数据的超时时间
	defaultReadTimeout = 30 * time.Second
	// defaultWriteTimeout 写入数据的超时时间
	defaultWriteTimeout = 30 * time.Second
)

// ErrDBAlreadyInitialized 表示 MySQL 单例已经初始化，调用方需要先 Close() 再重新初始化.
var ErrDBAlreadyInitialized = errors.New("db: mysql has already been initialized, call Close() first")

// DSN 从 MySQLOptions 返回 DSN.
// 显式设置读写/连接超时，避免下游异常时 goroutine 长时间阻塞；
// loc 固定为 UTC，时间统一按 UTC 存储，展示时再转换为本地时区.
//
// 转义/校验说明（go-sql-driver/mysql 的 DSN 不支持对用户名、密码转义）：
//   - database 通过 url.PathEscape 转义，因此可以包含 '/'、'?'、'&'、'%' 等字符；
//   - username 拒绝 ':'（第一处 ':' 会被驱动解析为密码分隔符）和控制字符；
//   - password 保留原样，驱动从 DSN 末尾向前查找最后一个 '@' 和 '/'，因此密码中的
//     ':'、'@'、'/'、'?' 不会破坏解析；但仍拒绝换行等控制字符；
//   - host 必须非空且不能包含控制字符或空白（net 地址位于 tcp(...) 中）.
func (o *MySQLOptions) DSN() (string, error) {
	if o == nil {
		return "", errors.New("db: mysql options is nil")
	}
	if err := validateMySQLOptions(o); err != nil {
		return "", err
	}

	return fmt.Sprintf(`%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=%t&loc=%s&timeout=%s&readTimeout=%s&writeTimeout=%s`,
		o.Username,
		o.Password,
		o.Host,
		url.PathEscape(o.Database),
		true,
		"UTC",
		defaultDialTimeout,
		defaultReadTimeout,
		defaultWriteTimeout), nil
}

// validateMySQLOptions 校验会破坏 DSN 结构的字符，并给出明确的错误.
func validateMySQLOptions(o *MySQLOptions) error {
	if o.Host == "" {
		return errors.New("db: mysql host is required")
	}
	if strings.ContainsAny(o.Host, " \t\r\n") || hasControlChars(o.Host) {
		return fmt.Errorf("db: mysql host %q contains whitespace or control characters", o.Host)
	}
	if strings.Contains(o.Username, ":") {
		return fmt.Errorf("db: mysql username %q must not contain ':' (the driver treats the first ':' as the password separator)", o.Username)
	}
	if hasControlChars(o.Username) {
		return errors.New("db: mysql username contains control characters")
	}
	if hasControlChars(o.Password) {
		return errors.New("db: mysql password contains control characters")
	}
	if hasControlChars(o.Database) {
		return errors.New("db: mysql database name contains control characters")
	}

	return nil
}

// hasControlChars 判断字符串中是否包含 ASCII 控制字符.
func hasControlChars(s string) bool {
	for _, r := range s {
		if r < 0x20 || r == 0x7f {
			return true
		}
	}

	return false
}

var (
	mysqlDB *gorm.DB
	mysqlMu sync.Mutex
)

// InitMySQL 使用给定的选项创建一个新的 gorm 数据库实例.
// 初始化成功后复用该实例；失败时不会缓存错误，允许调用方重试.
// 若单例已存在则返回 ErrDBAlreadyInitialized，需要通过 Close() 关闭后再重新初始化.
func InitMySQL(opts *MySQLOptions) (*gorm.DB, error) {
	if opts == nil {
		return nil, errors.New("db: mysql options is nil")
	}

	mysqlMu.Lock()
	defer mysqlMu.Unlock()

	if mysqlDB != nil {
		return nil, ErrDBAlreadyInitialized
	}

	// 填充连接池默认值，避免使用 database/sql 的零值导致连接不可复用
	o := *opts
	if o.MaxOpenConnections <= 0 {
		o.MaxOpenConnections = 100
	}
	if o.MaxIdleConnections <= 0 {
		o.MaxIdleConnections = 10
	}
	if o.MaxIdleConnections > o.MaxOpenConnections {
		o.MaxIdleConnections = o.MaxOpenConnections
	}
	if o.MaxConnectionLifeTime <= 0 {
		o.MaxConnectionLifeTime = time.Hour
	}

	dsn, err := o.DSN()
	if err != nil {
		return nil, fmt.Errorf("invalid mysql options: %w", err)
	}

	logLevel := logger.Silent
	if o.LogLevel != 0 {
		logLevel = logger.LogLevel(o.LogLevel)
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(o.MaxOpenConnections)
	sqlDB.SetMaxIdleConns(o.MaxIdleConnections)
	sqlDB.SetConnMaxLifetime(o.MaxConnectionLifeTime)

	mysqlDB = db

	return mysqlDB, nil
}

// GetDB 返回单例 *gorm.DB
func GetDB() *gorm.DB {
	mysqlMu.Lock()
	defer mysqlMu.Unlock()

	return mysqlDB
}

// Close 关闭 MySQL 与 Redis 连接，返回聚合后的错误
func Close() error {
	return errors.Join(closeMySQL(), closeRedis())
}

// closeMySQL 先关闭连接池，成功后再重置单例，便于重新初始化.
// 关闭失败时保留单例，调用方可以重试 Close().
func closeMySQL() error {
	mysqlMu.Lock()
	instance := mysqlDB
	mysqlMu.Unlock()

	if instance == nil {
		return nil
	}

	sqlDB, err := instance.DB()
	if err != nil {
		return fmt.Errorf("get sql.DB: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("close mysql: %w", err)
	}

	// 仅当期间没有被其他调用替换时才清空，避免误清新实例
	mysqlMu.Lock()
	if mysqlDB == instance {
		mysqlDB = nil
	}
	mysqlMu.Unlock()

	return nil
}
