// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/auroralzdf/biddingAI.

package db

import (
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
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
	redisOnce   sync.Once
	redisErr    error
)

func InitRedis(opts *RedisOptions) (*redis.Client, error) {
	redisOnce.Do(func() {
		redisClient = redis.NewClient(&redis.Options{
			Addr:         opts.Addr,
			Password:     opts.Password,
			DB:           opts.DB,
			PoolSize:     opts.PoolSize,
			MinIdleConns: opts.MinIdleConns,
			DialTimeout:  opts.DialTimeout,
			ReadTimeout:  opts.ReadTimeout,
			WriteTimeout: opts.WriteTimeout,
		})

		// 测试连接
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := redisClient.Ping(ctx).Err(); err != nil {
			redisErr = err
		}
	})
	return redisClient, redisErr
}

// GetRedis 返回单例 Redis 客户端
func GetRedis() *redis.Client {
	return redisClient
}
