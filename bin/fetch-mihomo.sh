#!/usr/bin/env bash
set -euo pipefail

BASE_DIR="$(cd "$(dirname "$0")/.." && pwd)"
VERSION="${MIHOMO_VERSION:-v1.19.25}"
TARGET="${1:-mac}"

case "$TARGET" in
  mac|macos|darwin)
    TARGET="mac"
    ASSET="mihomo-darwin-arm64-${VERSION}.gz"
    OUT="mihomo-darwin-arm64"
    ;;
  linux)
    ASSET="mihomo-linux-amd64-v3-${VERSION}.gz"
    OUT="mihomo-linux-amd64-v3"
    ;;
  -h|--help|help)
    echo "Usage: $0 [mac|linux]"
    echo
    echo "默认 mac:   mihomo-darwin-arm64-${VERSION}.gz"
    echo "linux:      mihomo-linux-amd64-v3-${VERSION}.gz"
    exit 0
    ;;
  *)
    echo "不支持的目标平台: $TARGET"
    echo "Usage: $0 [mac|linux]"
    exit 1
    ;;
esac

URL="https://github.com/MetaCubeX/mihomo/releases/download/${VERSION}/${ASSET}"
ARCHIVE="$BASE_DIR/archives/${ASSET}"
DST="$BASE_DIR/bin/${OUT}"

mkdir -p "$BASE_DIR/archives" "$BASE_DIR/bin"

echo "下载: $URL"
curl --noproxy '*' -L --fail --retry 3 --connect-timeout 15 -o "$ARCHIVE" "$URL"

echo "解压: $ARCHIVE -> $DST"
gunzip -c "$ARCHIVE" > "$DST"
chmod +x "$DST"

case "$(uname -s)" in
  Darwin) HOST="mac" ;;
  Linux) HOST="linux" ;;
  *) HOST="unknown" ;;
esac

if [[ "$TARGET" == "$HOST" ]]; then
  if [[ "$HOST" == "mac" ]] && command -v xattr >/dev/null 2>&1; then
    xattr -dr com.apple.quarantine "$DST" 2>/dev/null || true
  fi

  echo "完成: $DST"
  "$DST" -v || true
else
  echo "完成: $DST"
  echo "当前系统不是 $TARGET，仅保留目标平台二进制。"
fi
