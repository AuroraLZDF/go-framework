// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/auroralzdf/biddingAI.

/**
 * 问题：
 * 使用 json 格式化 struct 时，time.Time 被格式化成
 * "2006-01-02T15:04:05.999999999Z07:00" 格式
 *
 * golang 的 time.Time 的默认 json 格式化格式叫做 RFC3339。
 * 好像是一种国际标准，被推荐用作 json 时间的标准格式。
 * 但是使用中不需要这种，而且不容易解析。
 * 示例：
	type Model struct {
		ID        int64      `gorm:"column:id;primary_key" json:"id"`    //
		CreatedAt util.Time `gorm:"column:createdAt" json:"created_at"` //
		UpdatedAt util.Time `gorm:"column:updatedAt" json:"updated_at"` //
	}
*/

package util

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"log"
	"runtime/debug"
	"time"
)

const (
	DefaultTimeFormat = time.DateTime
)

type Time time.Time

// UnmarshalJSON 格式化自定义时间转换为 time.Time
func (t *Time) UnmarshalJSON(data []byte) (err error) {
	defer func() {
		if e := recover(); e != nil {
			log.Printf("Time unmarshal panic: %v\n%s", e, debug.Stack())
		}
	}()

	// 检查 JSON 字符串是否以引号开头和结尾
	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return errors.New("time: invalid json format")
	}

	// 去掉引号
	str := string(data[1 : len(data)-1])

	// 解析时间字符串
	parsed, err := time.ParseInLocation(DefaultTimeFormat, str, time.Local)
	if err != nil {
		return fmt.Errorf("time: parse error: %w", err)
	}

	*t = Time(parsed)
	return nil
}

// MarshalJSON 自定义时间格式化输出
func (t Time) MarshalJSON() ([]byte, error) {
	// 分配一个字节切片，长度为时间字符串长度 + 2（引号）
	buf := make([]byte, 0, len(DefaultTimeFormat)+2)
	buf = append(buf, '"')
	buf = time.Time(t).AppendFormat(buf, DefaultTimeFormat)
	buf = append(buf, '"')
	return buf, nil
}

// String 格式化输出
func (t Time) String() string {
	return time.Time(t).Format(DefaultTimeFormat)
}

// Scan 实现 GORM 的 Scanner 接口
// 将数据库返回的值转换为 util.Time 类型
func (t *Time) Scan(value interface{}) error {
	switch v := value.(type) {
	case nil:
		*t = Time(time.Time{})
	case time.Time:
		*t = Time(v)
	case []byte:
		parsed, err := time.ParseInLocation(DefaultTimeFormat, string(v), time.Local)
		if err != nil {
			return fmt.Errorf("time: db scan parse error: %w", err)
		}
		*t = Time(parsed)
	default:
		return fmt.Errorf("time: unsupported scan type: %T", value)
	}
	return nil
}

// Value 实现 GORM 的 Valuer 接口
// 将 util.Time 类型转换为数据库可接受的值
func (t Time) Value() (driver.Value, error) {
	return time.Time(t).UTC(), nil // 存储为UTC时间
}

// Now 获取当前时间 (实用方法)
func Now() Time {
	return Time(time.Now())
}

// FromTime 标准时间转换 (辅助方法)
func FromTime(t time.Time) Time {
	return Time(t)
}

// ToTime 转换为标准时间 (辅助方法)
func (t Time) ToTime() time.Time {
	return time.Time(t)
}
