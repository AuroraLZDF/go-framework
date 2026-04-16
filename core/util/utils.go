// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/auroralzdf/biddingAI.

package util

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// StructToMap converts a struct to a map
func StructToMap(obj interface{}) map[string]string {
	objValue := reflect.ValueOf(obj)
	if objValue.Kind() != reflect.Struct {
		return nil
	}

	m := make(map[string]string)
	objType := objValue.Type()

	for i := 0; i < objValue.NumField(); i++ {
		fieldName := objType.Field(i).Name
		// 转换为下划线形式的字段名称
		snakeFieldName := toSnakeCase(fieldName)
		fieldValue := fmt.Sprintf("%v", objValue.Field(i).Interface())
		m[snakeFieldName] = fieldValue
	}

	return m
}

// toSnakeCase Converts a string to snake case.
func toSnakeCase(name string) string {
	var result []rune

	for i, r := range name {
		if unicode.IsUpper(r) {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
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

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
		time.Sleep(time.Nanosecond)
	}
	return string(b)
}
