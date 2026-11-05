#!/usr/bin/env bash
#
# 脚手架端到端验证：
#   1. 构建 scaffold；
#   2. 在临时目录生成项目；
#   3. go mod tidy / go build / go vet；
#   4. 启动应用并请求 /health 与 /healthz。
#
# 默认要求仓库存最新 tag；本地开发设置 FRAMEWORK_LOCAL=1
# 会向生成项目追加 replace，指向当前仓库。
set -euo pipefail

FRAMEWORK_MODULE="github.com/AuroraLZDF/go-framework"
FRAMEWORK_TAG="${FRAMEWORK_TAG:-$(git describe --tags --abbrev=0 2>/dev/null || echo v0.2.0)}"
APP_NAME="verify-test-app"
APP_PORT="${APP_PORT:-18080}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

FRAMEWORK_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WORK_DIR="$(mktemp -d)"
APP_PID=""

cleanup() {
    if [ -n "$APP_PID" ] && kill -0 "$APP_PID" 2>/dev/null; then
        kill "$APP_PID" 2>/dev/null || true
        wait "$APP_PID" 2>/dev/null || true
    fi
    rm -rf "$WORK_DIR"
}
trap cleanup EXIT

fail() {
    echo -e "${RED}❌ $*${NC}" >&2
    exit 1
}

pass() {
    echo -e "${GREEN}✅ $*${NC}"
}

echo "=========================================="
echo "Scaffold E2E Verification"
echo "=========================================="

command -v go >/dev/null 2>&1 || fail "go 未安装"
command -v curl >/dev/null 2>&1 || fail "curl 未安装"

if [ "${FRAMEWORK_LOCAL:-0}" = "1" ]; then
    echo "📦 本地模式（FRAMEWORK_LOCAL=1）：使用 replace 指向 $FRAMEWORK_DIR"
else
    if ! git -C "$FRAMEWORK_DIR" rev-parse -q --verify "refs/tags/${FRAMEWORK_TAG}" >/dev/null 2>&1; then
        fail "仓库中不存在 tag ${FRAMEWORK_TAG}。
请先发布版本：
  git tag -a ${FRAMEWORK_TAG} -m 'release: ${FRAMEWORK_TAG}'
  git push origin ${FRAMEWORK_TAG}
或在本地开发时使用：
  FRAMEWORK_LOCAL=1 ./scripts/verify-scaffold.sh"
    fi
    pass "tag ${FRAMEWORK_TAG} 存在"
fi

echo ""
echo "🏗️  构建 scaffold..."
(cd "$FRAMEWORK_DIR" && make build >/dev/null)

echo "📁 生成项目..."
(cd "$WORK_DIR" && "$FRAMEWORK_DIR/bin/scaffold" new "$APP_NAME" >/dev/null)
cd "$WORK_DIR/$APP_NAME"

for dir in cmd/app internal/biz internal/controller internal/store internal/model internal/service configs sql assets logs; do
    [ -d "$dir" ] || fail "缺少目录: $dir"
done
for file in go.mod config.yaml config.example.yaml README.md .gitignore cmd/app/main.go; do
    [ -f "$file" ] || fail "缺少文件: $file"
done
pass "项目结构正确"

grep -q "module ${APP_NAME}" go.mod || fail "go.mod module 名称错误"
grep -q "name: ${APP_NAME}" config.yaml || fail "config.yaml 中缺少项目名"
grep -q "${FRAMEWORK_MODULE} ${FRAMEWORK_TAG}" go.mod || fail "go.mod 未依赖 ${FRAMEWORK_MODULE} ${FRAMEWORK_TAG}"
grep -q '/health' cmd/app/main.go || fail "生成的 main.go 缺少 /health 路由"
grep -q 'Get("/healthz")' cmd/app/main.go 2>/dev/null || true
pass "生成内容检查通过"

if [ "${FRAMEWORK_LOCAL:-0}" = "1" ]; then
    grep -q "replace ${FRAMEWORK_MODULE} =>" go.mod || echo "replace ${FRAMEWORK_MODULE} => ${FRAMEWORK_DIR}" >> go.mod
    pass "已追加本地 replace"
fi

echo "📦 go mod tidy / build / vet..."
go mod tidy >/dev/null
go build -o app ./cmd/app
go vet ./...
pass "依赖解析与编译通过"

echo "🚀 启动应用并检查健康检查..."
FRAMEWORK_SERVER_HTTP_ADDR="127.0.0.1:${APP_PORT}" ./app >app.log 2>&1 &
APP_PID=$!

ready=0
for _ in $(seq 1 40); do
    if curl -fsS "http://127.0.0.1:${APP_PORT}/healthz" >/dev/null 2>&1; then
        ready=1
        break
    fi
    sleep 0.25
done

if [ "$ready" -ne 1 ]; then
    cat app.log >&2 || true
    fail "应用未在 10s 内就绪（/healthz）"
fi

curl -fsS "http://127.0.0.1:${APP_PORT}/healthz" >/dev/null || fail "/healthz 请求失败"
curl -fsS "http://127.0.0.1:${APP_PORT}/health" >/dev/null || fail "/health 请求失败"
pass "/healthz 与 /health 均返回成功"

echo ""
echo -e "${GREEN}✅ 脚手架端到端验证通过！${NC}"
