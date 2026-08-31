# Banner 指纹识别系统 — 需求说明

## 任务目标

使用 Golang 语言开发一个 Banner 指纹识别系统，能力为接收一批网络扫描原始数据（ip、port、banner），识别出对应的协议、软件与版本信息，并以 client + server 架构、Docker Compose 一键启动的方式交付。系统输出的识别深度至少要达到示例所示。

## 交付要求

| 项 | 要求 |
| --- | --- |
| 语言 | Golang |
| 架构 | client + server |
| 启动方式 | Docker Compose 一键启动 |
| 识别深度 | 至少覆盖下方示例中的协议与软件 |

## 输入：网络扫描原始数据

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `ip` | string | 目标 IP 地址 |
| `port` | int | 目标端口 |
| `banner` | string | 扫描抓取的原始 banner 文本 |

## 输出：识别结果

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `ip` | string | 目标 IP，原样透传 |
| `port` | int | 目标端口，原样透传 |
| `protocol` | string | 协议类型，如 `SSH` / `HTTP` / `MySQL` / `Redis` / `FTP` |
| `product` | string | 软件名称，如 `OpenSSH` / `nginx` / `Apache` |
| `version` | string | 版本号，无法识别时为空字符串 `""` |
| `os_hint` | string | 操作系统线索，无法推断时为空字符串 `""` |
| `confidence` | float | 识别置信度，取值 0 ~ 1 |

约定：识别不出来的字段留空字符串，**不猜测、不硬编码假数据**；置信度随匹配依据的强度浮动（精确版本号 > 仅产品名 > 仅端口推断）。

## 示例响应

> 自测数据，非最终验收数据。

```json
[
  {"ip":"1.2.3.4","port":22,"protocol":"SSH","product":"OpenSSH","version":"8.9p1","os_hint":"Ubuntu","confidence":0.95},
  {"ip":"1.2.3.5","port":80,"protocol":"HTTP","product":"nginx","version":"1.24.0","os_hint":"","confidence":0.9},
  {"ip":"1.2.3.6","port":443,"protocol":"HTTP","product":"Apache","version":"2.4.57","os_hint":"","confidence":0.9},
  {"ip":"1.2.3.7","port":3306,"protocol":"MySQL","product":"MySQL","version":"8.0.32","os_hint":"","confidence":0.9},
  {"ip":"1.2.3.8","port":6379,"protocol":"Redis","product":"Redis","version":"","os_hint":"","confidence":0.7},
  {"ip":"1.2.3.9","port":21,"protocol":"FTP","product":"ProFTPD","version":"1.3.7","os_hint":"","confidence":0.9},
  {"ip":"1.2.3.10","port":8080,"protocol":"HTTP","product":"Jetty","version":"9.4.51","os_hint":"","confidence":0.85}
]
```

## 识别深度要求

示例涉及的识别目标即为最低基准：

| 端口 | 协议 | 产品 | 版本 | 备注 |
| --- | --- | --- | --- | --- |
| 22 | SSH | OpenSSH | `8.9p1` | 需额外给出 `os_hint`（Ubuntu） |
| 80 | HTTP | nginx | `1.24.0` | |
| 443 | HTTP | Apache | `2.4.57` | TLS 端口，banner 通常来自 HTTP 响应头 |
| 3306 | MySQL | MySQL | `8.0.32` | 握手包特征 |
| 6379 | Redis | Redis | 空 | banner 无版本时留空，置信度降至 0.7 |
| 21 | FTP | ProFTPD | `1.3.7` | 欢迎横幅识别 |
| 8080 | HTTP | Jetty | `9.4.51` | 非标准端口，不能依赖 80/443 推断 |

## 待确认事项

- **输入投递方式**：client 以什么形式提交扫描数据（JSON 文件 / stdin / HTTP 批量接口）
- **client 形态**：纯 CLI 命令行工具，还是带界面的客户端
- **server 接口**：单次批量提交，还是支持流式/分片上传超大批次
- **规则可扩展性**：识别规则是否需要外置配置文件、支持热加载
- **性能要求**：单批次规模量级、是否需要并发识别
- **验收数据**：最终验收用的输入样本是否会提供
