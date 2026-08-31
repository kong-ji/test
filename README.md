# Banner 指纹识别系统

基于 Golang 的 Banner 指纹识别系统：接收一批网络扫描原始数据（`ip`、`port`、`banner`），识别出对应的协议、软件与版本信息。采用 client + server 架构，Docker Compose 一键启动，识别规则外置、可扩展。

## 架构

```
cmd/client          CLI 客户端（提交扫描数据、打印识别结果）
cmd/server          识别服务（HTTP API）
internal/
  config            环境变量配置加载
  fingerprint       识别引擎（规则匹配 → TLS 启发式 → 端口回退 → fallback）
  rules             规则文件加载与正则预编译
  handler           HTTP 处理器（health / identify）
  response          统一 JSON 响应
  router            路由注册
  server            HTTP 服务封装与优雅关闭
configs/rules.json  外置识别规则（可挂载、可替换）
```

## 快速开始

### Docker Compose 一键启动

```bash
docker compose up --build
```

服务默认监听 `8080`。规则文件通过只读 volume 挂载到 `/etc/banner-fp/rules.json`，替换该文件后重启容器即可更新规则，无需重建镜像。

### 本地运行

```bash
# 启动 server
go run ./cmd/server

# 另开终端，用 client 提交数据（-file 指定文件，或从 stdin 读取）
go run ./cmd/client -file examples/input.json
echo '[{"ip":"1.2.3.4","port":22,"banner":"SSH-2.0-OpenSSH_8.9p1 Ubuntu-3"}]' | go run ./cmd/client
```

### 运行测试

```bash
go test ./...
```

## 接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/` | 服务信息 |
| GET | `/health` | 健康检查 |
| POST | `/fingerprint` | 批量识别（请求/响应均为 JSON 数组） |

### 识别请求

```json
[{"ip": "1.2.3.4", "port": 22, "banner": "SSH-2.0-OpenSSH_8.9p1 Ubuntu-3"}]
```

### 识别响应

```json
[{"ip":"1.2.3.4","port":22,"protocol":"SSH","product":"OpenSSH","version":"8.9p1","os_hint":"Ubuntu","confidence":0.95}]
```

字段说明：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `ip` / `port` | string / int | 原样透传 |
| `protocol` | string | 协议类型，如 `SSH` / `HTTP` / `MySQL` / `Redis` / `FTP` / `TLS`；识别不出时返回 `unknown` |
| `product` | string | 软件名称，如 `OpenSSH` / `nginx` |
| `version` | string | 版本号，无法识别时为空字符串 |
| `os_hint` | string | 操作系统线索，无法推断时为空字符串 |
| `confidence` | float | 识别置信度，取值 0 ~ 1 |

## CLI 客户端

```bash
go run ./cmd/client [flags]
```

| 参数 | 默认值 | 说明 |
| --- | --- | --- |
| `-server` | `http://localhost:8080` | 服务地址 |
| `-file` | 空 | 输入 JSON 文件路径，为空则从 stdin 读取 |
| `-timeout` | `10s` | HTTP 请求超时 |

## 识别规则

规则完全外置，代码中不硬编码任何指纹。编辑 `configs/rules.json` 即可扩展覆盖范围：

- `rules`：正则规则列表，按 `confidence` 取最高分命中，支持命名捕获组 `version` / `product` 提取，可用 `ports` 限定端口
- `os_hints`：操作系统线索，命中后写入 `os_hint`
- `port_fallbacks`：规则全部未命中时按端口回退（弱信号）
- `fallback`：最终兜底，全部字段留空

识别优先级：规则匹配 > TLS 记录层启发式 > 端口回退 > fallback。识别不出的记录统一返回 `protocol: "unknown"`（其余字段留空），不猜测、不硬编码假数据，也绝不因识别失败而报错或崩溃；置信度随匹配依据强度浮动（精确版本号 > 仅产品名 > 仅端口推断）。

## 配置

全部通过环境变量注入：

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `APP_ENV` | `development` | 运行环境标识 |
| `APP_PORT` | `8080` | 监听端口 |
| `APP_LOG_LEVEL` | `info` | 日志级别：`debug`/`info`/`warn`/`error` |
| `APP_RULES_PATH` | `configs/rules.json` | 规则文件路径 |
| `APP_READ_TIMEOUT` | `5s` | 读超时 |
| `APP_WRITE_TIMEOUT` | `10s` | 写超时 |
| `APP_IDLE_TIMEOUT` | `60s` | 空闲连接超时 |
| `APP_SHUTDOWN_TIMEOUT` | `10s` | 优雅关闭最长等待时间 |

## 行为说明

- **结构化日志**：标准库 `log/slog`，JSON 输出到标准输出
- **优雅关闭**：收到 `SIGINT`/`SIGTERM` 后停止接收新请求，等待在途请求完成
- **请求限流**：识别接口单次请求体上限 1 MiB
- **热加载**：`SIGHUP` 信号当前为占位（打印日志），真正规则热加载待接入

## 环境要求

Go 1.24 及以上；Docker（使用 Docker Compose 时）。
