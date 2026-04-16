// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// NewProject 创建新项目
func NewProject(projectName string) error {
	// 验证项目名称
	if err := validateProjectName(projectName); err != nil {
		return err
	}

	// 检查目录是否已存在
	if _, err := os.Stat(projectName); err == nil {
		return fmt.Errorf("directory '%s' already exists", projectName)
	}

	fmt.Printf("Creating project: %s\n", projectName)

	// 创建项目目录结构
	if err := createProjectStructure(projectName); err != nil {
		return fmt.Errorf("failed to create project structure: %w", err)
	}

	// 生成 go.mod
	if err := generateGoMod(projectName); err != nil {
		return fmt.Errorf("failed to generate go.mod: %w", err)
	}

	// 生成主程序文件
	if err := generateMainFile(projectName); err != nil {
		return fmt.Errorf("failed to generate main.go: %w", err)
	}

	// 生成配置文件
	if err := generateConfigFile(projectName); err != nil {
		return fmt.Errorf("failed to generate config: %w", err)
	}

	// 生成 .gitignore
	if err := generateGitignore(projectName); err != nil {
		return fmt.Errorf("failed to generate .gitignore: %w", err)
	}

	// 生成 README
	if err := generateReadme(projectName); err != nil {
		return fmt.Errorf("failed to generate README.md: %w", err)
	}

	fmt.Println("\nProject structure created successfully!")
	return nil
}

// validateProjectName 验证项目名称
func validateProjectName(name string) error {
	if name == "" {
		return fmt.Errorf("project name cannot be empty")
	}

	// 检查是否包含非法字符
	if strings.ContainsAny(name, `/\:*?"<>|`) {
		return fmt.Errorf("project name contains invalid characters")
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
func generateGoMod(projectName string) error {
	content := fmt.Sprintf(`module %s

go 1.23

require (
	github.com/gin-gonic/gin v1.12.0
	github.com/AuroraLZDF/go-framework v0.1.0
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

	"github.com/AuroraLZDF/go-framework/core/config"
	"github.com/AuroraLZDF/go-framework/core/response"
	"github.com/AuroraLZDF/go-framework/core/server"
)

func main() {
	// 加载配置
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 创建应用
	app, err := server.NewApplication(cfg)
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}

	// 注册路由
	app.RegisterRoutes(func(engine *gin.Engine, db interface{}) {
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

app:
  name: ` + projectName + `
  version: 1.0.0
  mode: debug  # debug, release, test

server:
  http-addr: :8080
  https-addr: :8443
  tls-cert: ./cert/server.crt
  tls-key: ./cert/server.key

database:
  mysql:
    host: 127.0.0.1:3306
    username: root
    password: your_password
    database: ` + projectName + `_db
    max-idle-connections: 100
    max-open-connections: 20
    max-connection-life-time: 30s
    log-level: 4  # 1: silent, 2: error, 3: warn, 4: info

redis:
  addr: 127.0.0.1
  port: 6379
  password: ""
  db: 0

jwt:
  secret: change-this-to-a-random-secret-key-in-production
  expire: 24  # hours
  token-id: X-User-ID
  blacklist-path: tmp/blacklist.log

log:
  disable-caller: false
  disable-stacktrace: false
  level: info  # debug, info, warn, error
  format: json  # console, json
  dir: logs
`

	return writeFile(filepath.Join(projectName, "config.example.yaml"), content)
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

- Go 1.23+
- MySQL 5.7+ (optional)
- Redis 6.0+ (optional)

### Installation

1. Clone the repository
2. Install dependencies:
   ` + "```bash" + `
   go mod tidy
   ` + "```" + `

3. Copy and configure:
   ` + "```bash" + `
   cp config.example.yaml config.yaml
   # Edit config.yaml with your settings
   ` + "```" + `

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
