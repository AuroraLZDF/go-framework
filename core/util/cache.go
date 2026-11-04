// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/auroralzdf/biddingAI.

package util

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/AuroraLZDF/go-framework/core/database/db"
)

// ErrCacheUnavailable 表示 Redis 客户端尚未初始化
var ErrCacheUnavailable = errors.New("redis client is not initialized")

// Cache 缓存接口
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Close()
}

// cache 缓存实现
type cache struct {
	client *redis.Client
}

// NewCache 创建 Redis 缓存
func NewCache() Cache {
	return &cache{
		client: db.GetRedis(),
	}
}

func (c *cache) Get(ctx context.Context, key string) (string, error) {
	if c.client == nil {
		return "", ErrCacheUnavailable
	}

	return c.client.Get(ctx, key).Result()
}

func (c *cache) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	if c.client == nil {
		return ErrCacheUnavailable
	}

	return c.client.Set(ctx, key, value, expiration).Err()
}

func (c *cache) Delete(ctx context.Context, key string) error {
	if c.client == nil {
		return ErrCacheUnavailable
	}

	return c.client.Del(ctx, key).Err()
}

func (c *cache) Exists(ctx context.Context, key string) (bool, error) {
	if c.client == nil {
		return false, ErrCacheUnavailable
	}

	n, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return n > 0, nil
}

func (c *cache) Close() {
	if c.client != nil {
		// 客户端由 db 包共享，关闭时必须同时重置单例，否则后续请求会拿到已关闭的连接
		_ = db.CloseRedis()
	}
}
