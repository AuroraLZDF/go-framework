// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/auroralzdf/biddingAI.

package db

import (
	"fmt"
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

// DSN 从 MySQLOptions 返回 DSN.
func (o *MySQLOptions) DSN() string {
	return fmt.Sprintf(`%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=%t&loc=%s`,
		o.Username,
		o.Password,
		o.Host,
		o.Database,
		true,
		"Local")
}

var (
	mysqlDB  *gorm.DB
	initOnce sync.Once
	initErr  error
)

// InitMySQL 使用给定的选项创建一个新的 gorm 数据库实例.
func InitMySQL(opts *MySQLOptions) (*gorm.DB, error) {
	initOnce.Do(func() {
		logLevel := logger.Silent
		if opts.LogLevel != 0 {
			logLevel = logger.LogLevel(opts.LogLevel)
		}

		db, err := gorm.Open(mysql.Open(opts.DSN()), &gorm.Config{
			Logger: logger.Default.LogMode(logLevel),
		})
		if err != nil {
			initErr = err
			return
		}

		sqlDB, err := db.DB()
		if err != nil {
			initErr = err
			return
		}

		sqlDB.SetMaxOpenConns(opts.MaxOpenConnections)
		sqlDB.SetMaxIdleConns(opts.MaxIdleConnections)
		sqlDB.SetConnMaxLifetime(opts.MaxConnectionLifeTime)

		mysqlDB = db
	})
	return mysqlDB, initErr
}

// GetDB 返回单例 *gorm.DB
func GetDB() *gorm.DB {
	return mysqlDB
}
