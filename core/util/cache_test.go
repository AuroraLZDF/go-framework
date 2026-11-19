// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package util

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
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

func TestCacheExistsWithUnreachableRedis(t *testing.T) {
	// 直接注入指向不可达地址的客户端，覆盖 Exists 的错误分支
	client := redis.NewClient(&redis.Options{
		Addr:        "127.0.0.1:1",
		DialTimeout: 200 * time.Millisecond,
		ReadTimeout: 200 * time.Millisecond,
	})
	t.Cleanup(func() { _ = client.Close() })

	c := &cache{client: client}
	ok, err := c.Exists(context.Background(), "key")
	if ok {
		t.Error("Redis 不可达时 Exists 不应返回 true")
	}
	if err == nil {
		t.Error("Redis 不可达时 Exists 应返回错误")
	}

	// client 非 nil 时 Close 不应 panic
	c.Close()
}
