// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	jwtpkg "github.com/AuroraLZDF/go-framework/core/auth/jwt"
	"github.com/AuroraLZDF/go-framework/core/config"
	ctxkey "github.com/AuroraLZDF/go-framework/core/context"
	"github.com/AuroraLZDF/go-framework/core/database/db"
	"github.com/AuroraLZDF/go-framework/core/errno"
	"github.com/AuroraLZDF/go-framework/core/log"
	"github.com/AuroraLZDF/go-framework/core/middleware"
	"github.com/AuroraLZDF/go-framework/core/response"
)

// Application 应用框架结构体
type Application struct {
	config *config.Config
	engine *gin.Engine
	db     *gorm.DB

	closeOnce sync.Once
	closeErr  error
}

// RouteRegisterFunc 路由注册函数类型
type RouteRegisterFunc func(*gin.Engine, *gorm.DB)

// NewApplication 创建应用实例
func NewApplication(cfg *config.Config) (*Application, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// 先设置 Gin 模式再创建 Engine，避免 gin.New() 按默认模式打印调试信息
	gin.SetMode(cfg.App.Mode)

	// 将配置中的 JWT 参数接线到 jwt 包，使 middleware.Authn 可用
	if err := jwtpkg.Init(&jwtpkg.Options{
		TokenId:             cfg.JWT.TokenID,
		Secret:              cfg.JWT.Secret,
		Expire:              cfg.JWT.Expire,
		BlacklistPrefix:     cfg.JWT.BlacklistPrefix,
		BlacklistFailClosed: cfg.JWT.BlacklistFailClosed,
	}); err != nil {
		return nil, fmt.Errorf("failed to init jwt: %w", err)
	}

	app := &Application{config: cfg}

	// 初始化日志
	if err := app.initLog(); err != nil {
		jwtpkg.Reset()
		return nil, fmt.Errorf("failed to init log: %w", err)
	}

	app.engine = gin.New()

	// 安全默认：不信任任何代理（ClientIP 直接取 RemoteAddr）；仅在显式配置
	// server.trusted-proxies 后才采信 X-Forwarded-For / X-Real-IP。
	trustedProxies := app.config.Server.TrustedProxies
	if trustedProxies == nil {
		trustedProxies = []string{}
	}
	if err := app.engine.SetTrustedProxies(trustedProxies); err != nil {
		log.Sync()
		_ = log.Close()
		jwtpkg.Reset()
		return nil, fmt.Errorf("failed to set trusted proxies: %w", err)
	}
	for _, proxy := range trustedProxies {
		if _, network, err := net.ParseCIDR(proxy); err == nil {
			if ones, bits := network.Mask.Size(); ones == 0 {
				log.Warning("server.trusted-proxies contains an all-addresses CIDR; "+
					"any client can spoof X-Forwarded-For",
					"cidr", proxy, "prefix_len", ones, "bits", bits)
			}
		}
	}

	// 初始化数据库（未配置时跳过）
	if app.config.Database.MySQL.Host != "" {
		if err := app.initDatabase(); err != nil {
			// 日志文件已在 initLog 中打开，创建失败时需要递减全局日志引用计数，避免文件句柄泄漏
			log.Sync()
			_ = log.Close()
			jwtpkg.Reset()
			return nil, fmt.Errorf("failed to init database: %w", err)
		}
	}

	// 初始化 Redis（未配置时跳过）
	if app.config.Redis.Addr != "" {
		if err := app.initRedis(); err != nil {
			log.Warning("Failed to init redis, cache is unavailable", "err", err)
		}
	}

	// JWT 黑名单依赖 Redis；未配置 Redis 时黑名单校验会 fail-open，显式提示避免误以为可用
	if app.config.JWT.Secret != "" && app.config.Redis.Addr == "" {
		log.Warning("JWT token blacklist is unavailable because redis.addr is empty; revoked tokens will NOT be rejected (fail-open)")
	}

	// 注册通用中间件
	app.registerMiddlewares()

	log.Info("Application initialized successfully")

	return app, nil
}

