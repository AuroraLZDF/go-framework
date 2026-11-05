// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package validator

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sync"

	"github.com/go-playground/validator/v10"
)

const (
	defaultLimitValue = 20

	// maxRegexCacheEntries 限制正则缓存条目数，避免不可信模式导致内存无限增长.
	maxRegexCacheEntries = 256
)

// regexCacheEntry 缓存正则编译结果：valid=false 时表示该模式编译失败（negative cache）.
type regexCacheEntry struct {
	re    *regexp.Regexp
	valid bool
}

var (
	// regexCache 缓存已编译的正则表达式（含非法模式），避免重复编译.
	regexCacheMu sync.RWMutex
	regexCache   = make(map[string]regexCacheEntry, maxRegexCacheEntries)
)

// compiledRegex 返回模式对应的缓存项，未命中时编译并写入缓存（含非法模式）.
func compiledRegex(pattern string) regexCacheEntry {
	regexCacheMu.RLock()
	entry, ok := regexCache[pattern]
	regexCacheMu.RUnlock()
	if ok {
		return entry
	}

	compiled, err := regexp.Compile(pattern)
	entry = regexCacheEntry{re: compiled, valid: err == nil}

	regexCacheMu.Lock()
	if len(regexCache) >= maxRegexCacheEntries {
		// 简单淘汰：缓存满时删除任意一个条目，保证容量有界
		for key := range regexCache {
			delete(regexCache, key)
			break
		}
	}
	regexCache[pattern] = entry
	regexCacheMu.Unlock()

	return entry
}

// DefaultLimit 设置默认查询记录数.
func DefaultLimit(limit int) int {
	if limit == 0 {
		limit = defaultLimitValue
	}

	return limit
}

// ValidateRegex 自定义正则验证函数（标签为"customRegex"）
func ValidateRegex(fl validator.FieldLevel) bool {
	// 获取字段值（字符串类型）
	value := fl.Field().String()
	// 获取标签中定义的正则表达式（如 binding:"customRegex=^1[3-9]\\d{9}$"）
	entry := compiledRegex(fl.Param())
	if !entry.valid {
		// 非法正则表达式一律校验失败（negative cache 避免重复编译）
		return false
	}

	return entry.re.MatchString(value)
}

// RegisterCustomValidations 注册自定义校验规则（customRegex）
func RegisterCustomValidations(v *validator.Validate) error {
	if v == nil {
		return fmt.Errorf("validator is nil")
	}

	return v.RegisterValidation("customRegex", ValidateRegex)
}

// JSONToMarkdown 把任意对象序列化为 JSON 字符串，并包裹 ```json 标记
func JSONToMarkdown(data interface{}) (string, error) {
	// 1. 序列化 JSON（Indent 可选，格式化输出更易读；不想要格式化用 json.Marshal）
	jsonBytes, err := json.MarshalIndent(data, "\n", "  ")
	if err != nil {
		return "", fmt.Errorf("JSON序列化失败: %w", err)
	}

	// 2. 拼接 Markdown 标记（首尾分别加 ```json 和 ```）
	// 注意：jsonBytes 转 string 已自动处理转义（如双引号、换行），无需额外处理
	markdownStr := fmt.Sprintf("```json\n%s\n```", string(jsonBytes))
	return markdownStr, nil
}
