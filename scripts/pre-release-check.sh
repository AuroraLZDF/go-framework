#!/usr/bin/env bash
#
# 发布前检查脚本（只读检查，不修改仓库、不创建 tag）。
# 用法：./scripts/pre-release-check.sh
set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

PASS=0
FAIL=0
WARN=0

pass() {
    echo -e "${GREEN}✅ PASS${NC}: $*"
    PASS=$((PASS + 1))
}

fail() {
    echo -e "${RED}❌ FAIL${NC}: $*"
    FAIL=$((FAIL + 1))
}

warn() {
    echo -e "${YELLOW}⚠️  WARN${NC}: $*"
    WARN=$((WARN + 1))
}

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

echo "=========================================="
echo "Framework Pre-Release Check"
echo "=========================================="
echo ""

# ---------------------------------------------------------------------------
# 1. 版本信息（以 git tag 为唯一来源）
# ---------------------------------------------------------------------------
echo "📌 版本检查"

TAG="$(git describe --tags --abbrev=0 2>/dev/null || true)"
if [ -z "$TAG" ]; then
    fail "未找到任何 git tag，无法确定版本"
    TAG="v0.2.0"
else
    pass "最新 tag: $TAG"
fi

FILE_VERSION="$(sed -n 's/.*GitVersion = "\([^"]*\)".*/\1/p' core/version/version.go | head -1)"
if [ "$FILE_VERSION" = "$TAG" ]; then
    pass "core/version/version.go GitVersion = $TAG"
else
    fail "version.go GitVersion='$FILE_VERSION' 与最新 tag '$TAG' 不一致"
fi

CURRENT_BRANCH="$(git branch --show-current 2>/dev/null || true)"
case "$CURRENT_BRANCH" in
    main|master)
        pass "当前分支: $CURRENT_BRANCH"
        ;;
    "")
        warn "未处于分支（detached HEAD）"
        ;;
    *)
        warn "当前分支为 '$CURRENT_BRANCH'，发布应使用 main"
        ;;
esac

if [ -z "$(git status --porcelain)" ]; then
    pass "工作区干净"
else
    fail "工作区存在未提交改动，请先提交"
fi

echo ""

# ---------------------------------------------------------------------------
# 2. 构建与静态检查
# ---------------------------------------------------------------------------
echo "🔨 构建与静态检查"

if go build ./... 2>/dev/null; then
    pass "go build ./..."
else
    fail "go build ./... 失败"
fi

if go vet ./... 2>/dev/null; then
    pass "go vet ./..."
else
    fail "go vet ./... 失败"
fi

UNFORMATTED="$(gofmt -l . 2>/dev/null | grep -v '^vendor/' || true)"
if [ -z "$UNFORMATTED" ]; then
    pass "gofmt 检查通过"
else
    fail "以下文件未格式化：$UNFORMATTED"
fi

echo ""

# ---------------------------------------------------------------------------
# 3. 必需文件与文档
# ---------------------------------------------------------------------------
echo "📄 文件检查"

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
        pass "存在 $file"
    else
        fail "缺少 $file"
    fi
done

RELEASE_DOC="RELEASE_${TAG}.md"
if [ -f "$RELEASE_DOC" ]; then
    pass "存在发布说明 $RELEASE_DOC"
else
    warn "缺少 $RELEASE_DOC（建议在发布前补充）"
fi

DOCS=("docs/quick-start.md" "docs/architecture.md" "docs/CONFIG_USAGE.md")
for doc in "${DOCS[@]}"; do
    if [ -f "$doc" ]; then
        pass "存在 $doc"
    else
        fail "缺少 $doc"
    fi
done

echo ""

# ---------------------------------------------------------------------------
# 4. 敏感文件检查
# ---------------------------------------------------------------------------
echo "🔒 敏感文件检查"

if git check-ignore -q config.yaml 2>/dev/null; then
    pass "config.yaml 已被 .gitignore 忽略"
else
    warn "config.yaml 未被忽略，确认没有提交本地配置"
fi

if git ls-files --error-unmatch .env >/dev/null 2>&1; then
    fail ".env 被 Git 跟踪"
else
    pass ".env 未被跟踪"
fi

TRACKED_SECRETS="$(git ls-files | grep -E '\.(key|pem)$' || true)"
if [ -z "$TRACKED_SECRETS" ]; then
    pass "未跟踪 .key/.pem 文件"
else
    fail "存在被跟踪的密钥文件：$TRACKED_SECRETS"
fi

echo ""

# ---------------------------------------------------------------------------
# 5. .gitignore 检查
# ---------------------------------------------------------------------------
echo "📝 .gitignore 检查"

for pattern in "bin/" "config.yaml" "*.log" ".env"; do
    if grep -qF "$pattern" .gitignore; then
        pass ".gitignore 包含 $pattern"
    else
        fail ".gitignore 缺少 $pattern"
    fi
done

echo ""

# ---------------------------------------------------------------------------
# 6. 脚手架快速验证（在临时目录执行，不写入仓库）
# ---------------------------------------------------------------------------
echo "🏗️  脚手架检查"

if [ -x "bin/scaffold" ] || [ -f "bin/scaffold" ]; then
    pass "bin/scaffold 存在"
else
    warn "bin/scaffold 不存在，跳过快速验证（可先 make build）"
fi

if [ -x "bin/scaffold" ]; then
    TMP_DIR="$(mktemp -d)"
    cleanup() { rm -rf "$TMP_DIR"; }
    trap cleanup EXIT

    if (cd "$TMP_DIR" && "$ROOT_DIR/bin/scaffold" new test-release-check >/dev/null 2>&1); then
        if [ -f "$TMP_DIR/test-release-check/go.mod" ] && [ -f "$TMP_DIR/test-release-check/config.yaml" ]; then
            pass "脚手架生成项目结构正确（临时目录）"
        else
            fail "脚手架生成的 go.mod/config.yaml 缺失"
        fi
    else
        fail "脚手架生成项目失败"
    fi
fi

echo ""

# ---------------------------------------------------------------------------
# 7. 汇总
# ---------------------------------------------------------------------------
echo "=========================================="
echo "检查结果"
echo "=========================================="
echo -e "${GREEN}PASS: $PASS${NC}  ${RED}FAIL: $FAIL${NC}  ${YELLOW}WARN: $WARN${NC}"
echo ""

if [ "$FAIL" -gt 0 ]; then
    echo -e "${RED}发布检查未通过，请修复后重试。${NC}"
    exit 1
fi

echo -e "${GREEN}✅ 发布检查通过！${NC}"
echo ""
echo "后续步骤（需手动执行）："
echo "  1. git tag -a ${TAG} -m 'release: ${TAG}'"
echo "  2. git push origin ${CURRENT_BRANCH:-main}"
echo "  3. git push origin ${TAG}"