// RegisterRoutes 注册业务路由
func (app *Application) RegisterRoutes(registerFunc RouteRegisterFunc) {
	if registerFunc != nil {
		registerFunc(app.engine, app.db)
	}
}

// Run 启动应用，并在收到信号或 ctx 取消后优雅退出
func (app *Application) Run(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	useTLS := app.config.Server.HTTPSAddr != ""

	// 主服务器：未启用 HTTPS 时监听 http-addr，启用后监听 https-addr
	primaryAddr := app.config.Server.HTTPAddr
	if useTLS {
		primaryAddr = app.config.Server.HTTPSAddr
	}
	primarySrv := app.createHTTPServer(primaryAddr)

	// 同时显式配置 http-addr 与 https-addr 时，HTTP 端口只负责 308 重定向到 HTTPS
	var redirectSrv *http.Server
	if useTLS && app.config.Server.HTTPAddr != "" {
		redirectSrv = app.createHTTPServer(app.config.Server.HTTPAddr)
		redirectSrv.Handler = app.httpsRedirectHandler()
	}

	// 先注册信号监听，再启动 HTTP 服务器，避免启动瞬间收到 SIGTERM 时被默认处理直接退出进程
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	// 启动 HTTP 服务器，启动失败通过 channel 返回，避免在 goroutine 中直接退出进程
	srvErr := make(chan error, 2)
	startServer := func(name string, srv *http.Server, tlsEnabled bool) {
		go func() {
			log.Info("Start to listening the incoming requests", "server", name, "addr", srv.Addr, "tls", tlsEnabled)

			var err error
			if tlsEnabled {
				err = srv.ListenAndServeTLS(app.config.Server.TLSCert, app.config.Server.TLSKey)
			} else {
				err = srv.ListenAndServe()
			}
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				srvErr <- fmt.Errorf("%s server: %w", name, err)
			}
		}()
	}

	startServer("primary", primarySrv, useTLS)
	if redirectSrv != nil {
		startServer("http-redirect", redirectSrv, false)
	}

	var runErr error
	select {
	case <-ctx.Done():
		log.Info("Context canceled, shutting down server ...")
	case <-quit:
		log.Info("Shutting down server ...")
	case runErr = <-srvErr:
		log.Error("HTTP server failed", "err", runErr)
	}

	// 创建上下文用于优雅关闭，超时时间可通过 server.shutdown-timeout 配置
	shutdownCtx, cancel := context.WithTimeout(context.Background(), app.config.Server.ShutdownTimeout)
	defer cancel()

	// 优雅关闭所有监听器；超时后强制关闭，避免进程无法退出
	shutdownServer := func(name string, srv *http.Server) {
		if srv == nil {
			return
		}
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Error("HTTP server graceful shutdown failed, forcing close", "server", name, "err", err)
			if closeErr := srv.Close(); closeErr != nil {
				log.Error("Failed to force close HTTP server", "server", name, "err", closeErr)
			}
			if runErr == nil {
				runErr = err
			}
		}
	}
	shutdownServer("primary", primarySrv)
	shutdownServer("http-redirect", redirectSrv)

	// 在关闭日志前记录最终日志，确保日志文件中保留退出信息
	log.Info("Server exiting")

	// 释放数据库、Redis、日志等资源
	if err := app.Close(); err != nil {
		if runErr == nil {
			runErr = err
		}
	}

	return runErr
}

// Close 释放应用占用的资源.
// 可重复调用，只会真正释放一次；先关闭数据库等业务资源，最后再关闭日志.
func (app *Application) Close() error {
	app.closeOnce.Do(func() {
		if err := db.Close(); err != nil {
			app.closeErr = err
		}

		log.Sync()
		if err := log.Close(); err != nil && app.closeErr == nil {
			app.closeErr = err
		}
	})

	return app.closeErr
}

// GetDB 获取数据库实例
func (app *Application) GetDB() *gorm.DB {
	return app.db
}

// GetEngine 获取 Gin 引擎
func (app *Application) GetEngine() *gin.Engine {
	return app.engine
}

// ==================== 私有方法 ====================

