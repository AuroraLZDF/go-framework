#!/bin/bash

# 脚手架工具验证脚本

set -e

echo "================================"
echo "Framework Scaffold 验证测试"
echo "================================"
echo ""

# 1. 构建脚手架
echo "1️⃣  构建脚手架工具..."
make build
if [ $? -eq 0 ]; then
    echo "✅ 构建成功"
else
    echo "❌ 构建失败"
    exit 1
fi
echo ""

# 2. 创建测试项目
echo "2️⃣  创建测试项目..."
./bin/scaffold new verify-test-app
if [ $? -eq 0 ]; then
    echo "✅ 项目创建成功"
else
    echo "❌ 项目创建失败"
    exit 1
fi
echo ""

# 3. 验证项目结构
echo "3️⃣  验证项目结构..."
if [ -d "verify-test-app/cmd/app" ]; then
    echo "✅ cmd/app 目录存在"
else
    echo "❌ cmd/app 目录缺失"
    exit 1
fi

if [ -d "verify-test-app/internal/biz" ]; then
    echo "✅ internal/biz 目录存在"
else
    echo "❌ internal/biz 目录缺失"
    exit 1
fi

if [ -f "verify-test-app/go.mod" ]; then
    echo "✅ go.mod 文件存在"
else
    echo "❌ go.mod 文件缺失"
    exit 1
fi

if [ -f "verify-test-app/config.example.yaml" ]; then
    echo "✅ config.example.yaml 文件存在"
else
    echo "❌ config.example.yaml 文件缺失"
    exit 1
fi

if [ -f "verify-test-app/README.md" ]; then
    echo "✅ README.md 文件存在"
else
    echo "❌ README.md 文件缺失"
    exit 1
fi
echo ""

# 4. 验证配置文件内容
echo "4️⃣  验证配置文件内容..."
if grep -q "verify-test-app" verify-test-app/config.example.yaml; then
    echo "✅ 项目名称已正确配置"
else
    echo "❌ 项目名称配置错误"
    exit 1
fi
echo ""

# 5. 验证主程序文件
echo "5️⃣  验证主程序文件..."
if grep -q "verify-test-app" verify-test-app/cmd/app/main.go; then
    echo "✅ 主程序包含正确的项目名称"
else
    echo "❌ 主程序项目名称错误"
    exit 1
fi
echo ""

# 6. 清理测试项目
echo "6️⃣  清理测试项目..."
rm -rf verify-test-app
echo "✅ 清理完成"
echo ""

echo "================================"
echo "✅ 所有验证测试通过！"
echo "================================"
echo ""
echo "脚手架工具已准备就绪，可以开始使用："
echo "  ./bin/scaffold new my-app"
echo ""
