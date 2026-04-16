// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/auroralzdf/biddingAI.

package util

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/AuroraLZDF/go-framework/core/database/db"
)

// Cache 缓存接口
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Close()
}

// Cache 缓存实现
type cache struct {
	client *redis.Client
}

// NewCache 创建 Redis 缓存
func NewCache() *cache {
	return &cache{
		client: db.GetRedis(),
	}
}

func (c *cache) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c *cache) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

func (c *cache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

func (c *cache) Close() {
	err := c.client.Close()
	if err != nil {
		return
	}
}
