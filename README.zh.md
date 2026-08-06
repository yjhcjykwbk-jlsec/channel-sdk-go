# Lark Channel SDK for Go

[English](README.md)

`channel-sdk-go` 是用于构建飞书和 Lark 会话式机器人的 Go 包。它以一个高阶
Channel 作为入口，覆盖 WebSocket 事件监听、入站事件归一化、消息发送、媒体
处理、卡片回调、策略控制和流式回复。

要求 Go 1.18 或更高版本。

## 安装

```bash
go get github.com/larksuite/channel-sdk-go
```

## 最小示例

```go
package main

import (
	"context"
	"log"
	"os"

	channel "github.com/larksuite/channel-sdk-go"
)

func main() {
	ch, err := channel.New(os.Getenv("APP_ID"), os.Getenv("APP_SECRET"))
	if err != nil {
		log.Fatalf("create channel: %v", err)
	}

	ch.OnMessage(func(ctx context.Context, msg *channel.NormalizedMessage) error {
		_, err := ch.Send(ctx, &channel.SendInput{
			ReceiveID: msg.ChatID,
			Text:      "received: " + msg.Content,
		})
		return err
	})

	if err := ch.Start(context.Background()); err != nil {
		log.Fatalf("start channel: %v", err)
	}
}
```

`Start(ctx)` 会建立 WebSocket 长连接，并阻塞直到连接结束。`Send` 和 `Stream`
使用 REST API，不调用 `Start` 也可以使用。

## 文档

| 主题 | 内容 |
|---|---|
| [文档索引](docs/zh-CN/README.md) | 全部用户指南与示例 |
| [快速开始](docs/zh-CN/quickstart.md) | 准备应用、配置事件并运行 echo bot |
| [API 参考](docs/zh-CN/reference.md) | 构造参数、默认值、方法、模型和错误码 |
| [接收事件](docs/zh-CN/events.md) | 消息归一化、回调、顺序与并发语义 |
| [发送消息](docs/zh-CN/sending-messages.md) | `SendInput` 参数、目标识别、消息类型与回复 |
| [流式回复](docs/zh-CN/streaming.md) | Markdown 与交互卡片流控制器 |
| [媒体](docs/zh-CN/media.md) | 使用 key、文件、字节或 URL 上传及下载资源 |
| [策略与安全](docs/zh-CN/policy-and-safety.md) | 准入策略、批处理、去重和过期事件 |
| [迁移指南](docs/zh-CN/migration-from-oapi-sdk-go.md) | 从 `oapi-sdk-go/v3/channel` 迁移 |
| [问题排查](docs/zh-CN/troubleshooting.md) | 事件、权限、发送、媒体与流式回复排障 |
| [E2E 测试](e2e/README.md) | 运行真实飞书/Lark 端到端测试 |

## 示例

- [发送消息](examples/send_message)
- [Echo bot](examples/echo_bot)
- [流式回复](examples/stream_reply)
- [卡片回调](examples/card_action)
- [自定义域名](examples/custom_domain)

## 包边界

业务代码通常只需要导入根包：

```go
import channel "github.com/larksuite/channel-sdk-go"
```

需要显式类型或默认配置时，可以导入公开的 [`types`](types) 包。`internal`
下的包属于实现细节，不在兼容性承诺范围内。

当前版本通过 WebSocket 长连接接收事件和回调，不提供 HTTP Webhook 适配器。
如果集成需要 Channel 之外的 OpenAPI 能力，可以使用 `RawClient()` 或主
`oapi-sdk-go/v3` 模块。

## 本地开发

```bash
go test ./...
go test -race ./...
go test ./examples/...
```

真实 E2E 测试使用 `e2e` build tag，不包含在普通单元测试中：

```bash
CHANNEL_E2E_DRY_RUN=1 go test -tags=e2e ./e2e -run TestChannelE2E -v
```

## 贡献

提交 Pull Request 前，请先阅读 [CONTRIBUTING.md](CONTRIBUTING.md)。所有贡献者
均须遵守项目的 [Code of Conduct](CODE_OF_CONDUCT.md)。

## 安全

发现潜在安全漏洞时，请按照 [SECURITY.md](SECURITY.md) 中的方式私下报告，
不要通过公开 GitHub Issue 披露安全漏洞。

## 许可证

本项目采用 [MIT License](LICENSE)。
