# mihomo-box

一个面向 macOS 和 Ubuntu 命令行环境的 Mihomo 管理包。

默认策略：

1. SSR 订阅解析和 `config/config.yaml` 生成使用 Go。
2. `bin/ssr_2_mihomo` 是主力转换器，由 `cmd/ssr_2_mihomo/main.go` 构建。
3. `bin/mihomo_ctl` 是菜单 JSON/API 辅助工具，由 `cmd/mihomo_ctl/main.go` 构建。
4. `bin/backup/ssr_to_mihomo.py` 只作为旧版备份，不在菜单流程中调用。
5. 当前菜单仍是 Bash 脚本，按平台编排 macOS 或 Ubuntu CLI 命令。

## 目录结构

```text
mihomo-box/
├── bin/
│   ├── mihomo-tui                  # Bash 菜单入口
│   ├── fetch-mihomo.sh              # 下载 mihomo 内核，默认 mac，支持 linux 参数
│   ├── mihomo-darwin-arm64          # 下载/手动放入，未纳入源码
│   ├── mihomo-linux-amd64-v3        # 下载/手动放入，未纳入源码
│   ├── ssr_2_mihomo                 # Go 构建产物，未纳入源码
│   ├── mihomo_ctl                   # Go 构建产物，未纳入源码
│   └── backup/
│       └── ssr_to_mihomo.py         # Python 备份转换器
├── cmd/
│   ├── ssr_2_mihomo/
│   │   └── main.go                  # Go SSR -> mihomo 配置转换器
│   └── mihomo_ctl/
│       └── main.go                  # Go 菜单 JSON/API 辅助工具
├── config/
│   └── config.yaml.example
├── data/                            # 运行时订阅地址
├── logs/                            # 运行时日志
├── service/                         # 预留服务文件目录
├── archives/                        # 下载包或临时归档
├── install.sh
├── uninstall.sh
└── README.md
```

## macOS 第一次使用

```bash
tar -xzf mihomo-box.tar.gz
cd mihomo-box

./bin/fetch-mihomo.sh
./install.sh
./bin/mihomo-tui
```

`fetch-mihomo.sh` 用参数区分目标平台，默认是 mac：

```bash
# 默认 macOS Apple Silicon:
./bin/fetch-mihomo.sh

# Linux amd64-v3:
./bin/fetch-mihomo.sh linux
```

默认下载文件：

```text
mac:   mihomo-darwin-arm64-v1.19.25.gz
linux: mihomo-linux-amd64-v3-v1.19.25.gz
```

如果本机安装了 Go，`./install.sh` 会自动构建：

```bash
go build -o bin/ssr_2_mihomo cmd/ssr_2_mihomo/main.go
go build -o bin/mihomo_ctl cmd/mihomo_ctl/main.go
```

如果 macOS 已经提前下载了 mihomo 内核，可以直接放到：

```text
bin/mihomo-darwin-arm64
```

然后执行：

```bash
chmod +x bin/mihomo-darwin-arm64
./install.sh
./bin/mihomo-tui
```

## Ubuntu 命令行使用

适合无图形界面的 Ubuntu，只配置当前用户，不影响其他用户。

```bash
./bin/fetch-mihomo.sh linux
./install.sh
./bin/mihomo-tui
```

Ubuntu 下菜单里的服务管理使用 `systemd --user`：

```text
systemctl --user start/stop/restart mihomo-box.service
```

“打开命令行代理”不会修改桌面设置，也不会影响其他用户。它会写入：

```text
~/.config/mihomo-box/proxy.env
```

并在当前用户的 `~/.bashrc` 和 `~/.profile` 里加入 source 入口。新打开的 terminal 会自动获得：

```text
http_proxy=http://127.0.0.1:1087
https_proxy=http://127.0.0.1:1087
all_proxy=socks5h://127.0.0.1:1087
```

已经打开的 terminal 需要手动执行一次：

```bash
source ~/.config/mihomo-box/proxy.env
```

## 推荐菜单顺序

```text
1. 检查环境
2. 导入 SSR 订阅链接，并生成配置（Go）
4. 启动 Mihomo 服务
7. 打开 macOS 系统代理 / Ubuntu 命令行代理
9. 选择代理站点 / 节点
10. 刷新状态
```

## 默认端口

```text
mixed-port: 127.0.0.1:1087
external-controller: 127.0.0.1:9090
proxy group: PROXY
```

## 菜单实现建议

现在的菜单是 `bin/mihomo-tui` Bash 脚本。macOS 使用 `launchctl` 和 `networksetup`；Ubuntu 命令行使用 `systemd --user` 和 shell 代理环境文件。

菜单里的 SSR 解析和 JSON 处理已经走 Go，Python 只保留在 `bin/backup/` 里作为旧版备份。

长期如果想做单文件发布，再把菜单整体迁到 Go：统一参数、状态展示、API JSON 处理和 launchd plist 生成，最后只保留一个 `mihomo-box` 可执行文件。

## 卸载

```bash
./uninstall.sh
```

它会停止服务、关闭当前网络服务或命令行代理，但不会删除整个目录。直接删除 `mihomo-box` 文件夹即可彻底清理。
