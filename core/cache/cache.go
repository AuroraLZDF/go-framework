// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

// Package cache 提供基于 Redis 的类型化缓存封装：
//   - Get/Set/Delete 使用 JSON 作为统一序列化约定；
//   - GetOrSet 基于 singleflight 合并同一进程内的并发加载，
//     防止缓存击穿（同一个 key 只触发一次回源）。
//
// 使用示例：
//
//	c := cache.New(db.GetRedis())
//	var user User
//	err := cache.GetOrSet(ctx, c, "user:42", 10*time.Minute,
//	    func(ctx context.Context) (User, error) { return loadUser(ctx, 42) }, &user)
package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

// ErrNotFound 表示缓存中不存在目标键.
var ErrNotFound = errors.New("cache: key not found")

// ErrClientRequired 表示未提供 Redis 客户端.
var ErrClientRequired = errors.New("cache: redis client is required")

// ErrWriteFailed 表示缓存回写失败（GetOrSet 中视为非致命：结果仍会返回给调用方）.
var ErrWriteFailed = errors.New("cache: write to redis failed")

// Cache 基于 Redis 的类型化缓存.
type Cache struct {
	client redis.UniversalClient
	sf     singleflight.Group
}

// New 创建缓存实例，client 不能为 nil.
func New(client redis.UniversalClient) (*Cache, error) {
	if client == nil {
		return nil, ErrClientRequired
	}

	return &Cache{client: client}, nil
}

// Get 读取缓存并反序列化到 dst（dst 必须是指针）.
func (c *Cache) Get(ctx context.Context, key string, dst any) error {
	data, err := c.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("cache: get %s: %w", key, err)
	}

	if err := json.Unmarshal(data, dst); err != nil {
		return fmt.Errorf("cache: unmarshal %s: %w", key, err)
	}

	return nil
}

// Set 将 value JSON 序列化后写入缓存.
func (c *Cache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cache: marshal %s: %w", key, err)
	}

	if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("cache: set %s: %w", key, err)
	}

	return nil
}

// Delete 删除缓存键.
func (c *Cache) Delete(ctx context.Context, key string) error {
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("cache: delete %s: %w", key, err)
	}

	return nil
}

// GetOrSet 缓存未命中时调用 loader 回源，并合并同一 key 的并发加载（singleflight 防击穿）。
// 同一进程内针对同一 key 的多个并发请求只会触发一次 loader，
// 其余请求共享同一次结果；loader 的结果回写缓存后返回到 dst.
//
// 注意：singleflight 只保护单实例进程，多实例部署时同一 key 可能跨实例回源，
// 对一致性要求更高的场景请配合分布式锁或短 TTL 使用。
func GetOrSet[T any](ctx context.Context, c *Cache, key string, ttl time.Duration, loader func(context.Context) (T, error), dst *T) error {
	if c == nil {
		return ErrClientRequired
	}

	// 先查缓存，命中直接返回
	if err := c.Get(ctx, key, dst); err == nil {
		return nil
	}

	// 未命中：singleflight 合并并发加载
	result, err, _ := c.sf.Do(key, func() (any, error) {
		value, err := loader(ctx)
		if err != nil {
			return nil, err
		}

		if setErr := c.Set(ctx, key, value, ttl); setErr != nil {
			// 回写失败不阻塞业务：value 仍会返回给等待方，仅以哨兵错误标记
			return value, ErrWriteFailed
		}

		return value, nil
	})
	// 回写失败不影响调用方拿到回源结果
	if err != nil && !errors.Is(err, ErrWriteFailed) {
		return err
	}

	value, ok := result.(T)
	if !ok {
		return fmt.Errorf("cache: unexpected loader result type %T", result)
	}
	*dst = value

	return nil
}
