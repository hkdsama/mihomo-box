# mihomo-box

macOS（Apple Silicon）和 Ubuntu（amd64）的 Mihomo 命令行管理包，release 版本自带所有二进制，开箱即用。

## 部署步骤

### 1. 解压并进入目录

```bash
tar -xzf mihomo-box.tar.gz
cd mihomo-box
```

### 2. 初始化

```bash
./install.sh
```

设置文件权限，macOS 同时清除系统隔离标记。

### 3. 打开菜单

```bash
./bin/mihomo-tui
```

首次使用按以下顺序操作：

```
1  → 检查环境
2  → 导入 SSR 订阅（输入链接，自动生成配置）
4  → 安装服务（选 Agent/user 或 Daemon/system）
5  → 启动服务
8  → 打开系统代理（macOS）/ 命令行代理（Ubuntu）
10 → 选择节点
```

---

## 菜单说明

| 编号 | 功能 |
|------|------|
| 1 | 检查环境和工具状态 |
| 2 | 导入 SSR 订阅链接（首次或换订阅） |
| 3 | 更新订阅并重新生成配置 |
| 4 | 安装服务（写入启动项，见下方说明） |
| 5 | 启动服务 |
| 6 | 停止服务 |
| 7 | 重启服务 |
| 8 | 打开系统代理 / 命令行代理 |
| 9 | 关闭系统代理 / 命令行代理 |
| 10 | 选择代理节点 |
| 11 | 查看日志 |
| 12 | 编辑基础配置（端口、DNS、规则） |

### 安装服务两种模式

| | macOS | Linux |
|---|---|---|
| **选项 1** | LaunchAgent | systemd user |
| 权限 | 无需 sudo | 无需 sudo |
| 启动时机 | 登录后 | 登录后（可选开启 linger 实现开机自启） |
| **选项 2** | LaunchDaemon | systemd system |
| 权限 | 需要 sudo | 需要 sudo |
| 启动时机 | 开机即启（登录前） | 开机即启（登录前） |

个人使用推荐选项 1，服务器或需要开机自启选项 2。

---

## 默认端口

```
mixed-port:          127.0.0.1:1087
external-controller: 127.0.0.1:9090
```

---

## 卸载

```bash
./uninstall.sh
```

停止服务并关闭代理，不删除目录。彻底清理直接删除 `mihomo-box` 文件夹即可。

---

## Go 模块代理（国内加速）

```bash
go env -w GO111MODULE=on
go env -w GOPROXY=https://proxy.golang.org,https://goproxy.io,direct
```
