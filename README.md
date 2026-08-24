# hostinfo

轻量级主机自省 HTTP 服务，通过 REST API 暴露运行环境、用户、网络、系统、内存、磁盘与进程等信息。适用于容器、虚拟机、云主机或本地开发环境的调试与排查。

## 功能

- 环境变量、运行用户、家目录、临时目录
- 主机名、内网 IP、网卡详情
- 公网 IP 探测（需出站网络）
- 操作系统、内核、CPU、内存、磁盘、进程信息
- Go 运行时统计
- 纯 JSON 响应，无额外依赖运行时

## 快速开始

### 从 Release 下载（推荐）

在 [Releases](https://github.com/vevc/hostinfo/releases) 页面下载对应平台的二进制文件，解压后运行：

```bash
# Linux / macOS
chmod +x hostinfo-linux-amd64
./hostinfo-linux-amd64

# Windows
hostinfo-windows-amd64.exe
```

### 从源码运行

```bash
git clone https://github.com/vevc/hostinfo.git
cd hostinfo
go run .
```

### 从源码构建

```bash
go build -o hostinfo .
./hostinfo
```

### 使用 Docker

镜像发布在 [GHCR](https://github.com/vevc/hostinfo/pkgs/container/hostinfo)，支持 `linux/amd64` 与 `linux/arm64`：

```bash
docker pull ghcr.io/vevc/hostinfo:latest
docker run --rm -p 8080:8080 ghcr.io/vevc/hostinfo:latest

# 指定版本
docker run --rm -p 8080:8080 ghcr.io/vevc/hostinfo:v0.1.0
```

本地构建：

```bash
docker build -t hostinfo .
docker run --rm -p 8080:8080 hostinfo
```

服务默认监听 `:8080`。

## 配置

| 方式 | 示例 | 说明 |
|------|------|------|
| 命令行参数 | `-addr :9090` | 监听地址 |
| 环境变量 | `ADDR=:9090` | 同上，优先级低于命令行参数 |

```bash
./hostinfo -addr 127.0.0.1:8080
ADDR=:3000 ./hostinfo
```

## API 接口

访问根路径可查看完整接口列表：

```bash
curl http://127.0.0.1:8080/
```

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/health` | 健康检查 |
| GET | `/api/info` | 汇总快照（用户、系统、网络、内存、进程） |
| GET | `/api/env` | 全部环境变量 |
| GET | `/api/env/{key}` | 单个环境变量 |
| GET | `/api/user` | 运行用户、UID/GID、家目录、临时/缓存/配置目录 |
| GET | `/api/network` | 主机名、内网 IP、网卡信息 |
| GET | `/api/public-ip` | 公网 IP（出站探测，需互联网） |
| GET | `/api/system` | OS、内核、CPU、时区、开机时间 |
| GET | `/api/host` | 详细宿主机信息 |
| GET | `/api/cpu` | CPU 详情 |
| GET | `/api/memory` | 系统内存 + Go 运行时内存 |
| GET | `/api/disk` | 磁盘分区与用量 |
| GET | `/api/process` | 当前进程 PID/PPID、可执行文件、工作目录 |
| GET | `/api/workdir` | 工作目录 |
| GET | `/api/hostname` | 主机名 |
| GET | `/api/runtime` | Go 运行时信息 |

### 示例

```bash
curl http://127.0.0.1:8080/api/info
curl http://127.0.0.1:8080/api/user
curl http://127.0.0.1:8080/api/network
curl http://127.0.0.1:8080/api/public-ip
curl http://127.0.0.1:8080/api/env/PATH
```

`/api/public-ip` 成功时返回：

```json
{
  "ip": "203.0.113.10",
  "source": "ipify",
  "latency_ms": 320,
  "reachable": true,
  "attempted_at": "2026-08-24T03:00:00Z"
}
```

无法访问外网时返回 `503`，`reachable` 为 `false`。

## 发布 Release

在 GitHub 仓库 **Actions → Release → Run workflow** 中手动触发发版：

1. 选择要发布的分支（通常为 `main`）
2. 填写 **tag**，例如 `v0.1.0`
3. 如需预发布版本，勾选 **prerelease**
4. 点击 **Run workflow**

Workflow 会基于当前选中的 commit：

1. 交叉编译二进制并创建 GitHub Release（含 `checksums.txt`）
2. 构建并推送多架构 Docker 镜像到 GHCR

**二进制文件**

| 平台 | 文件名 |
|------|--------|
| Linux amd64 | `hostinfo-linux-amd64` |
| Linux arm64 | `hostinfo-linux-arm64` |
| macOS amd64 | `hostinfo-darwin-amd64` |
| macOS arm64 | `hostinfo-darwin-arm64` |
| Windows amd64 | `hostinfo-windows-amd64.exe` |

**Docker 镜像**（`ghcr.io/vevc/hostinfo`）

| 标签 | 说明 |
|------|------|
| `v0.1.0` 等 | 与发版 tag 一致 |
| `latest` | 仅正式版（非 prerelease）会更新 |

平台：`linux/amd64`、`linux/arm64`。

若镜像默认为私有，可在 GitHub **Packages** 中将该 package 设为 Public。

## 项目结构

```
.
├── main.go                  # 入口与路由
├── Dockerfile               # 多阶段构建镜像
├── internal/
│   ├── api/                 # HTTP handlers
│   └── sysinfo/             # 信息采集逻辑
└── .github/workflows/
    ├── ci.yml               # 构建与测试
    └── release.yml          # Release 与 GHCR 镜像发布
```

## 依赖

- [gopsutil](https://github.com/shirou/gopsutil) — 系统信息采集（CPU、内存、磁盘、主机信息）

## 开发

```bash
go test ./...
go build -o hostinfo .
```

## 注意事项

- `/api/env` 会暴露全部环境变量，**请勿将服务直接暴露到公网**。
- `/api/public-ip` 会向第三方服务（ipify、AWS checkip、ifconfig.me）发起出站 HTTP 请求。
- 部分接口在 Windows 与 Unix 上返回字段可能略有差异（如 UID/GID、用户组）。
