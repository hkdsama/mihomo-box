# mihomo-box

macOS（Apple Silicon）和 Ubuntu（amd64）的 Mihomo 命令行管理包。

## 部署步骤

### 1. 解压并进入目录

```bash
tar -xzf mihomo-box.tar.gz
cd mihomo-box
```

### 2. 下载 mihomo 内核

```bash
# macOS Apple Silicon
./bin/fetch-mihomo.sh

# Ubuntu amd64
./bin/fetch-mihomo.sh linux
```

### 3. 初始化

```bash
./install.sh
```

如果本机安装了 Go，会自动构建 `bin/ssr_2_mihomo` 和 `bin/mihomo_ctl`。

### 4. 创建基础配置

首次使用需要创建 `config/config_base.yaml`（如果不存在）：

```bash
cp config/config_base.yaml.example config/config_base.yaml  # 若有示例
# 或直接编辑已有的 config/config_base.yaml，按需修改端口、DNS 等
```

`config_base.yaml` 是手动维护的配置，包含端口、DNS、路由规则等，**不会被订阅更新覆盖**。

### 5. 导入 SSR 订阅，生成配置

```bash
./bin/ssr_2_mihomo --url "你的SSR订阅链接"
```

执行后自动完成两件事：
1. 解析订阅，将节点写入 `config/proxies.yaml`
2. 合并 `config_base.yaml` + `proxies.yaml` → `config/config.yaml`

如果只想重新合并（不拉订阅）：

```bash
./bin/ssr_2_mihomo merge
```

### 6. 启动菜单

```bash
./bin/mihomo-tui
```

推荐操作顺序：

```
1. 检查环境
2. 导入 SSR 订阅（生成配置）
3. 启动 Mihomo 服务
4. 打开系统代理（macOS）/ 命令行代理（Ubuntu）
5. 选择节点
```

---

## 配置文件说明

| 文件 | 说明 |
|------|------|
| `config/config_base.yaml` | 手动维护，含端口/DNS/规则，纳入版本控制 |
| `config/proxies.yaml` | 自动生成，含节点列表，**勿手动修改** |
| `config/config.yaml` | 合并结果，mihomo 实际读取，不纳入版本控制 |

修改端口或规则时只需编辑 `config_base.yaml`，然后执行 `./bin/ssr_2_mihomo merge` 重新合并。

---

## 默认端口

```
mixed-port:          127.0.0.1:1087
external-controller: 127.0.0.1:9090
```

---

## Ubuntu 命令行代理

启用后会在新 terminal 中自动生效：

```
http_proxy=http://127.0.0.1:1087
https_proxy=http://127.0.0.1:1087
all_proxy=socks5h://127.0.0.1:1087
```

已打开的 terminal 需手动执行一次：

```bash
source ~/.config/mihomo-box/proxy.env
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
