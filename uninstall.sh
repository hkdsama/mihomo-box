#!/usr/bin/env bash
set -euo pipefail
PLIST="$HOME/Library/LaunchAgents/com.local.mihomo.box.plist"

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
NET="$(network_service)"

launchctl bootout "gui/$(id -u)" "$PLIST" >/dev/null 2>&1 || true
rm -f "$PLIST"
networksetup -setwebproxystate "$NET" off 2>/dev/null || true
networksetup -setsecurewebproxystate "$NET" off 2>/dev/null || true
networksetup -setsocksfirewallproxystate "$NET" off 2>/dev/null || true

echo "已停止服务并关闭系统代理。"
echo "如需彻底删除，直接删除当前 mihomo-box-macos 文件夹。"
