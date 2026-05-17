#!/usr/bin/env bash
set -euo pipefail
BASE_DIR="$(cd "$(dirname "$0")" && pwd)"

if [[ "$(uname -s)" != "Darwin" ]]; then
  echo "这个包是 macOS 版本。当前系统: $(uname -s)"
  exit 1
fi

mkdir -p "$BASE_DIR"/{config,data,logs,service,bin}

if [[ ! -f "$BASE_DIR/config/config.yaml" && -f "$BASE_DIR/config/config.yaml.example" ]]; then
  cp "$BASE_DIR/config/config.yaml.example" "$BASE_DIR/config/config.yaml"
fi

ARCH="$(uname -m)"
case "$ARCH" in
  arm64|aarch64) MIHOMO_BIN="$BASE_DIR/bin/mihomo-darwin-arm64" ;;
  *) MIHOMO_BIN="" ;;
esac

chmod +x "$BASE_DIR/bin/mihomo-tui" "$BASE_DIR/bin/mihomo-darwin-arm64" "$BASE_DIR/bin/ssr_2_mihomo" "$BASE_DIR/bin/mihomo_ctl" "$BASE_DIR/bin/fetch-mihomo.sh" 2>/dev/null || true

if [[ -f "$BASE_DIR/cmd/ssr_2_mihomo/main.go" ]] && [[ ! -x "$BASE_DIR/bin/ssr_2_mihomo" || "$BASE_DIR/cmd/ssr_2_mihomo/main.go" -nt "$BASE_DIR/bin/ssr_2_mihomo" ]] && command -v go >/dev/null 2>&1; then
  echo "构建 Go SSR 转换器..."
  if ! go build -o "$BASE_DIR/bin/ssr_2_mihomo" "$BASE_DIR/cmd/ssr_2_mihomo/main.go"; then
    echo "Go SSR 转换器构建失败，稍后可手动执行："
    echo "  go build -o bin/ssr_2_mihomo cmd/ssr_2_mihomo/main.go"
  fi
fi

if [[ -f "$BASE_DIR/cmd/mihomo_ctl/main.go" ]] && [[ ! -x "$BASE_DIR/bin/mihomo_ctl" || "$BASE_DIR/cmd/mihomo_ctl/main.go" -nt "$BASE_DIR/bin/mihomo_ctl" ]] && command -v go >/dev/null 2>&1; then
  echo "构建 Go 菜单辅助工具..."
  if ! go build -o "$BASE_DIR/bin/mihomo_ctl" "$BASE_DIR/cmd/mihomo_ctl/main.go"; then
    echo "Go 菜单辅助工具构建失败，稍后可手动执行："
    echo "  go build -o bin/mihomo_ctl cmd/mihomo_ctl/main.go"
  fi
fi

if command -v xattr >/dev/null 2>&1; then
  xattr -dr com.apple.quarantine "$BASE_DIR" 2>/dev/null || true
fi

echo "初始化完成: $BASE_DIR"
if [[ -n "${MIHOMO_BIN:-}" && -x "$MIHOMO_BIN" ]]; then
  "$MIHOMO_BIN" -v || true
else
  echo
  echo "还没有 mihomo 内核。请选择一种方式："
  echo "1) 联网下载官方 macOS 内核："
  echo "   ./bin/fetch-mihomo.sh"
  echo
  echo "2) 手动下载后放到："
  echo "   $BASE_DIR/bin/mihomo-darwin-arm64"
  echo "   chmod +x $BASE_DIR/bin/mihomo-darwin-arm64"
fi

echo
if [[ ! -x "$BASE_DIR/bin/ssr_2_mihomo" ]]; then
  echo "还没有 Go SSR 转换器。导入订阅前请在 macOS 上执行："
  echo "  go build -o bin/ssr_2_mihomo cmd/ssr_2_mihomo/main.go"
  echo
fi

if [[ ! -x "$BASE_DIR/bin/mihomo_ctl" ]]; then
  echo "还没有 Go 菜单辅助工具。选择节点前请在 macOS 上执行："
  echo "  go build -o bin/mihomo_ctl cmd/mihomo_ctl/main.go"
  echo
fi

echo
echo "打开菜单："
echo "  ./bin/mihomo-tui"
