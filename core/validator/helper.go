// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/auroralzdf/biddingAI.

package tools

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/go-playground/validator/v10"
)

const defaultLimitValue = 20

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
	regexPattern := fl.Param()

	// 编译正则并匹配
	matched, err := regexp.MatchString(regexPattern, value)
	if err != nil {
		// 正则表达式无效时返回 false
		return false
	}
	return matched
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