// initLog 初始化日志系统
func (app *Application) initLog() error {
	opts := &log.Options{
		DisableCaller:     app.config.Log.DisableCaller,
		DisableStacktrace: app.config.Log.DisableStacktrace,
		Level:             app.config.Log.Level,
		Format:            app.config.Log.Format,
		OutputPaths:       []string{"stdout"}, // 默认输出到标准输出
	}

	// 配置了日志目录时，同时按天轮转写入文件
	if dir := app.config.Log.Dir; dir != "" {
		basePath := app.config.App.Name
		if basePath == "" {
			basePath = "app"
		}

		opts.Rotation = &log.RotateConfig{
			Interval: "daily",
			BasePath: filepath.Join(dir, basePath),
			Stdout:   true,
		}
	}

	if err := log.Init(opts); err != nil {
		return fmt.Errorf("log init: %w", err)
	}
	return nil
}

// initDatabase 初始化数据库连接
func (app *Application) initDatabase() error {
	cfg := app.config.Database.MySQL

	mysqlOpts := &db.MySQLOptions{
		Host:                  cfg.Host,
		Username:              cfg.Username,
		Password:              cfg.Password,
		Database:              cfg.Database,
		MaxIdleConnections:    cfg.MaxIdleConnections,
		MaxOpenConnections:    cfg.MaxOpenConnections,
		MaxConnectionLifeTime: cfg.MaxConnectionLifeTime,
		LogLevel:              cfg.LogLevel,
	}

	var err error
	app.db, err = db.InitMySQL(mysqlOpts)
	if err != nil {
		return err
	}

	log.Info("Database connected successfully")
	return nil
}

// initRedis 初始化 Redis 连接
func (app *Application) initRedis() error {
	cfg := app.config.Redis

	if _, err := db.InitRedis(&db.RedisOptions{
		Addr:         redisServerAddr(cfg),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}); err != nil {
		return err
	}

	log.Info("Redis connected successfully")
	return nil
}

// registerMiddlewares 注册通用中间件
func (app *Application) registerMiddlewares() {
	// 先注册通用中间件，保证 404 与 panic 响应同样带有 RequestID、CORS 等头信息
	app.engine.Use(
		gin.CustomRecovery(func(c *gin.Context, recovered any) {
			log.Error("Panic recovered",
				"err", fmt.Sprint(recovered),
				"path", c.Request.URL.Path,
				"request_id", c.GetString(ctxkey.XRequestIDKey),
			)
			response.Error(c, errno.InternalServerError)
			c.Abort()
		}),
		middleware.RequestID(),
		// 访问日志依赖 RequestID 中间件，必须注册在 RequestID 之后
		accessLog(),
		// CORS 必须在 BodyLimit 之前注册，保证 413 响应同样带有跨域头
		middleware.CORS(),
		// 限制请求体大小，确保超大请求不会进入业务处理
		middleware.BodyLimit(app.config.Server.MaxBodyBytes),
	)

	// 注册 404 Handler，与业务错误使用统一的响应结构
	app.engine.NoRoute(func(c *gin.Context) {
		response.Error(c, errno.ErrPageNotFound)
	})

	// 注册健康检查：/healthz 仅表示进程存活（liveness），不检查外部依赖
	app.engine.GET("/healthz", func(c *gin.Context) {
		response.Success(c, gin.H{"status": "ok"}, "ok")
	})

	// /readyz 检查已配置的依赖（MySQL/Redis），用于负载均衡摘流与发布探针（readiness）
	app.engine.GET("/readyz", app.readinessHandler)
}

// createHTTPServer 创建 HTTP 服务器
func (app *Application) createHTTPServer(addr string) *http.Server {
	srv := app.config.Server

	return &http.Server{
		Addr:              addr,
		Handler:           app.engine,
		ReadHeaderTimeout: srv.ReadHeaderTimeout,
		ReadTimeout:       srv.ReadTimeout,
		WriteTimeout:      srv.WriteTimeout,
		IdleTimeout:       srv.IdleTimeout,
		MaxHeaderBytes:    srv.MaxHeaderBytes,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}
}

