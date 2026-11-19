// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

// Package store 提供基于 GORM 的通用 CRUD 仓储基类：
// 开箱即用的分页、字段过滤、排序校验与软删除约定。
//
// 使用示例：
//
//	type User struct {
//	    gorm.Model
//	    Name string
//	}
//
//	repo := store.NewRepository[model.User](db)
//	repo.Create(ctx, &user)
//	page, err := repo.List(ctx, store.Query{Page: 1, PageSize: 20, Filters: map[string]any{"name": "alice"}})
//
// 过滤条件为等值匹配，字段名必须满足列名规范，否则返回错误（fail-closed），
// 避免 SQL 注入。排序字段同样受白名单校验。
package store

import (
	"context"
	"errors"
	"reflect"
	"regexp"
	"strings"

	"gorm.io/gorm"
)

// 默认分页参数.
const (
	// DefaultPageSize 未指定分页大小时每页返回的记录数
	DefaultPageSize = 20
	// MaxPageSize 单页允许的最大记录数，超出会被截断
	MaxPageSize = 100
)

// ErrNotFound 表示按条件未查询到记录.
var ErrNotFound = gorm.ErrRecordNotFound

// ErrInvalidField 表示过滤/排序字段名不合法（fail-closed，拒绝执行）.
var ErrInvalidField = errors.New("store: invalid field name")

// ErrInvalidSort 表示排序方向不合法.
var ErrInvalidSort = errors.New("store: invalid sort direction")

// primaryKey 通过反射读取实体的 ID 字段（gorm.Model 或自定义主键名为 ID 的模型）.
// ID 字段不存在或为零值时返回 ok=false.
func primaryKey[T any](entity *T) (uint, bool) {
	v := reflect.ValueOf(entity)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return 0, false
	}

	field := v.Elem().FieldByName("ID")
	if !field.IsValid() || field.Kind() != reflect.Uint {
		return 0, false
	}

	id := field.Uint()

	return uint(id), id != 0
}

// columnNamePattern 限定数据库列名：字母或下划线开头，仅允许字母、数字、下划线与点（表.列）.
var columnNamePattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_.]*$`)

// Query 列表查询的分页与过滤参数.
type Query struct {
	// Page 页码，从 1 开始，默认 1
	Page int
	// PageSize 每页记录数，默认 DefaultPageSize，最大 MaxPageSize
	PageSize int
	// Sort 排序字段，可带 " asc"/" desc" 方向后缀，如 "created_at desc"；默认按 id 倒序
	Sort string
	// Filters 等值过滤条件，键为数据库列名（如 "name" 或 "users.name"）
	Filters map[string]any
}

// normalize 归一化分页参数，返回 page/pageSize.
func (q *Query) normalize() (int, int) {
	page := q.Page
	if page <= 0 {
		page = 1
	}

	pageSize := q.PageSize
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	return page, pageSize
}

// PageResult 分页查询结果.
type PageResult[T any] struct {
	// Items 当前页记录
	Items []T
	// Total 过滤条件下的记录总数
	Total int64
	// Page 当前页码（归一化后）
	Page int
	// PageSize 每页记录数（归一化后）
	PageSize int
}

// Repository 基于 GORM 的泛型 CRUD 仓储基类.
type Repository[T any] struct {
	db *gorm.DB
}

// NewRepository 创建仓储实例，db 不能为 nil.
func NewRepository[T any](db *gorm.DB) (*Repository[T], error) {
	if db == nil {
		return nil, errors.New("store: db is required")
	}

	return &Repository[T]{db: db}, nil
}

// Create 插入一条记录.
func (r *Repository[T]) Create(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

// GetByID 按主键查询，未找到时返回 ErrNotFound.
func (r *Repository[T]) GetByID(ctx context.Context, id uint) (*T, error) {
	var entity T
	if err := r.db.WithContext(ctx).First(&entity, id).Error; err != nil {
		return nil, err
	}

	return &entity, nil
}

// Update 按主键更新全部字段（含零值），记录不存在时返回 ErrNotFound.
func (r *Repository[T]) Update(ctx context.Context, entity *T) error {
	id, ok := primaryKey(entity)
	if !ok {
		return ErrNotFound
	}

	// 不使用 Save：它对不存在的记录会隐式插入（upsert），不符合仓储语义；
	// Select("*") 保证零值字段也会被更新
	result := r.db.WithContext(ctx).Model(new(T)).Where("id = ?", id).Select("*").Updates(entity)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// Delete 按主键删除。模型包含 gorm.DeletedAt 字段时执行软删除，否则物理删除.
func (r *Repository[T]) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(new(T), id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// Count 返回过滤条件下的记录总数.
func (r *Repository[T]) Count(ctx context.Context, q Query) (int64, error) {
	tx, err := applyFilters(r.db.WithContext(ctx), q)
	if err != nil {
		return 0, err
	}

	var total int64
	if err := tx.Model(new(T)).Count(&total).Error; err != nil {
		return 0, err
	}

	return total, nil
}

// List 分页查询。过滤、排序、分页一并应用；items 为空时返回空切片而非 nil.
func (r *Repository[T]) List(ctx context.Context, q Query) (*PageResult[T], error) {
	tx, err := applyFilters(r.db.WithContext(ctx), q)
	if err != nil {
		return nil, err
	}

	total, err := count[T](tx)
	if err != nil {
		return nil, err
	}

	page, pageSize := q.normalize()
	tx, err = applySort(tx, q)
	if err != nil {
		return nil, err
	}

	items := make([]T, 0, pageSize)
	if err := tx.Model(new(T)).Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, err
	}

	return &PageResult[T]{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// count 基于过滤后的查询统计总数（独立会话，避免分页子句影响）.
func count[T any](tx *gorm.DB) (int64, error) {
	var total int64
	if err := tx.Model(new(T)).Count(&total).Error; err != nil {
		return 0, err
	}

	return total, nil
}

// applyFilters 将等值过滤条件追加到查询，字段名必须通过列名校验.
func applyFilters(tx *gorm.DB, q Query) (*gorm.DB, error) {
	for field, value := range q.Filters {
		if !columnNamePattern.MatchString(field) {
			return nil, ErrInvalidField
		}

		tx = tx.Where(field+" = ?", value)
	}

	return tx, nil
}

// applySort 应用排序，字段名与方向均需通过校验；未指定时默认 id 倒序.
func applySort(tx *gorm.DB, q Query) (*gorm.DB, error) {
	sort := strings.TrimSpace(q.Sort)
	if sort == "" {
		return tx.Order("id DESC"), nil
	}

	field, direction, hasDirection := strings.Cut(sort, " ")
	if !columnNamePattern.MatchString(field) {
		return nil, ErrInvalidField
	}

	dir := "ASC"
	if hasDirection {
		switch strings.ToLower(strings.TrimSpace(direction)) {
		case "asc":
		case "desc":
			dir = "DESC"
		default:
			return nil, ErrInvalidSort
		}
	}

	return tx.Order(field + " " + dir), nil
}
