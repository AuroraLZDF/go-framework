.PHONY: build install clean test lint sec test-race cover check test-scaffold verify-scaffold release-tag help

# 版本信息：优先使用 git describe，未打 tag 时回退到最近一次发布版本
VERSION    ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo v0.2.0)
COMMIT     ?= $(shell git rev-parse HEAD 2>/dev/null || echo unknown)
BUILD_TIME ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_TREE   ?= $(shell if [ -z "$$(git status --porcelain 2>/dev/null)" ]; then echo clean; else echo dirty; fi)

VERSION_PKG := github.com/AuroraLZDF/go-framework/core/version

# 安装目录：优先 GOBIN，其次 GOPATH/bin
GOBIN_DIR := $(shell go env GOBIN)
ifeq ($(GOBIN_DIR),)
GOBIN_DIR := $(shell go env GOPATH)/bin
endif

LDFLAGS := -s -w \
	-X $(VERSION_PKG).GitVersion=$(VERSION) \
	-X $(VERSION_PKG).GitCommit=$(COMMIT) \
	-X $(VERSION_PKG).BuildDate=$(BUILD_TIME) \
	-X $(VERSION_PKG).GitTreeState=$(GIT_TREE)

build: ## 构建 scaffold 二进制到 bin/scaffold
	@echo "Building scaffold $(VERSION)..."
	@go build -ldflags "$(LDFLAGS)" -o bin/scaffold ./scaffold/cmd/main.go
	@echo "✅ Built: bin/scaffold"

install: ## 安装 scaffold 到 GOBIN（或 GOPATH/bin）
	@go install -ldflags "$(LDFLAGS)" ./scaffold/cmd/main.go
	@echo "✅ Installed to $(GOBIN_DIR)/scaffold"

clean: ## 清理构建产物
	@rm -rf bin
	@echo "✅ Cleaned bin/"

test: ## 运行全部单元测试
	@go test ./...

lint: ## 运行 golangci-lint（需本地安装，CI 使用 v2.6.2）
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "golangci-lint 未安装，请参考 https://golangci-lint.run/docs/welcome/install/"; \
		exit 1; \
	}
	@golangci-lint run ./...

sec: ## 运行 govulncheck 依赖漏洞扫描（需本地安装）
	@command -v govulncheck >/dev/null 2>&1 || { \
		echo "govulncheck 未安装，请执行 go install golang.org/x/vuln/cmd/govulncheck@v1.1.4"; \
		exit 1; \
	}
	@govulncheck ./...

test-race: ## 运行竞态检测测试
	@go test -race -count=1 ./...

cover: ## 运行测试并输出总覆盖率
	@go test -coverprofile=coverage.out -count=1 ./...
	@go tool cover -func=coverage.out | tail -1

check: lint test test-race ## 本地常用检查（lint + 单测 + 竞态）

test-scaffold: build ## 本地模式端到端验证脚手架（FRAMEWORK_LOCAL=1）
	@FRAMEWORK_LOCAL=1 ./scripts/verify-scaffold.sh

verify-scaffold: build ## 发布模式验证脚手架（要求最新 tag 存在）
	@./scripts/verify-scaffold.sh

release-tag: ## 打印发布打 tag 指引（不会实际创建 tag）
	@echo "Release $(VERSION)"
	@echo "  1. 确认 CHANGELOG.md / RELEASE_v0.2.0.md 已更新"
	@echo "  2. git tag -a v0.2.0 -m 'release: v0.2.0'"
	@echo "  3. git push origin main"
	@echo "  4. git push origin v0.2.0"
	@echo "打 tag 后，生成的脚手架项目才能通过 go mod tidy 拉取 v0.2.0。"

help: ## 显示帮助
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@echo "  build           构建 scaffold 到 bin/scaffold"
	@echo "  install         安装 scaffold 到 GOBIN/GOPATH/bin"
	@echo "  clean           清理 bin/"
	@echo "  test            运行单元测试 (go test ./...)"
	@echo "  lint            运行 golangci-lint"
	@echo "  sec             运行 govulncheck 依赖漏洞扫描"
	@echo "  test-race       运行竞态检测测试"
	@echo "  cover           运行测试并输出总覆盖率"
	@echo "  check           lint + test + test-race"
	@echo "  test-scaffold   本地模式端到端验证脚手架"
	@echo "  verify-scaffold 发布模式验证脚手架（需最新 release tag，如 v0.2.0）"
	@echo "  release-tag     打印发布指引"
	@echo "  help            显示本帮助"
