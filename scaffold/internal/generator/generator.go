// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

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
type generateStep struct {
	name string
	run  func(projectName string) error
}

// projectSteps 按顺序执行；任一步失败会回滚整个项目目录.
var projectSteps = []generateStep{
	{"project structure", createProjectStructure},
	{"go.mod", generateGoMod},
	{"main.go", generateMainFile},
	{"config", generateConfigFile},
	{".gitignore", generateGitignore},
	{"README.md", generateReadme},
}

// NewProject 创建新项目.
// 生成过程中任一步失败都会删除已创建的项目目录，避免留下半成品.
func NewProject(projectName string) error {
	// 验证项目名称
	if err := validateProjectName(projectName); err != nil {
		return err
	}

	// 检查目录是否已存在
	if _, err := os.Stat(projectName); err == nil {
		return fmt.Errorf("directory '%s' already exists", projectName)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to check directory '%s': %w", projectName, err)
	}

	fmt.Printf("Creating project: %s\n", projectName)

	for _, step := range projectSteps {
		if err := step.run(projectName); err != nil {
			// 失败回滚：清理已经创建的文件/目录
			if rmErr := os.RemoveAll(projectName); rmErr != nil {
				return fmt.Errorf("failed to generate %s: %w (rollback failed: %v)", step.name, err, rmErr)
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
func createProjectStructure(projectName string) error {
	dirs := []string{
		filepath.Join(projectName, "cmd", "app"),
		filepath.Join(projectName, "internal", "biz"),
		filepath.Join(projectName, "internal", "controller"),
		filepath.Join(projectName, "internal", "store"),
		filepath.Join(projectName, "internal", "model"),
		filepath.Join(projectName, "internal", "service"),
		filepath.Join(projectName, "configs"),
		filepath.Join(projectName, "sql"),
		filepath.Join(projectName, "assets"),
		filepath.Join(projectName, "logs"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}

// generateGoMod 生成 go.mod 文件
// 版本需与脚手架模板使用的 API 保持一致（RouteRegisterFunc、jwt.blacklist-prefix 等）
func generateGoMod(projectName string) error {
	content := fmt.Sprintf(`module %s

go 1.25.1

require (
	github.com/gin-gonic/gin v1.12.0
	github.com/AuroraLZDF/go-framework v0.1.3
	gorm.io/gorm v1.31.1
)
`, projectName)

	return writeFile(filepath.Join(projectName, "go.mod"), content)
}

// generateMainFile 生成主程序文件
func generateMainFile(projectName string) error {
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
		// 健康检查
		engine.GET("/health", func(c *gin.Context) {
			response.Success(c, gin.H{"status": "healthy"}, "ok")
		})

		// Hello World
		engine.GET("/", func(c *gin.Context) {
			response.Success(c, gin.H{
				"message": "Welcome to ` + projectName + `!",
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

	return writeFile(filepath.Join(projectName, "cmd", "app", "main.go"), content)
}

// generateConfigFile 生成配置文件
func generateConfigFile(projectName string) error {
	content := `# Application Configuration
# 所有配置项均可用 FRAMEWORK_ 前缀的环境变量覆盖，
# 例如 FRAMEWORK_APP_MODE=release、FRAMEWORK_LOG_LEVEL=debug

app:
  name: ` + projectName + `
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
  # 启用 HTTPS 时必须同时提供证书和私钥，否则启动会失败
  # https-addr: :8443
  # tls-cert: ./cert/server.crt
  # tls-key: ./cert/server.key

database:
  mysql:
    # 留空表示不连接 MySQL，按需填写
    host: ""
    username: root
    password: your_password
    database: ` + projectName + `_db
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
`

	// config.example.yaml 作为模板保留，config.yaml 直接可用
	if err := writeFile(filepath.Join(projectName, "config.example.yaml"), content); err != nil {
		return err
	}

	return writeFile(filepath.Join(projectName, "config.yaml"), content)
}

// generateGitignore 生成 .gitignore 文件
func generateGitignore(projectName string) error {
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

	return writeFile(filepath.Join(projectName, ".gitignore"), content)
}

// generateReadme 生成 README 文件
func generateReadme(projectName string) error {
	content := `# ` + projectName + `

A web application built with Go Framework.

## Getting Started

### Prerequisites

- Go 1.25.1+
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
   go run cmd/app/main.go
   ` + "```" + `

## Project Structure

` + "```" + `
` + projectName + `/
├── cmd/app/              # Application entry point
├── internal/
│   ├── biz/             # Business logic layer
│   ├── controller/      # HTTP handlers
│   ├── store/           # Data access layer
│   ├── model/           # Data models
│   └── service/         # External services
├── configs/             # Configuration files
├── sql/                 # Database scripts
├── assets/              # Static assets
└── logs/                # Log files
` + "```" + `

## API Endpoints

- ` + "`GET /health`" + ` - Health check
- ` + "`GET /`" + ` - Welcome message
- ` + "`GET /api/v1/users`" + ` - Get users list

## License

MIT License
`

	return writeFile(filepath.Join(projectName, "README.md"), content)
}

// writeFile 写入文件
func writeFile(path string, content string) error {
	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 写入文件
	return os.WriteFile(path, []byte(content), 0644)
}
