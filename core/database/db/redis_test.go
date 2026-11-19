// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package db

import (
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// resetRedisSingleton 将 Redis 单例置为 nil 并在测试结束后恢复原值，避免用例间全局状态污染.
func resetRedisSingleton(t *testing.T) {
	t.Helper()

	redisMu.Lock()
	prev := redisClient
	redisClient = nil
	redisMu.Unlock()

	t.Cleanup(func() {
		redisMu.Lock()
		redisClient = prev
		redisMu.Unlock()
	})
}

func TestInitRedisUnreachableAddr(t *testing.T) {
	resetRedisSingleton(t)

	// 本机未监听端口会被立即拒绝，用较小的拨号超时加速失败（无需真实 Redis）
	opts := &RedisOptions{Addr: "127.0.0.1:1", DialTimeout: 500 * time.Millisecond}

	client, err := InitRedis(opts)
	if err == nil {
		t.Fatal("连接不可达的 Redis 应返回错误")
	}
	if client != nil {
		t.Errorf("InitRedis 失败时返回的 client = %v, want nil", client)
	}
	if !strings.Contains(err.Error(), "redis ping failed") {
		t.Errorf("error = %v, 应包含 ping 失败的原因", err)
	}
	if GetRedis() != nil {
		t.Error("初始化失败不应缓存客户端")
	}

	// 失败不缓存错误，重试应再次尝试连接
	if _, err := InitRedis(opts); err == nil {
		t.Error("失败后重试 InitRedis 应再次返回错误")
	}
	if GetRedis() != nil {
		t.Error("重试失败后不应缓存客户端")
	}
}

func TestGetRedisReturnsSingleton(t *testing.T) {
	// 放置占位客户端（NewClient 懒连接，不会真正建连）
	redisMu.Lock()
	prev := redisClient
	redisClient = redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})
	redisMu.Unlock()
	t.Cleanup(func() {
		redisMu.Lock()
		redisClient = prev
		redisMu.Unlock()
	})

	if GetRedis() == nil {
		t.Fatal("GetRedis 应返回已初始化的客户端")
	}

	if err := CloseRedis(); err != nil {
		t.Fatalf("CloseRedis() error = %v", err)
	}
	if GetRedis() != nil {
		t.Error("CloseRedis 后单例应被重置")
	}
}

func TestCloseRedisWithoutInit(t *testing.T) {
	resetRedisSingleton(t)

	if err := CloseRedis(); err != nil {
		t.Fatalf("未初始化时 CloseRedis() error = %v, want nil", err)
	}
	if GetRedis() != nil {
		t.Error("未初始化时 GetRedis 应返回 nil")
	}
}
