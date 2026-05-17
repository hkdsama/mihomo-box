这里应放置 mihomo 内核：

Apple Silicon:
  bin/mihomo-darwin-arm64

Linux amd64-v3:
  bin/mihomo-linux-amd64-v3

联网但不用 Homebrew 的下载方式：
  ./bin/fetch-mihomo.sh
  ./bin/fetch-mihomo.sh linux

说明：
  bin/ 目录尽量只放可执行入口和运行时二进制。
  下载包、.gz 归档可以放在 archives/。
