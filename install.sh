#!/usr/bin/env bash
set -euo pipefail
BASE_DIR="$(cd "$(dirname "$0")" && pwd)"

mkdir -p "$BASE_DIR"/{config,data,logs}

chmod +x "$BASE_DIR/bin/mihomo-tui" \
         "$BASE_DIR/bin/fetch-mihomo.sh" \
         "$BASE_DIR/bin/ssr_2_mihomo" \
         "$BASE_DIR/bin/mihomo_ctl" 2>/dev/null || true

OS_NAME="$(uname -s)"
case "$OS_NAME" in
  Darwin)
    chmod +x "$BASE_DIR/bin/mihomo-darwin-arm64" 2>/dev/null || true
    xattr -dr com.apple.quarantine "$BASE_DIR" 2>/dev/null || true
    ;;
  Linux)
    chmod +x "$BASE_DIR/bin/mihomo-linux-amd64-v3" 2>/dev/null || true
    ;;
esac

echo "初始化完成: $BASE_DIR"
echo
echo "打开菜单："
echo "  ./bin/mihomo-tui"
