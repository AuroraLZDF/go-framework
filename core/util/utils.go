// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package util

import (
	"crypto/rand"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// StructToMap converts a struct (or a non-nil pointer to struct) to a map.
// nil、非结构体返回 nil；未导出字段会被跳过，避免 reflect.Value.Interface panic.
func StructToMap(obj interface{}) map[string]string {
	if obj == nil {
		return nil
	}

	objValue := reflect.ValueOf(obj)

	// 支持传入 *struct，同时安全处理 nil 指针/nil interface
	for objValue.Kind() == reflect.Ptr || objValue.Kind() == reflect.Interface {
		if objValue.IsNil() {
			return nil
		}
		objValue = objValue.Elem()
	}

	if objValue.Kind() != reflect.Struct {
		return nil
	}

	m := make(map[string]string, objValue.NumField())
	objType := objValue.Type()

	for i := 0; i < objValue.NumField(); i++ {
		field := objType.Field(i)
		// 未导出字段的 Interface() 会 panic，直接跳过
		if !field.IsExported() {
			continue
		}
		m[toSnakeCase(field.Name)] = fmt.Sprintf("%v", objValue.Field(i).Interface())
	}

	return m
}

// toSnakeCase Converts a string to snake case.
// 处理连续大写/缩写：UserID -> user_id、HTTPServer -> http_server、ID -> id.
func toSnakeCase(name string) string {
	runes := []rune(name)
	var result []rune

	for i, r := range runes {
		if !unicode.IsUpper(r) {
			result = append(result, r)
			continue
		}

		if i > 0 {
			prev := runes[i-1]
			nextIsLower := i+1 < len(runes) && unicode.IsLower(runes[i+1])
			// 小写/数字之后的大写、或缩写末尾（前一字母大写且下一个字母小写）前插入下划线
			if unicode.IsLower(prev) || unicode.IsDigit(prev) || (unicode.IsUpper(prev) && nextIsLower) {
				result = append(result, '_')
			}
		}
		result = append(result, unicode.ToLower(r))
	}

	return string(result)
}

// IsWithinDuration 检查时间是否在指定时间段内
func IsWithinDuration(lastTime time.Time, duration time.Duration) bool {
	if lastTime.IsZero() {
		return false // 处理零值情况
	}

	return time.Since(lastTime) <= duration
}

// IsExpired 检查时间是否过期
func IsExpired(lastTime time.Time, duration time.Duration) bool {
	return !IsWithinDuration(lastTime, duration)
}

// JoinInt64s 连接int64数组
func JoinInt64s(nums []int64) string {
	strSlice := make([]string, len(nums))
	for i, num := range nums {
		strSlice[i] = strconv.FormatInt(num, 10)
	}
	return strings.Join(strSlice, ",")
}

// GenerateRandomString 生成指定长度的随机字符串
func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	// 丢弃会造成取模偏差的字节，保证字符分布均匀
	const maxByte = 256 - 256%len(charset)

	if length <= 0 {
		return ""
	}

	out := make([]byte, 0, length)
	buf := make([]byte, length)
	for len(out) < length {
		// crypto/rand.Read 在 Go 1.24+ 保证不会返回错误
		_, _ = rand.Read(buf)
		for _, v := range buf {
			if int(v) >= maxByte {
				continue
			}

			out = append(out, charset[int(v)%len(charset)])
			if len(out) == length {
				break
			}
		}
	}

	return string(out)
}
