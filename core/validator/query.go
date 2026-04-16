// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/auroralzdf/biddingAI.

package tools

import (
	"fmt"

	"gorm.io/gorm"
)

// QueryOption 定义查询选项函数类型
type QueryOption func(*gorm.DB) *gorm.DB

// WithId 查询选项函数
func WithId(id int64) QueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", id)
	}
}

func WithUserId(uid int64) QueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uid = ?", uid)
	}
}

// WithPhone 查询选项函数
func WithPhone(phone string) QueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("phone = ?", phone)
	}
}

func WithCode(code string) QueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ?", code)
	}
}

func WithStatus(status int64) QueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ?", status)
	}
}

func WithTitle(title string) QueryOption {
	return func(db *gorm.DB) *gorm.DB {
		if title != "" {
			return db.Where("title LIKE ?", "%"+title+"%")
		}
		return db // 返回未修改的db
	}
}

func WithArchived(archived int64) QueryOption {
	return func(db *gorm.DB) *gorm.DB {
		if archived != 0 {
			return db.Where("archived = ?", archived)
		}
		return db // 返回未修改的db
	}
}

func WithType(t int64) QueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("type = ?", t)
	}
}

func WithMcId(mcId int64) QueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("mc_id = ?", mcId)
	}
}

func WithName(name string) QueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("name = ?", name)
	}
}

// FilterOperator 定义过滤器操作符类型
type FilterOperator string

const (
	OpEqual FilterOperator = "="
	OpLike  FilterOperator = "LIKE"
	OpGt    FilterOperator = ">"
	OpLt    FilterOperator = "<"
	OpGte   FilterOperator = ">="
	OpLte   FilterOperator = "<="
	OpIn    FilterOperator = "IN"
	OpNotIn FilterOperator = "NOT IN"
	// 可扩展其他操作符...
)

// Filter 定义过滤器结构体
type Filter struct {
	Field    string
	Value    interface{}
	Operator FilterOperator
}

// WithFilters 支持多种条件查询（包括LIKE）
func WithFilters(filters ...Filter) QueryOption {
	return func(db *gorm.DB) *gorm.DB {
		for _, f := range filters {
			switch f.Operator {
			case OpEqual:
				db = db.Where(fmt.Sprintf("%s = ?", f.Field), f.Value)
			case OpLike:
				db = db.Where(fmt.Sprintf("%s LIKE ?", f.Field), "%"+fmt.Sprint(f.Value)+"%")
			case OpGt:
				db = db.Where(fmt.Sprintf("%s > ?", f.Field), f.Value)
			case OpLt:
				db = db.Where(fmt.Sprintf("%s < ?", f.Field), f.Value)
			case OpGte:
				db = db.Where(fmt.Sprintf("%s >= ?", f.Field), f.Value)
			case OpLte:
				db = db.Where(fmt.Sprintf("%s <= ?", f.Field), f.Value)
			case OpIn:
				db = db.Where(fmt.Sprintf("%s IN (?)", f.Field), f.Value)
			case OpNotIn:
				db = db.Where(fmt.Sprintf("%s NOT IN (?)", f.Field), f.Value)
			// 添加其他操作符...
			default:
				db = db.Where(fmt.Sprintf("%s = ?", f.Field), f.Value) // 默认等于
			}
		}
		return db
	}
}
