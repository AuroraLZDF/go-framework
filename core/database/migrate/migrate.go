// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

// Package migrate 提供基于 golang-migrate 的版本化数据库迁移能力：
//
//   - Create：在迁移目录中生成一对带时间戳版本的 up/down SQL 文件；
//   - Up / Down / Version：对 MySQL 执行迁移。
//
// 迁移文件位于项目的 migrations 目录，命名格式为：
//
//	<版本号>_<名称>.up.sql / <版本号>_<名称>.down.sql
//
// 版本号是创建时间戳（14 位数字），保证与 golang-migrate 的版本语义兼容。
//
// 使用示例：
//
//	if err := migrate.Up(migrate.Options{DSN: dsn, Dir: "migrations"}); err != nil {
//	    log.Fatalf("migrate up failed: %v", err)
//	}
//
// 注意：MySQL 驱动默认不允许多语句迁移文件，如需在单个 .sql 中写多条语句，
// DSN 必须包含 multiStatements=true 参数。
package migrate

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql" // 注册 mysql 驱动
	migrate "github.com/golang-migrate/migrate/v4"
	mysqlmig "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file" // 注册 file:// 源
)

// 默认迁移目录
const defaultDir = "migrations"

// migrationNamePattern 限制迁移名称：字母或数字开头，仅允许字母、数字、'-'、'_'.
var migrationNamePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

// Options 迁移执行选项.
type Options struct {
	// Dir 迁移 SQL 文件所在目录，默认 "migrations"
	Dir string
	// DSN go-sql-driver/mysql 格式的 DSN（必填），
	// 多语句迁移文件需要额外包含 multiStatements=true
	DSN string
}

// dir 返回规范化后的迁移目录
func (o *Options) dir() string {
	if o.Dir == "" {
		return defaultDir
	}

	return o.Dir
}

// validate 校验执行迁移前的选项与迁移目录.
func (o *Options) validate() error {
	if o == nil || o.DSN == "" {
		return errors.New("migrate: dsn is required")
	}

	return nil
}

// Create 在 dir 目录中生成一对迁移文件，返回 up 文件路径.
// 目录不存在时自动创建；名称校验失败或文件已存在时返回错误.
func Create(dir, name string) (string, error) {
	if !migrationNamePattern.MatchString(name) {
		return "", fmt.Errorf("migrate: invalid migration name %q: must match [a-zA-Z0-9][a-zA-Z0-9_-]*", name)
	}

	if dir == "" {
		dir = defaultDir
	}

	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", fmt.Errorf("migrate: create dir %s: %w", dir, err)
	}

	version := time.Now().Format("20060102150405")
	up := filepath.Join(dir, fmt.Sprintf("%s_%s.up.sql", version, name))
	down := filepath.Join(dir, fmt.Sprintf("%s_%s.down.sql", version, name))

	for _, p := range []string{up, down} {
		if _, err := os.Stat(p); err == nil {
			return "", fmt.Errorf("migrate: file %s already exists", p)
		}
	}

	if err := os.WriteFile(up, migrationTemplate(up), 0o600); err != nil {
		return "", fmt.Errorf("migrate: write %s: %w", up, err)
	}
	if err := os.WriteFile(down, migrationTemplate(down), 0o600); err != nil {
		return "", fmt.Errorf("migrate: write %s: %w", down, err)
	}

	return up, nil
}

// migrationTemplate 生成迁移文件的注释模板.
func migrationTemplate(path string) []byte {
	comment := fmt.Sprintf("-- Migration: %s\n-- Created: %s\n\n",
		filepath.Base(path), time.Now().Format(time.RFC3339))

	return []byte(comment)
}

