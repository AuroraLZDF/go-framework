// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package cache

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

type cachedUser struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func newTestCache(t *testing.T) (*Cache, *miniredis.Miniredis) {
	t.Helper()

	srv := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: srv.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	c, err := New(client)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	return c, srv
}

func TestNewNilClient(t *testing.T) {
	if _, err := New(nil); !errors.Is(err, ErrClientRequired) {
		t.Errorf("New(nil) = %v, want ErrClientRequired", err)
	}
}

func TestSetGetDelete(t *testing.T) {
	c, _ := newTestCache(t)
	ctx := context.Background()

	user := cachedUser{ID: 42, Name: "alice"}
	if err := c.Set(ctx, "user:42", user, time.Minute); err != nil {
		t.Fatalf("Set: %v", err)
	}

	var got cachedUser
	if err := c.Get(ctx, "user:42", &got); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ID != 42 || got.Name != "alice" {
		t.Errorf("Get = %+v, want {42 alice}", got)
	}

	if err := c.Delete(ctx, "user:42"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if err := c.Get(ctx, "user:42", &got); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get after Delete = %v, want ErrNotFound", err)
	}
}

func TestGetNotFound(t *testing.T) {
	c, _ := newTestCache(t)

	var got cachedUser
	if err := c.Get(context.Background(), "missing", &got); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get missing key = %v, want ErrNotFound", err)
	}
}

func TestGetCorruptedValue(t *testing.T) {
	c, srv := newTestCache(t)

	if err := srv.Set("bad", "not-json"); err != nil {
		t.Fatalf("miniredis set: %v", err)
	}

	var got cachedUser
	err := c.Get(context.Background(), "bad", &got)
	if err == nil || errors.Is(err, ErrNotFound) {
		t.Errorf("Get corrupted value should fail with unmarshal error, got %v", err)
	}
}

func TestGetOrSet(t *testing.T) {
	c, _ := newTestCache(t)
	ctx := context.Background()
	loader := func(context.Context) (cachedUser, error) {
		return cachedUser{ID: 1, Name: "bob"}, nil
	}

	// 未命中：回源并写入缓存
	var got cachedUser
	if err := GetOrSet(ctx, c, "user:1", time.Minute, loader, &got); err != nil {
		t.Fatalf("GetOrSet: %v", err)
	}
	if got.Name != "bob" {
		t.Errorf("GetOrSet = %+v", got)
	}

	// 命中：不再调用 loader
	if err := GetOrSet(ctx, c, "user:1", time.Minute, func(context.Context) (cachedUser, error) {
		return cachedUser{ID: 1, Name: "should-not-load"}, nil
	}, &got); err != nil {
		t.Fatalf("GetOrSet hit: %v", err)
	}
	if got.Name != "bob" {
		t.Errorf("cache hit returned %q, want cached \"bob\"", got.Name)
	}
}

func TestGetOrSetLoaderError(t *testing.T) {
	c, _ := newTestCache(t)
	ctx := context.Background()
	boom := errors.New("boom")

	loaderErr := func(context.Context) (cachedUser, error) { return cachedUser{}, boom }
	var got cachedUser
	if err := GetOrSet(ctx, c, "user:err", time.Minute, loaderErr, &got); !errors.Is(err, boom) {
		t.Errorf("GetOrSet loader error = %v, want boom", err)
	}

	// 失败的结果不应写入缓存
	if srv := c.client.Get(ctx, "user:err"); errors.Is(srv.Err(), nil) {
		t.Error("failed loader result should not be cached")
	}
}

func TestGetOrSetSingleflight(t *testing.T) {
	c, _ := newTestCache(t)
	ctx := context.Background()

	var loads atomic.Int64
	loader := func(context.Context) (cachedUser, error) { //nolint:unparam // singleflight 测试 loader 恒成功，错误路径由 TestGetOrSetLoaderError 覆盖
		loads.Add(1)
		time.Sleep(100 * time.Millisecond) // 保证并发请求重叠

		return cachedUser{ID: 7, Name: "carol"}, nil
	}

	const concurrency = 20
	var wg sync.WaitGroup
	errs := make([]error, concurrency)
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			var u cachedUser
			errs[idx] = GetOrSet(ctx, c, "user:7", time.Minute, loader, &u)
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("request %d failed: %v", i, err)
		}
	}
	if n := loads.Load(); n != 1 {
		t.Errorf("loader called %d times, want 1 (singleflight should merge concurrent loads)", n)
	}
}

func TestGetOrSetNilCache(t *testing.T) {
	var got cachedUser
	if err := GetOrSet(context.Background(), nil, "k", time.Minute,
		func(context.Context) (cachedUser, error) { return cachedUser{}, nil }, &got); err == nil {
		t.Error("GetOrSet with nil cache should fail")
	}
}

func TestGetOrSetWriteFailure(t *testing.T) {
	c, srv := newTestCache(t)
	ctx := context.Background()

	// 关闭 miniredis 模拟写失败，回源结果仍应返回给调用方
	srv.Close()

	var got cachedUser
	err := GetOrSet(ctx, c, "user:down", time.Minute,
		func(context.Context) (cachedUser, error) { return cachedUser{ID: 9, Name: "dave"}, nil }, &got)
	if err != nil {
		t.Errorf("GetOrSet should not fail when cache write fails, got %v", err)
	}
	if got.Name != "dave" {
		t.Errorf("GetOrSet result = %+v, want name=dave", got)
	}
}
