// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package validator

import (
	"fmt"
	"strings"
	"testing"

	govalidator "github.com/go-playground/validator/v10"
)

type phoneForm struct {
	Phone string `validate:"required,customRegex=^1[3-9]\\d{9}$"`
}

type badRegexForm struct {
	Value string `validate:"customRegex=("`
}

func TestRegisterCustomValidations(t *testing.T) {
	v := govalidator.New()
	if err := RegisterCustomValidations(v); err != nil {
		t.Fatalf("RegisterCustomValidations() error = %v", err)
	}

	if err := v.Struct(phoneForm{Phone: "13800138000"}); err != nil {
		t.Errorf("合法手机号校验失败: %v", err)
	}
	if err := v.Struct(phoneForm{Phone: "12345"}); err == nil {
		t.Error("非法手机号应校验失败")
	}
}

func TestRegisterCustomValidationsNil(t *testing.T) {
	if err := RegisterCustomValidations(nil); err == nil {
		t.Error("传入 nil validator 时应返回错误")
	}
}

func TestValidateRegexInvalidPattern(t *testing.T) {
	v := govalidator.New()
	if err := RegisterCustomValidations(v); err != nil {
		t.Fatalf("RegisterCustomValidations() error = %v", err)
	}

	if err := v.Struct(badRegexForm{Value: "anything"}); err == nil {
		t.Error("非法正则表达式应校验失败")
	}
}

func resetRegexCache() {
	regexCacheMu.Lock()
	regexCache = make(map[string]regexCacheEntry, maxRegexCacheEntries)
	regexCacheMu.Unlock()
}

func regexCacheSize() int {
	regexCacheMu.RLock()
	defer regexCacheMu.RUnlock()

	return len(regexCache)
}

func TestRegexCacheCachesInvalidPatterns(t *testing.T) {
	resetRegexCache()

	if entry := compiledRegex("("); entry.valid {
		t.Fatal("非法正则不应编译成功")
	}
	if entry := compiledRegex("("); entry.valid {
		t.Fatal("非法正则不应编译成功")
	}
	if got := regexCacheSize(); got != 1 {
		t.Errorf("非法正则应被 negative cache 缓存, cache size = %d, want 1", got)
	}

	if entry := compiledRegex("^a+$"); !entry.valid {
		t.Fatal("合法正则应编译成功")
	}
	if got := regexCacheSize(); got != 2 {
		t.Errorf("cache size = %d, want 2", got)
	}
}

func TestRegexCacheIsBounded(t *testing.T) {
	resetRegexCache()

	for i := 0; i < maxRegexCacheEntries*2; i++ {
		compiledRegex(fmt.Sprintf("^value%d$", i))
	}

	if got := regexCacheSize(); got > maxRegexCacheEntries {
		t.Errorf("缓存条目数 = %d, 超过上限 %d", got, maxRegexCacheEntries)
	}
}

func TestDefaultLimit(t *testing.T) {
	if got := DefaultLimit(0); got != defaultLimitValue {
		t.Errorf("DefaultLimit(0) = %d, want %d", got, defaultLimitValue)
	}
	if got := DefaultLimit(5); got != 5 {
		t.Errorf("DefaultLimit(5) = %d, want 5", got)
	}
}

func TestJSONToMarkdown(t *testing.T) {
	got, err := JSONToMarkdown(map[string]string{"name": "test"})
	if err != nil {
		t.Fatalf("JSONToMarkdown() error = %v", err)
	}
	if !strings.HasPrefix(got, "```json\n") || !strings.HasSuffix(got, "\n```") {
		t.Errorf("返回内容缺少 Markdown 代码块标记: %q", got)
	}
	if !strings.Contains(got, `"name": "test"`) {
		t.Errorf("返回内容缺少序列化结果: %q", got)
	}
}
