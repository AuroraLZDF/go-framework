// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/AuroraLZDF/go-framework/core/config"
	"github.com/AuroraLZDF/go-framework/core/database/db"
	"github.com/AuroraLZDF/go-framework/core/log"
	"github.com/AuroraLZDF/go-framework/core/middleware"
)

// Application 应用框架结构体
type Application struct {
	config *config.Config
	engine *gin.Engine
	db     *gorm.DB
}

// RouteRegisterFunc 路由注册函数类型
type RouteRegisterFunc func(*gin.Engine, interface{})

// NewApplication 创建应用实例
func NewApplication(cfg *config.Config) (*Application, error) {
	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	app := &Application{
		config: cfg,
		engine: gin.New(),
	}

	// 初始化日志
	if err := app.initLog(); err != nil {
		return nil, fmt.Errorf("failed to init log: %w", err)
	}

	// 初始化数据库
	if err := app.initDatabase(); err != nil {
		return nil, fmt.Errorf("failed to init database: %w", err)
	}

	// 设置 Gin 模式
	gin.SetMode(cfg.App.Mode)

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

// Run 启动应用
func (app *Application) Run(ctx context.Context) error {
	// 创建 HTTP 服务器
	httpSrv := app.createHTTPServer()

	// 启动 HTTP 服务器
	go func() {
		log.Info("Start to listening the incoming requests on http address", "addr", app.config.Server.HTTPAddr)
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("HTTP server failed", "err", err)
		}
	}()

	// 等待中断信号优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server ...")

	// 创建上下文用于优雅关闭（10秒超时）
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 优雅关闭 HTTP 服务器
	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		log.Error("HTTP server forced to shutdown", "err", err)
		return err
	}

	log.Info("Server exiting")
	return nil
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

	log.Init(opts)
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

// registerMiddlewares 注册通用中间件
func (app *Application) registerMiddlewares() {
	// 注册 404 Handler
	app.engine.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"code":    404,
			"message": "Page not found",
		})
	})

	// 注册健康检查
	app.engine.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 使用通用中间件
	app.engine.Use(
		gin.Logger(),
		gin.Recovery(),
		middleware.RequestID(),
		middleware.CORS(),
	)
}

// createHTTPServer 创建 HTTP 服务器
func (app *Application) createHTTPServer() *http.Server {
	return &http.Server{
		Addr:    app.config.Server.HTTPAddr,
		Handler: app.engine,
	}
}
