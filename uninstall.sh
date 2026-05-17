#!/usr/bin/env bash
set -euo pipefail

OS_NAME="$(uname -s)"
PLIST="$HOME/Library/LaunchAgents/com.local.mihomo.box.plist"
SYSTEMD_SERVICE="$HOME/.config/systemd/user/mihomo-box.service"
PROXY_ENV_FILE="$HOME/.config/mihomo-box/proxy.env"

network_service() {
  local iface service
  iface="$(route get default 2>/dev/null | awk '/interface:/{print $2; exit}')"
  if [[ -n "${iface:-}" ]]; then
    service="$(networksetup -listallhardwareports | awk -v dev="$iface" '
      BEGIN { RS=""; FS="\n" }
      $0 ~ "Device: " dev { for (i=1;i<=NF;i++) if ($i ~ /^Hardware Port:/) { sub(/^Hardware Port: /,"",$i); print $i; exit } }
    ')"
  fi
  [[ -n "${service:-}" ]] && echo "$service" || echo "Wi-Fi"
}

if [[ "$OS_NAME" == "Darwin" ]]; then
  NET="$(network_service)"
  launchctl bootout "gui/$(id -u)" "$PLIST" >/dev/null 2>&1 || true
  rm -f "$PLIST"
  networksetup -setwebproxystate "$NET" off 2>/dev/null || true
  networksetup -setsecurewebproxystate "$NET" off 2>/dev/null || true
  networksetup -setsocksfirewallproxystate "$NET" off 2>/dev/null || true
  echo "已停止 launchd 服务并关闭 macOS 系统代理。"
elif [[ "$OS_NAME" == "Linux" ]]; then
  systemctl --user disable --now mihomo-box.service >/dev/null 2>&1 || true
  rm -f "$SYSTEMD_SERVICE"
  systemctl --user daemon-reload >/dev/null 2>&1 || true
  if [[ -f "$PROXY_ENV_FILE" ]]; then
    : > "$PROXY_ENV_FILE"
  fi
  echo "已停止 systemd user 服务并关闭当前用户命令行代理。"
else
  echo "当前系统暂不支持自动卸载: $OS_NAME"
fi

echo "如需彻底删除，直接删除当前 mihomo-box 文件夹。"
