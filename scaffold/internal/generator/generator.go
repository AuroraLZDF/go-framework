// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package generator

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// projectNamePattern 限制项目名：字母开头，仅允许字母、数字、'-'、'_'.
var projectNamePattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)

// generateStep 表示项目生成过程中的一个步骤.
// dir 为项目目录（可为路径），name 为校验过的项目名（用于模板内容）.
type generateStep struct {
	name string
	run  func(dir, name string) error
}

// projectSteps 按顺序执行；任一步失败会回滚整个项目目录.
var projectSteps = []generateStep{
	{"project structure", createProjectStructure},
	{"go.mod", generateGoMod},
	{"main.go", generateMainFile},
	{"config", generateConfigFile},
	{"OpenAPI example spec", generateOpenAPISpec},
	{"air config", generateAirConfig},
	{"Makefile", generateMakefile},
	{"Dockerfile", generateDockerfile},
	{"docker-compose", generateDockerCompose},
	{"CI workflow", generateCIWorkflow},
	{".env.example", generateEnvExample},
	{".gitignore", generateGitignore},
	{"README.md", generateReadme},
}

// NewProject 创建新项目.
// projectPath 可以是纯项目名（当前目录下创建），也可以是相对/绝对路径.
// 项目名取路径最后一段，用于 go.mod module、配置和 README 等模板内容.
// 生成过程中任一步失败都会删除已创建的项目目录，避免留下半成品.
func NewProject(projectPath string) error {
	// 验证项目名称（路径最后一段）
	name := filepath.Base(filepath.Clean(projectPath))
	if err := validateProjectName(name); err != nil {
		return err
	}

	// 检查目录是否已存在
	if _, err := os.Stat(projectPath); err == nil {
		return fmt.Errorf("directory '%s' already exists", projectPath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to check directory '%s': %w", projectPath, err)
	}

	fmt.Printf("Creating project: %s\n", projectPath)

	for _, step := range projectSteps {
		if err := step.run(projectPath, name); err != nil {
			// 失败回滚：清理已经创建的文件/目录
			if rmErr := os.RemoveAll(projectPath); rmErr != nil {
				return fmt.Errorf("failed to generate %s: %w (rollback failed: %w)", step.name, err, rmErr)
			}

			return fmt.Errorf("failed to generate %s: %w", step.name, err)
		}
	}

	fmt.Println("\nProject structure created successfully!")
	return nil
}

// validateProjectName 验证项目名称：必须以字母开头，仅包含字母、数字、'-'、'_'.
func validateProjectName(name string) error {
	if name == "" {
		return fmt.Errorf("project name cannot be empty")
	}

	if !projectNamePattern.MatchString(name) {
		return fmt.Errorf("invalid project name %q: must start with a letter and contain only letters, digits, '-' and '_'", name)
	}

	return nil
}

// createProjectStructure 创建项目目录结构
func createProjectStructure(dir, name string) error {
	dirs := []string{
		filepath.Join(dir, "cmd", "app"),
		filepath.Join(dir, "internal", "biz"),
		filepath.Join(dir, "internal", "controller"),
		filepath.Join(dir, "internal", "store"),
		filepath.Join(dir, "internal", "model"),
		filepath.Join(dir, "internal", "service"),
		filepath.Join(dir, "configs"),
		filepath.Join(dir, "sql"),
		filepath.Join(dir, "migrations"),
		filepath.Join(dir, "docs"),
		filepath.Join(dir, "assets"),
		filepath.Join(dir, "logs"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return err
		}
	}

	return nil
}

// generateGoMod 生成 go.mod 文件
// 版本需与脚手架模板使用的 API 保持一致（RouteRegisterFunc、jwt.blacklist-prefix 等）
func generateGoMod(dir, name string) error {
	content := fmt.Sprintf(`module %s

go 1.26.0

require (
	github.com/gin-gonic/gin v1.12.0
	github.com/AuroraLZDF/go-framework v0.5.0
	gorm.io/gorm v1.31.1
)
`, name)

	return writeFile(filepath.Join(dir, "go.mod"), content)
}

// generateMainFile 生成主程序文件
func generateMainFile(dir, name string) error {
	content := `package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/AuroraLZDF/go-framework/core/config"
	"github.com/AuroraLZDF/go-framework/core/response"
	"github.com/AuroraLZDF/go-framework/core/server"
)

func main() {
	// 1. 加载配置：默认读取当前目录下的 config.yaml，
	// 也可以通过 FRAMEWORK_* 环境变量覆盖（例如 FRAMEWORK_APP_MODE=release）
	cfg, err := config.Load("")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. 验证配置
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid config: %v", err)
	}

	// 3. 创建应用
	app, err := server.NewApplication(cfg)
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}

	// 注册路由
	app.RegisterRoutes(func(engine *gin.Engine, db *gorm.DB) {
		// 在线 API 文档（Swagger UI）：需要 docs/openapi.yaml，按需启用
		// if err := docs.Register(engine, docs.Options{SpecPath: "docs/openapi.yaml"}); err != nil {
		// 	log.Fatalf("Failed to register docs: %v", err)
		// }

		// 健康检查
		engine.GET("/health", func(c *gin.Context) {
			response.Success(c, gin.H{"status": "healthy"}, "ok")
		})

		// Hello World
		engine.GET("/", func(c *gin.Context) {
			response.Success(c, gin.H{
				"message": "Welcome to ` + name + `!",
				"version": cfg.App.Version,
			}, "ok")
		})

		// API v1 路由组
		v1 := engine.Group("/api/v1")
		{
			v1.GET("/users", func(c *gin.Context) {
				response.Success(c, []gin.H{
					{"id": 1, "name": "Alice"},
					{"id": 2, "name": "Bob"},
				}, "ok")
			})
		}
	})

	// 启动应用
	if err := app.Run(context.Background()); err != nil {
		log.Fatalf("Application failed: %v", err)
	}
}
`

	return writeFile(filepath.Join(dir, "cmd", "app", "main.go"), content)
}

// generateConfigFile 生成配置文件
func generateConfigFile(dir, name string) error {
	content := `# Application Configuration
# 所有配置项均可用 FRAMEWORK_ 前缀的环境变量覆盖，
# 例如 FRAMEWORK_APP_MODE=release、FRAMEWORK_LOG_LEVEL=debug

app:
  name: ` + name + `
  version: 1.0.0
  mode: debug  # debug, release, test

server:
  http-addr: :8080
  # 优雅关闭最长等待时间
  shutdown-timeout: 10s
  # HTTP 服务端超时（默认 10s / 30s / 60s / 120s）
  read-header-timeout: 10s
  read-timeout: 30s
  write-timeout: 60s
  idle-timeout: 120s
  # 请求体大小上限（字节），默认 4MB，超限返回 HTTP 413
  max-body-bytes: 4194304
  # 可信代理（IP 或 CIDR），默认不信任任何代理；部署在反向代理之后时按需配置
  # trusted-proxies:
  #   - 127.0.0.1
  # 启用 HTTPS 时必须同时提供证书和私钥，否则启动会失败；
  # 同时配置 http-addr 时该端口只做 308 重定向
  # https-addr: :8443
  # tls-cert: ./cert/server.crt
  # tls-key: ./cert/server.key

database:
  mysql:
    # 留空表示不连接 MySQL，按需填写
    host: ""
    username: root
    password: your_password
    database: ` + name + `_db
    max-idle-connections: 10
    max-open-connections: 100
    max-connection-life-time: 30s
    log-level: 4  # 1: silent, 2: error, 3: warn, 4: info

redis:
  # 留空表示不使用 Redis，按需填写
  addr: ""
  port: 6379
  password: ""
  db: 0
  # 连接池与超时，不填使用默认值 10 / 5s / 3s / 3s
  pool-size: 10
  dial-timeout: 5s
  read-timeout: 3s
  write-timeout: 3s

jwt:
  # 生产环境请替换为随机密钥（至少 32 字节，config.Validate 会校验）
  secret: change-this-to-a-random-secret-key-in-production
  expire: 24  # hours
  token-id: X-User-ID
  blacklist-prefix: "jwt:blacklist:"

log:
  disable-caller: false
  disable-stacktrace: false
  level: info  # debug, info, warn, error
  format: json  # console, json
  dir: logs

metrics:
  # Prometheus 指标暴露（/metrics），默认 true，设为 false 关闭
  enabled: true

rate-limit:
  # 按客户端 IP 的令牌桶限流，默认关闭；开启时 rate/burst 必须为正数
  enabled: false
  rate: 100    # 每秒允许的请求数
  burst: 200   # 突发容量
`

	// config.example.yaml 作为模板保留，config.yaml 直接可用
	if err := writeFile(filepath.Join(dir, "config.example.yaml"), content); err != nil {
		return err
	}

	return writeFile(filepath.Join(dir, "config.yaml"), content)
}

// generateAirConfig 生成 .air.toml 热更新配置
// （github.com/air-verse/air；make dev 在未安装 air 时用 go run 兜底）
func generateAirConfig(dir, name string) error {
	content := `# air 热更新配置（github.com/air-verse/air）
# 用法：make dev（优先使用已安装的 air；未安装时自动用 go run 兜底）
# 文档：https://github.com/air-verse/air

root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
cmd = "go build -o ./tmp/` + name + ` ./cmd/app"
bin = "./tmp/` + name + `"
# 框架自带优雅关闭（SIGTERM 后最多等 shutdown-timeout），
# air 默认 kill 等待太短，这里放宽到 3s
delay = 1000
stop_on_error = true
send_interrupt = true
kill_delay = 3000
# 只关心源码与配置变化，避免日志/临时文件触发重启
include_ext = ["go", "yaml", "yml"]
exclude_dir = ["assets", "tmp", "testdata", "bin", "logs", "vendor"]
exclude_file = []
include_file = []
exclude_regex = ["_test\\.go"]
exclude_unchanged = true
follow_symlink = true
args_bin = []
log = "build-errors.log"

[log]
time = true
main_only = false

[color]
main = "yellow,italic"
build = "yellow"
runner = "green"
watcher = "cyan"
build_error = "red"
build_warning = "yellow"

[misc]
# 退出时清理 tmp 目录
clean_on_exit = true
`

	return writeFile(filepath.Join(dir, ".air.toml"), content)
}

// generateMakefile 生成项目 Makefile（build/dev/test/lint 等）
func generateMakefile(dir, name string) error {
	content := `.PHONY: build run dev test lint cover clean help

APP_NAME := ` + name + `
AIR_PKG := github.com/air-verse/air
AIR_VERSION := v1.67.4

build: ## 编译应用到 tmp/
	@go build -o tmp/` + name + ` ./cmd/app
	@echo "✅ Built: tmp/` + name + `"

run: ## 直接运行（不带热更新）
	@go run ./cmd/app

dev: ## 热更新运行（air 优先，未安装时 go run 兜底）
	@if command -v air >/dev/null 2>&1; then \
		echo "🔥 air 热更新启动..."; \
		air -c .air.toml; \
	else \
		echo "⚠️  air 未安装，使用 go run 兜底（首次会自动下载 air $$(AIR_VERSION)）"; \
		echo "   提示：go install $$(AIR_PKG)@$$(AIR_VERSION) 可获得更快的热更新体验"; \
		go run $$(AIR_PKG)@$$(AIR_VERSION) -c .air.toml; \
	fi

test: ## 运行全部单元测试
	@go test ./...

lint: ## 运行 golangci-lint（需本地安装）
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "golangci-lint 未安装，请参考 https://golangci-lint.run/docs/welcome/install/"; \
		exit 1; \
	}
	@golangci-lint run ./...

cover: ## 运行测试并输出总覆盖率
	@go test -coverprofile=coverage.out -count=1 ./...
	@go tool cover -func=coverage.out | tail -1

clean: ## 清理构建产物
	@rm -rf tmp coverage.out
	@echo "✅ Cleaned tmp/ coverage.out"

help: ## 显示帮助
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@echo "  build   编译应用到 tmp/"
	@echo "  run     直接运行（不带热更新）"
	@echo "  dev     热更新运行（air 优先，go run 兜底）"
	@echo "  test    运行单元测试"
	@echo "  lint    运行 golangci-lint"
	@echo "  cover   测试覆盖率"
	@echo "  clean   清理 tmp/ coverage.out"
	@echo "  help    显示本帮助"
`

	return writeFile(filepath.Join(dir, "Makefile"), content)
}

// generateGitignore 生成 .gitignore 文件
func generateGitignore(dir, name string) error {
	content := `# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary
*.test

# Output of the go coverage tool
*.out

# Dependency directories
vendor/

# Go workspace file
go.work

# IDE
.idea/
.vscode/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db

# Logs
logs/
*.log

# Config (keep example, ignore actual)
config.yaml
!config.example.yaml

# Temporary files
tmp/
*.tmp

# Environment variables
.env
.env.local
`

	return writeFile(filepath.Join(dir, ".gitignore"), content)
}

// generateReadme 生成 README 文件
func generateReadme(dir, name string) error {
	content := `# ` + name + `

A web application built with Go Framework.

## Getting Started

### Prerequisites

- Go 1.26+
- MySQL 5.7+ (optional)
- Redis 6.0+ (optional)

### Installation

1. Clone the repository
2. Install dependencies:
   ` + "```bash" + `
   go mod tidy
   ` + "```" + `

3. 按需修改 ` + "`config.yaml`" + `（已自动生成，可直接启动；
   MySQL / Redis 留空即可不连接）

4. Run the application:
   ` + "```bash" + `
   # 开发模式（air 热更新；未安装 air 时自动用 go run 兜底）
   make dev

   # 或直接运行
   go run cmd/app/main.go
   ` + "```" + `
   安装 air 以获得更快的热更新体验：
   ` + "```bash" + `
   go install github.com/air-verse/air@v1.67.4
   ` + "```" + `

## Project Structure

` + "```" + `
` + name + `/
├── cmd/app/              # Application entry point
├── internal/
│   ├── biz/             # Business logic layer
│   ├── controller/      # HTTP handlers
│   ├── store/           # Data access layer
│   ├── model/           # Data models
│   └── service/         # External services
├── configs/             # Configuration files
├── sql/                 # Database scripts
├── migrations/          # Database migrations
├── assets/              # Static assets
├── logs/                # Log files
├── Dockerfile           # Multi-stage build
├── docker-compose.yaml  # App + MySQL + Redis for local dev
└── .github/workflows/   # CI (vet / test / build / docker)
` + "```" + `

## API Endpoints

- ` + "`GET /healthz`" + ` - Liveness probe
- ` + "`GET /readyz`" + ` - Readiness probe (checks MySQL/Redis)
- ` + "`GET /metrics`" + ` - Prometheus metrics
- ` + "`GET /`" + ` - Welcome message
- ` + "`GET /api/v1/users`" + ` - Get users list

## Deployment

` + "```bash" + `
# 本地开发（air 热更新）
make dev

# Docker 本地环境（应用 + MySQL + Redis）
docker compose up -d
` + "```" + `

## License

MIT License
`

	return writeFile(filepath.Join(dir, "README.md"), content)
}

// generateDockerfile 生成 Dockerfile（多阶段构建，最终镜像基于 distroless 风格的最小运行环境）
func generateDockerfile(dir, name string) error {
	content := `# 多阶段构建：编译与运行环境分离，最终镜像不含 Go 工具链
FROM golang:1.26-alpine AS builder

WORKDIR /app

# 先复制依赖清单，利用缓存加速重复构建
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/` + name + ` cmd/app/main.go

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata && \
    adduser -D -u 10001 appuser

WORKDIR /app
COPY --from=builder /app/bin/` + name + ` ./app
COPY --from=builder /app/config.example.yaml ./config.example.yaml

USER appuser

EXPOSE 8080

ENTRYPOINT ["./app"]
`

	return writeFile(filepath.Join(dir, "Dockerfile"), content)
}

// generateDockerCompose 生成 docker-compose.yaml（应用 + MySQL + Redis 开发环境）
func generateDockerCompose(dir, name string) error {
	content := `# 开发环境编排：应用 + MySQL + Redis
# 启动：docker compose up -d
services:
  ` + name + `:
    build: .
    ports:
      - "8080:8080"
    environment:
      FRAMEWORK_APP_MODE: release
      FRAMEWORK_DATABASE_MYSQL_HOST: mysql
      FRAMEWORK_DATABASE_MYSQL_DATABASE: ` + name + `_db
      FRAMEWORK_REDIS_ADDR: redis
      FRAMEWORK_JWT_SECRET: change-this-to-a-random-secret-key-in-production
    depends_on:
      mysql:
        condition: service_healthy
      redis:
        condition: service_started
    restart: unless-stopped

  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: root_password
      MYSQL_DATABASE: ` + name + `_db
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 5s
      timeout: 3s
      retries: 10

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    restart: unless-stopped

volumes:
  mysql_data:
  redis_data:
`

	return writeFile(filepath.Join(dir, "docker-compose.yaml"), content)
}

// generateCIWorkflow 生成 GitHub Actions CI 工作流
func generateCIWorkflow(dir, name string) error {
	content := `name: CI

on:
  push:
    branches: [main]
  pull_request:

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: "1.26.x"
          cache: true

      - name: Vet
        run: go vet ./...

      - name: Test
        run: go test -race -count=1 ./...

      - name: Build
        run: go build ./...

  docker:
    runs-on: ubuntu-latest
    needs: test
    steps:
      - uses: actions/checkout@v4

      - name: Build image
        run: docker build -t ` + name + `:ci .
`

	return writeFile(filepath.Join(dir, ".github", "workflows", "ci.yml"), content)
}

// generateEnvExample 生成 .env.example（docker-compose 与本地环境变量模板）
func generateEnvExample(dir, name string) error {
	content := `# 复制为 .env 使用（.env 已被 .gitignore 忽略，不要提交真实密钥）
# 所有配置项均可用 FRAMEWORK_ 前缀的环境变量覆盖

FRAMEWORK_APP_MODE=debug
FRAMEWORK_SERVER_HTTP_ADDR=:8080

FRAMEWORK_DATABASE_MYSQL_HOST=127.0.0.1
FRAMEWORK_DATABASE_MYSQL_USERNAME=root
FRAMEWORK_DATABASE_MYSQL_PASSWORD=your_password
FRAMEWORK_DATABASE_MYSQL_DATABASE=` + name + `_db

FRAMEWORK_REDIS_ADDR=127.0.0.1
FRAMEWORK_REDIS_PORT=6379

# 至少 32 字节
FRAMEWORK_JWT_SECRET=change-this-to-a-random-secret-key-in-production
`

	return writeFile(filepath.Join(dir, ".env.example"), content)
}

// writeFile 写入文件
func writeFile(path string, content string) error {
	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}

	// 写入文件
	return os.WriteFile(path, []byte(content), 0o600)
}
