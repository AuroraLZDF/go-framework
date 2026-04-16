// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/auroralzdf/biddingAI.

package util

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// ParamToInt64 从 URL 参数中获取 int64 类型的值.
func ParamToInt64(c *gin.Context, key string) (int64, error) {
	idStr := c.Request.FormValue(key)
	return strconv.ParseInt(idStr, 10, 64)
}

// AssignIfNotNil 检查指针是否非nil，非nil时赋值
func AssignIfNotNil[T any](dst *T, src *T) {
	if src != nil {
		*dst = *src
	}
}
