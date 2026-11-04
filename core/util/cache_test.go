// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package util

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCacheUnavailableWithoutRedis(t *testing.T) {
	c := NewCache()
	ctx := context.Background()

	if _, err := c.Get(ctx, "key"); !errors.Is(err, ErrCacheUnavailable) {
		t.Errorf("Get() error = %v, want ErrCacheUnavailable", err)
	}
	if err := c.Set(ctx, "key", "value", time.Minute); !errors.Is(err, ErrCacheUnavailable) {
		t.Errorf("Set() error = %v, want ErrCacheUnavailable", err)
	}
	if err := c.Delete(ctx, "key"); !errors.Is(err, ErrCacheUnavailable) {
		t.Errorf("Delete() error = %v, want ErrCacheUnavailable", err)
	}
	if _, err := c.Exists(ctx, "key"); !errors.Is(err, ErrCacheUnavailable) {
		t.Errorf("Exists() error = %v, want ErrCacheUnavailable", err)
	}

	// 客户端为空时 Close 不应 panic
	c.Close()
}
