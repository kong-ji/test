# test

Go 后端服务骨架。使用标准库 `net/http`，零外部依赖，开箱即跑。

## 快速开始

```bash
# 直接运行
go run ./cmd/server

# 编译成可执行文件
go build -o bin/server ./cmd/server

# 运行测试
go test ./...
```

服务默认监听 `:8080`。

## 接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/` | 服务信息 |
| GET | `/healthz` | 健康检查，返回状态、运行时长、Go 版本 |

```bash
curl http://localhost:8080/healthz
# {"go":"go1.24.0","goroutines":6,"status":"ok","uptime":"2.07s"}
```

未知路径返回 `404`，路径存在但方法不对返回 `405`（带 `Allow` 头）。

## 配置

全部通过环境变量注入，均有默认值：

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `APP_ENV` | `development` | 运行环境标识 |
| `APP_PORT` | `8080` | 监听端口 |
| `APP_LOG_LEVEL` | `info` | 日志级别：`debug`/`info`/`warn`/`error` |
| `APP_READ_TIMEOUT` | `5s` | 读超时 |
| `APP_WRITE_TIMEOUT` | `10s` | 写超时 |
| `APP_IDLE_TIMEOUT` | `60s` | 空闲连接超时 |
| `APP_SHUTDOWN_TIMEOUT` | `10s` | 优雅关闭最长等待时间 |

时间字段支持 `10s`、`1m30s` 等形式，纯数字按秒处理。

## 目录结构

```
cmd/server/        程序入口
internal/
  config/          环境变量配置加载
  handler/         HTTP 处理器
  response/        统一 JSON 响应
  router/          路由注册
  server/          HTTP 服务封装与优雅关闭
```

`internal/` 下的包无法被外部模块导入，方便后续重构。

## 行为说明

- **结构化日志**：使用标准库 `log/slog`，输出 JSON 到标准输出
- **优雅关闭**：收到 `SIGINT`/`SIGTERM` 后停止接收新请求，等待在途请求完成，最长等待 `APP_SHUTDOWN_TIMEOUT`

## 环境要求

Go 1.24 及以上。
