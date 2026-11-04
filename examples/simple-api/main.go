// 简单 API 示例 - 展示如何使用 framework
package main

import (
	"context"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/AuroraLZDF/go-framework/core/config"
	"github.com/AuroraLZDF/go-framework/core/middleware"
	"github.com/AuroraLZDF/go-framework/core/response"
	"github.com/AuroraLZDF/go-framework/core/server"
)

func main() {
	// 加载示例配置（默认不连接 MySQL/Redis，可直接启动）
	cfg, err := config.Load("config.example.yaml")
	if err != nil {
		panic(err)
	}

	// 创建应用
	app, err := server.NewApplication(cfg)
	if err != nil {
		panic(err)
	}

	// 注册路由
	app.RegisterRoutes(func(engine *gin.Engine, db *gorm.DB) {
		// 公开路由
		engine.GET("/", func(c *gin.Context) {
			response.Success(c, gin.H{
				"message": "Welcome to Framework!",
				"version": cfg.App.Version,
			}, "ok")
		})

		// API v1 路由组
		v1 := engine.Group("/api/v1")
		{
			v1.GET("/health", func(c *gin.Context) {
				response.Success(c, gin.H{"status": "healthy"}, "ok")
			})

			v1.GET("/users", func(c *gin.Context) {
				response.Success(c, []gin.H{
					{"id": 1, "name": "Alice"},
					{"id": 2, "name": "Bob"},
				}, "ok")
			})

			// 受保护路由示例：请求头需携带 Authorization: Bearer <token>.
			// token 由 core/auth/jwt 签发（见 middleware.Authn 的实现），
			// 校验失败返回 401；未配置 Redis 时黑名单校验按配置 fail-open.
			v1.GET("/profile", middleware.Authn(), func(c *gin.Context) {
				response.Success(c, gin.H{"message": "authenticated user only"}, "ok")
			})
		}
	})

	// 启动应用
	if err := app.Run(context.Background()); err != nil {
		panic(err)
	}
}
