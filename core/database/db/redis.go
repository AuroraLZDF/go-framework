// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/auroralzdf/biddingAI.

package db

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// defaultRedisPoolSize 默认连接池大小
	defaultRedisPoolSize = 10
	// defaultRedisDialTimeout 默认建连超时
	defaultRedisDialTimeout = 5 * time.Second
	// defaultRedisReadTimeout 默认读超时
	defaultRedisReadTimeout = 3 * time.Second
	// defaultRedisWriteTimeout 默认写超时
	defaultRedisWriteTimeout = 3 * time.Second
)

// RedisOptions 定义 Redis 连接选项
type RedisOptions struct {
	Addr         string        // "127.0.0.1:6379"
	Password     string        // 密码
	DB           int           // 数据库
	PoolSize     int           // 连接池大小
	MinIdleConns int           // 最小空闲连接数
	DialTimeout  time.Duration // 连接超时
	ReadTimeout  time.Duration // 读超时
	WriteTimeout time.Duration // 写超时
}

var (
	redisClient *redis.Client
	redisMu     sync.Mutex
)

// InitRedis 初始化 Redis 连接.
// 初始化成功后复用该实例；失败时不会缓存错误，允许调用方重试.
// CloseRedis 之后可以再次调用 InitRedis 重新建立连接.
func InitRedis(opts *RedisOptions) (*redis.Client, error) {
	if opts == nil {
		return nil, errors.New("db: redis options is nil")
	}

	redisMu.Lock()
	defer redisMu.Unlock()

	if redisClient != nil {
		return redisClient, nil
	}

	poolSize := opts.PoolSize
	if poolSize <= 0 {
		poolSize = defaultRedisPoolSize
	}

	dialTimeout := opts.DialTimeout
	if dialTimeout <= 0 {
		dialTimeout = defaultRedisDialTimeout
	}

	readTimeout := opts.ReadTimeout
	if readTimeout <= 0 {
		readTimeout = defaultRedisReadTimeout
	}

	writeTimeout := opts.WriteTimeout
	if writeTimeout <= 0 {
		writeTimeout = defaultRedisWriteTimeout
	}

	minIdleConns := opts.MinIdleConns
	if minIdleConns < 0 {
		minIdleConns = 0
	}
	if minIdleConns > poolSize {
		minIdleConns = poolSize
	}

	client := redis.NewClient(&redis.Options{
		Addr:         opts.Addr,
		Password:     opts.Password,
		DB:           opts.DB,
		PoolSize:     poolSize,
		MinIdleConns: minIdleConns,
		DialTimeout:  dialTimeout,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	})

	// 用 DialTimeout 作为 Ping 的上界，避免使用与配置无关的硬编码超时
	ctx, cancel := context.WithTimeout(context.Background(), dialTimeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		// 连接不可用，关闭并返回错误，不保留不可用的客户端
		_ = client.Close()

		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	redisClient = client

	return redisClient, nil
}

// GetRedis 返回单例 Redis 客户端
func GetRedis() *redis.Client {
	redisMu.Lock()
	defer redisMu.Unlock()

	return redisClient
}

// CloseRedis 关闭共享 Redis 客户端并重置单例，便于重新初始化.
// 供持有同一客户端的组件（如 util.Cache）释放连接时使用。
// 未初始化时调用是安全的，会返回 nil.
func CloseRedis() error {
	return closeRedis()
}

// closeRedis 关闭 Redis 连接并重置单例，便于重新初始化
func closeRedis() error {
	redisMu.Lock()
	client := redisClient
	redisClient = nil
	redisMu.Unlock()

	if client == nil {
		return nil
	}

	if err := client.Close(); err != nil {
		return fmt.Errorf("close redis: %w", err)
	}

	return nil
}