// newMigrator 打开数据库连接并构建 golang-migrate 实例.
// 返回的 close 函数负责关闭底层连接池.
func newMigrator(opts *Options) (*migrate.Migrate, func(), error) {
	if err := opts.validate(); err != nil {
		return nil, nil, err
	}

	dir := opts.dir()
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return nil, nil, fmt.Errorf("migrate: migrations dir %s does not exist", dir)
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, nil, fmt.Errorf("migrate: resolve dir %s: %w", dir, err)
	}

	sqlDB, err := sql.Open("mysql", opts.DSN)
	if err != nil {
		return nil, nil, fmt.Errorf("migrate: open mysql: %w", err)
	}

	driver, err := mysqlmig.WithInstance(sqlDB, &mysqlmig.Config{})
	if err != nil {
		_ = sqlDB.Close()

		return nil, nil, fmt.Errorf("migrate: build mysql driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://"+absDir, "mysql", driver)
	if err != nil {
		_ = sqlDB.Close()

		return nil, nil, fmt.Errorf("migrate: init migrator: %w", err)
	}

	return m, func() { _ = sqlDB.Close() }, nil
}

// Up 将所有未应用的迁移按版本顺序执行到最新.
// 没有任何待执行的迁移时不视为错误.
func Up(opts *Options) error {
	m, closeFn, err := newMigrator(opts)
	if err != nil {
		return err
	}
	defer closeFn()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate: up: %w", err)
	}

	return nil
}

// Down 回滚迁移：steps > 0 时回滚最近 steps 个版本，否则回滚到最初状态.
// 回滚步数超过当前版本序号时，一次性回滚到最初状态.
func Down(opts *Options, steps int) error {
	m, closeFn, err := newMigrator(opts)
	if err != nil {
		return err
	}
	defer closeFn()

	if steps <= 0 {
		if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("migrate: down all: %w", err)
		}

		return nil
	}

	// Migrate(version uint) 按"目标版本号"迁移，回滚 N 步需要先定位目标版本
	current, _, err := m.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		return nil // 尚未执行过迁移，无需回滚
	}
	if err != nil {
		return fmt.Errorf("migrate: version: %w", err)
	}

	versions, err := listVersions(opts.dir())
	if err != nil {
		return err
	}

	target, ok := rollbackTarget(versions, current, steps)
	if !ok {
		// 回滚步数超过当前版本序号：直接回滚到最初状态
		if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("migrate: down all: %w", err)
		}

		return nil
	}

	if err := m.Migrate(target); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate: down %d: %w", steps, err)
	}

	return nil
}

// migrationFilePattern 匹配 up 迁移文件名，捕获 14 位时间戳版本号.
var migrationFilePattern = regexp.MustCompile(`^(\d{14})_[^/\\]+\.up\.sql$`)

// listVersions 读取迁移目录中的全部版本号，升序返回.
func listVersions(dir string) ([]uint, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("migrate: read dir %s: %w", dir, err)
	}

	seen := make(map[uint]struct{})
	versions := make([]uint, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		matched := migrationFilePattern.FindStringSubmatch(e.Name())
		if matched == nil {
			continue
		}

		version, err := strconv.ParseUint(matched[1], 10, 64)
		if err != nil {
			continue
		}

		if _, dup := seen[uint(version)]; dup {
			continue
		}
		seen[uint(version)] = struct{}{}
		versions = append(versions, uint(version))
	}

	sort.Slice(versions, func(i, j int) bool { return versions[i] < versions[j] })

	return versions, nil
}

// rollbackTarget 从升序版本列表中定位回滚 steps 步后的目标版本.
// 目标不存在时返回 ok=false.
func rollbackTarget(versions []uint, current uint, steps int) (uint, bool) {
	index := -1
	for i, v := range versions {
		if v == current {
			index = i
			break
		}
	}
	if index < 0 {
		// 当前版本不在迁移目录中（文件被删除）：按列表首端处理
		index = 0
	}

	targetIndex := index - steps
	if targetIndex < 0 || targetIndex >= len(versions) {
		return 0, false
	}

	return versions[targetIndex], true
}

// Version 返回当前数据库的迁移版本号与 dirty 状态.
// 尚未执行过任何迁移时版本号为 0.
func Version(opts *Options) (uint, bool, error) {
	m, closeFn, err := newMigrator(opts)
	if err != nil {
		return 0, false, err
	}
	defer closeFn()

	version, dirty, err := m.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("migrate: version: %w", err)
	}

	return version, dirty, nil
}
