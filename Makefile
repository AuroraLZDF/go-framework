.PHONY: build install clean test

# 构建脚手架工具
build:
	@echo "Building scaffold..."
	@go build -o bin/scaffold ./scaffold/cmd/main.go
	@echo "✅ Build successful: bin/scaffold"

# 安装脚手架工具到 GOPATH/bin
install:
	@echo "Installing scaffold..."
	@go install ./scaffold/cmd/main.go
	@echo "✅ Installed to $(GOPATH)/bin/scaffold"

# 清理构建文件
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@echo "✅ Cleaned"

# 测试脚手架
test: build
	@echo "Testing scaffold..."
	@./bin/scaffold new test-project
	@echo "✅ Test project created"
	@rm -rf test-project
	@echo "✅ Test completed"

# 默认目标
help:
	@echo "Framework Scaffold - Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  build    - Build the scaffold tool"
	@echo "  install  - Install scaffold to GOPATH/bin"
	@echo "  clean    - Remove build artifacts"
	@echo "  test     - Test the scaffold tool"
	@echo "  help     - Show this help message"
