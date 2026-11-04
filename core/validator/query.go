// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/auroralzdf/biddingAI.

package validator

import (
	"fmt"
	"regexp"
	"strings"

	"gorm.io/gorm"
)

// fieldNamePattern 限制可用的字段名，避免拼接 SQL 时引入注入风险.
// 只允许简单标识符（字母/下划线开头，后接字母、数字、下划线），
// 不允许 "." 等限定名，避免越权访问其它表/列的表达式.
var fieldNamePattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

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
			return db.Where("title LIKE ?", "%"+escapeLikeValue(title)+"%")
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

// escapeLikeValue 转义 LIKE 模式中的通配符，避免用户输入被当作模式匹配.
// 反斜杠本身也要转义，否则用户可用 "\%" 绕过转义.
func escapeLikeValue(value string) string {
	return strings.NewReplacer(
		`\`, `\\`,
		`%`, `\%`,
		`_`, `\_`,
	).Replace(value)
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

// WithFilters 支持多种条件查询（包括LIKE）.
//
// 安全约束：
//   - Field 必须是简单标识符，否则 fail-closed（WHERE 1 = 0）；
//   - Value 始终参数化；LIKE 的 %/_ 会被转义；
//   - 未知 Operator 不再静默按 "=" 处理，而是 fail-closed 并通过 *gorm.DB 返回错误.
func WithFilters(filters ...Filter) QueryOption {
	return func(db *gorm.DB) *gorm.DB {
		for _, f := range filters {
			// 非法字段名直接返回空结果集（fail-closed），避免 SQL 注入
			if !fieldNamePattern.MatchString(f.Field) {
				db = db.Where("1 = 0")
				continue
			}

			switch f.Operator {
			case OpEqual:
				db = db.Where(fmt.Sprintf("%s = ?", f.Field), f.Value)
			case OpLike:
				pattern := "%" + escapeLikeValue(fmt.Sprint(f.Value)) + "%"
				db = db.Where(fmt.Sprintf("%s LIKE ?", f.Field), pattern)
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
				// 未知操作符：不再默认按等号处理，fail-closed 并返回错误
				db = db.Where("1 = 0")
				_ = db.AddError(fmt.Errorf("validator: unsupported filter operator %q on field %q", f.Operator, f.Field))
			}
		}
		return db
	}
}
