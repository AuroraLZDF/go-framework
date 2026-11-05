// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package response

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/AuroraLZDF/go-framework/core/errno"
)

func newTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", http.NoBody)

	return c, w
}

func TestErrorUsesErrnoHTTPStatusAndOmitsContent(t *testing.T) {
	c, w := newTestContext()

	Error(c, errno.ErrPageNotFound)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
	body := w.Body.String()
	if !strings.Contains(body, `"code":404`) {
		t.Errorf("响应缺少业务错误码, body = %s", body)
	}
	if strings.Contains(body, "content") {
		t.Errorf("错误响应不应包含 content 字段, body = %s", body)
	}
}

func TestErrorDoesNotLeakUnknownErrorDetail(t *testing.T) {
	c, w := newTestContext()

	Error(c, errors.New("connect mysql failed: password=secret"))

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
	body := w.Body.String()
	if strings.Contains(body, "password=secret") {
		t.Errorf("响应不应泄露内部错误信息, body = %s", body)
	}
	if !strings.Contains(body, errno.InternalServerError.Message) {
		t.Errorf("响应应返回统一的内部错误信息, body = %s", body)
	}
}