// redisServerAddr 把 redis.addr 规范化为 host:port.
// 兼容两种写法：addr 只写 host（配合 redis.port），或直接写 host:port.
func redisServerAddr(cfg config.RedisConfig) string {
	addr := strings.TrimSpace(cfg.Addr)
	if addr == "" {
		return ""
	}
	if _, _, err := net.SplitHostPort(addr); err == nil {
		return addr
	}
	// 兼容只写裸 IPv6 字面量的写法（如 "[::1]"），去掉方括号后由 JoinHostPort 重新加
	addr = strings.TrimPrefix(strings.TrimSuffix(addr, "]"), "[")

	port := cfg.Port
	if port == "" {
		port = "6379"
	}

	return net.JoinHostPort(addr, port)
}

// httpsRedirectHandler 返回把 HTTP 请求 308 重定向到 HTTPS 的 handler.
// 保留原始 path/query，并在 HTTPS 端口不是 443 时带上该端口.
func (app *Application) httpsRedirectHandler() http.Handler {
	httpsPort := ""
	if _, port, err := net.SplitHostPort(app.config.Server.HTTPSAddr); err == nil {
		httpsPort = port
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host == "" {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		// 取 Host 中的 hostname 部分（可能带端口或为 [IPv6] 字面量）
		host := r.Host
		if h, _, err := net.SplitHostPort(host); err == nil {
			host = h
		}
		// 去掉 IPv6 字面量可能残留的方括号，拼 Location 时再统一处理
		host = strings.TrimPrefix(strings.TrimSuffix(host, "]"), "[")

		// 308 默认可缓存；避免共享缓存把某个 Host 的重定向结果缓存后投毒
		w.Header().Set("Cache-Control", "no-store")

		locationHost := host
		if ip := net.ParseIP(host); ip != nil && ip.To4() == nil {
			// IPv6 字面量：Location 中必须带方括号
			if httpsPort != "" && httpsPort != "443" {
				locationHost = net.JoinHostPort(host, httpsPort)
			} else {
				locationHost = "[" + host + "]"
			}
		} else if httpsPort != "" && httpsPort != "443" {
			locationHost = host + ":" + httpsPort
		}

		//nolint:gosec // G710: 仅将同一 Host 的 HTTP 请求 308 重定向到 HTTPS，目标主机来自请求 Host 且已移除端口
		http.Redirect(w, r, "https://"+locationHost+r.URL.RequestURI(), http.StatusPermanentRedirect)
	})
}

// readinessHandler 检查已配置的外部依赖是否可用，失败返回 503.
func (app *Application) readinessHandler(c *gin.Context) {
	checks := gin.H{}
	healthy := true

	// MySQL：只有配置了 host 才会初始化，未配置视为跳过
	if app.config.Database.MySQL.Host != "" {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if app.db == nil {
			checks["mysql"] = "unavailable"
			healthy = false
		} else if sqlDB, err := app.db.DB(); err != nil {
			log.Error("Readiness check failed", "dependency", "mysql", "err", err)
			checks["mysql"] = "unavailable"
			healthy = false
		} else if err := sqlDB.PingContext(ctx); err != nil {
			log.Error("Readiness check failed", "dependency", "mysql", "err", err)
			checks["mysql"] = "unavailable"
			healthy = false
		} else {
			checks["mysql"] = "ok"
		}
	} else {
		checks["mysql"] = "skipped"
	}

	// Redis：只有配置了 addr 才会初始化，未配置视为跳过（独立超时，避免 MySQL 慢拖累判断）
	if app.config.Redis.Addr != "" {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		client := db.GetRedis()
		if client == nil {
			checks["redis"] = "unavailable"
			healthy = false
		} else if err := client.Ping(ctx).Err(); err != nil {
			log.Error("Readiness check failed", "dependency", "redis", "err", err)
			checks["redis"] = "unavailable"
			healthy = false
		} else {
			checks["redis"] = "ok"
		}
	} else {
		checks["redis"] = "skipped"
	}

	if !healthy {
		c.JSON(http.StatusServiceUnavailable, response.Response{
			Code:    errno.ErrServiceUnavailable.Code,
			Message: errno.ErrServiceUnavailable.Message,
			Content: checks,
		})
		return
	}

	response.Success(c, gin.H{"status": "ok", "checks": checks}, "ok")
}
