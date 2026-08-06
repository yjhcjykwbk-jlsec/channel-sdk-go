# Channel SDK for Go 文档

[English](../README.md) | [项目 README](../../README.zh.md)

这个索引用于从首次接入逐步进入 Channel 各项公开能力的详细说明。

## 从这里开始

| 文档 | 适用场景 |
|---|---|
| [快速开始](quickstart.md) | 第一次创建机器人，或验证应用配置 |
| [从 `oapi-sdk-go` 迁移](migration-from-oapi-sdk-go.md) | 现有代码正在导入 `oapi-sdk-go/v3/channel` |
| [问题排查](troubleshooting.md) | 连接、事件、发送、上传或流式回复行为不符合预期 |

## 核心能力

| 文档 | 内容 |
|---|---|
| [API 参考](reference.md) | `New`、全部构造选项、默认值、`Channel` 方法、公开模型和错误码 |
| [接收事件](events.md) | 消息、reaction、评论、机器人入群、卡片 action、拒绝和生命周期事件 |
| [发送消息](sending-messages.md) | 文本、Markdown、富文本、卡片、媒体、名片、贴纸、mention 和回复 |
| [流式回复](streaming.md) | 增量更新 Markdown 消息和交互卡片 |
| [媒体](media.md) | 已有媒体 key、自动上传、URL 安全控制和资源下载 |
| [策略与安全](policy-and-safety.md) | 群聊/私聊准入、批处理、去重、处理锁和过期事件 |

## 可运行示例

| 示例 | 用途 |
|---|---|
| [send_message](../../examples/send_message) | 不启动 WebSocket，直接通过 REST API 发送消息 |
| [echo_bot](../../examples/echo_bot) | 接收并回复归一化消息 |
| [stream_reply](../../examples/stream_reply) | 增量更新 Markdown 回复 |
| [card_action](../../examples/card_action) | 处理交互卡片 action |
| [custom_domain](../../examples/custom_domain) | 配置 OpenAPI、OAuth 和 WebSocket 域名 |

## 能力边界

稳定的业务入口是根包 `github.com/larksuite/channel-sdk-go`。`types` 是公开包；
`internal` 下的包不属于公开 API。

当前版本使用 WebSocket 长连接接收入站事件和回调，不提供 HTTP Webhook
适配器。底层 REST、WebSocket、事件分发以及生成的 OpenAPI 模型由
`github.com/larksuite/oapi-sdk-go/v3` 提供。

## 验证文档变更

```bash
go test ./...
go test -race ./...
go test ./examples/...
CHANNEL_E2E_DRY_RUN=1 go test -tags=e2e ./e2e -run TestChannelE2E -v
```

真实环境配置见 [E2E 指南](../../e2e/README.md)。
