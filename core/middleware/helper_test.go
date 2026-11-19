// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package middleware

import (
	"net/http"
	"testing"
	"time"
)

func TestNoCacheSetsHeaders(t *testing.T) {
	w := serve(t, NoCache, http.MethodGet, "")

	if got := w.Header().Get("Cache-Control"); got != "no-store, no-cache, must-revalidate, max-age=0" {
		t.Errorf("Cache-Control = %q, want 禁用缓存的组合头", got)
	}
	if got := w.Header().Get("Pragma"); got != "no-cache" {
		t.Errorf("Pragma = %q, want no-cache", got)
	}
	if got := w.Header().Get("Expires"); got != "Thu, 01 Jan 1970 00:00:00 GMT" {
		t.Errorf("Expires = %q, want 过去的固定时间", got)
	}

	lastModified := w.Header().Get("Last-Modified")
	if lastModified == "" {
		t.Error("Last-Modified 头不应为空")
		return
	}
	if _, err := http.ParseTime(lastModified); err != nil {
		t.Errorf("Last-Modified = %q 不是合法的 HTTP 时间格式: %v", lastModified, err)
	}
	if lm, _ := http.ParseTime(lastModified); time.Since(lm) < 0 {
		t.Errorf("Last-Modified = %q 不应在未来", lastModified)
	}
}
