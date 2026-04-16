#!/bin/bash

# Framework 发布前最终检查脚本
# 运行此脚本确保项目已准备好发布

set -e

echo "=========================================="
echo "Framework v0.1.0-alpha 发布前最终检查"
echo "=========================================="
echo ""

PASS=0
FAIL=0
WARN=0

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

check_pass() {
    echo -e "${GREEN}✅ PASS${NC}: $1"
    ((PASS++))
}

check_fail() {
    echo -e "${RED}❌ FAIL${NC}: $1"
    ((FAIL++))
}

check_warn() {
    echo -e "${YELLOW}⚠️  WARN${NC}: $1"
    ((WARN++))
}

echo "1️⃣  检查 Git 状态..."
echo "-------------------------------------------"

# 检查是否有未提交的更改
if git diff --quiet && git diff --cached --quiet; then
    check_pass "工作目录干净，没有未提交的更改"
else
    check_warn "有未提交的更改（这可能是正常的，如果正准备提交）"
fi

# 检查是否在正确的分支
BRANCH=$(git branch --show-current)
if [ "$BRANCH" = "master" ] || [ "$BRANCH" = "main" ]; then
    check_pass "当前分支: $BRANCH"
else
    check_warn "当前分支: $BRANCH (建议在 master/main 分支发布)"
fi

echo ""
echo "2️⃣  检查代码质量..."
echo "-------------------------------------------"

# 编译检查
if go build ./... 2>/dev/null; then
    check_pass "所有模块编译通过"
else
    check_fail "编译失败"
fi

# go vet 检查
if go vet ./... 2>/dev/null; then
    check_pass "go vet 检查通过"
else
    check_fail "go vet 检查失败"
fi

# 检查是否有格式化问题
if [ -z "$(gofmt -l . | grep -v vendor/ | head -n 1)" ]; then
    check_pass "代码格式化正确"
else
    check_warn "部分文件未格式化 (运行: gofmt -s -w .)"
fi

echo ""
echo "3️⃣  检查必要文件..."
echo "-------------------------------------------"

REQUIRED_FILES=(
    "README.md"
    "LICENSE"
    "CHANGELOG.md"
    "go.mod"
    "core/server/application.go"
    "scaffold/cmd/main.go"
)

for file in "${REQUIRED_FILES[@]}"; do
    if [ -f "$file" ]; then
        check_pass "$file 存在"
    else
        check_fail "$file 缺失"
    fi
done

echo ""
echo "4️⃣  检查敏感信息..."
echo "-------------------------------------------"

# 检查 config.yaml 是否被忽略
if git check-ignore -q config.yaml 2>/dev/null; then
    check_pass "config.yaml 已被 .gitignore 忽略"
else
    check_fail "config.yaml 未被忽略（可能泄露敏感信息！）"
fi

# 检查是否有 .key 或 .pem 文件被跟踪
KEY_FILES=$(git ls-files | grep -E '\.(key|pem)$' || true)
if [ -z "$KEY_FILES" ]; then
    check_pass "没有密钥文件被跟踪"
else
    check_fail "发现密钥文件被跟踪: $KEY_FILES"
fi

# 检查 .env 文件
ENV_FILES=$(git ls-files | grep -E '^\.env$' || true)
if [ -z "$ENV_FILES" ]; then
    check_pass ".env 文件未被跟踪"
else
    check_fail ".env 文件被跟踪（可能泄露敏感信息！）"
fi

echo ""
echo "5️⃣  检查版本号..."
echo "-------------------------------------------"

# 检查 version.go 中的版本号
if grep -q "v0.1.0-alpha" core/version/version.go; then
    check_pass "版本号已设置为 v0.1.0-alpha"
else
    check_warn "版本号可能不是 v0.1.0-alpha"
fi

echo ""
echo "6️⃣  检查文档完整性..."
echo "-------------------------------------------"

DOCS=(
    "docs/quick-start.md"
    "docs/architecture.md"
    "scaffold/README.md"
    "RELEASE_NOTES.md"
)

for doc in "${DOCS[@]}"; do
    if [ -f "$doc" ]; then
        check_pass "$doc 存在"
    else
        check_warn "$doc 缺失"
    fi
done

echo ""
echo "7️⃣  检查脚手架工具..."
echo "-------------------------------------------"

if [ -f "bin/scaffold" ]; then
    check_pass "脚手架工具已构建"
    
    # 测试脚手架
    if ./bin/scaffold new test-release-check 2>/dev/null; then
        check_pass "脚手架工具运行正常"
        rm -rf test-release-check
    else
        check_fail "脚手架工具运行失败"
    fi
else
    check_warn "脚手架工具未构建 (运行: make build)"
fi

echo ""
echo "8️⃣  检查 .gitignore..."
echo "-------------------------------------------"

if [ -f ".gitignore" ]; then
    check_pass ".gitignore 文件存在"
    
    # 检查关键项是否被忽略
    IGNORE_CHECKS=("bin/" "config.yaml" "*.log" ".env")
    for pattern in "${IGNORE_CHECKS[@]}"; do
        if grep -q "^$pattern" .gitignore; then
            check_pass ".gitignore 包含: $pattern"
        else
            check_warn ".gitignore 可能缺少: $pattern"
        fi
    done
else
    check_fail ".gitignore 文件缺失"
fi

echo ""
echo "=========================================="
echo "检查结果汇总"
echo "=========================================="
echo -e "${GREEN}通过: $PASS${NC}"
echo -e "${RED}失败: $FAIL${NC}"
echo -e "${YELLOW}警告: $WARN${NC}"
echo ""

if [ $FAIL -eq 0 ]; then
    echo -e "${GREEN}✅ 所有关键检查通过！项目可以发布。${NC}"
    echo ""
    echo "下一步操作："
    echo "1. 审查所有警告项"
    echo "2. 提交所有更改: git add . && git commit -m 'release: v0.1.0-alpha'"
    echo "3. 创建标签: git tag -a v0.1.0-alpha -m 'Release v0.1.0-alpha'"
    echo "4. 推送: git push origin master && git push origin v0.1.0-alpha"
    echo "5. 在 GitHub 上创建 Release"
    exit 0
else
    echo -e "${RED}❌ 发现 $FAIL 个失败项，请修复后再发布。${NC}"
    exit 1
fi
