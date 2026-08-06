# Channel 快速开始

[English](../quickstart.md) | [文档索引](README.md)

本指南用于准备一个飞书或 Lark 自建机器人，启动 WebSocket 长连接，并回复收到
的消息。

## 1. 准备应用

在飞书或 Lark 开发者后台中：

1. 创建自建应用并启用机器人能力。
2. 开通机器人需要的消息权限。常见权限包括 `im:message` 和
   `im:message:send_as_bot`；使用 reaction、群成员、文件或文档评论能力时，
   还需要开通相应权限。
3. 在事件与回调配置中选择长连接/WebSocket 订阅方式。
4. 订阅接收消息事件 `im.message.receive_v1`。
5. 发布应用版本并安装到测试租户。修改权限或订阅后，需要重新发布或安装。
6. 将机器人加入测试群，或者打开与机器人的单聊。

群聊默认要求消息 mention 机器人。修改这一行为前请先阅读
[策略与安全](policy-and-safety.md)。

## 2. 安装模块

```bash
go get github.com/larksuite/channel-sdk-go
```

在进程环境中设置凭证：

```bash
export APP_ID=cli_xxx
export APP_SECRET=your_app_secret
```

示例程序直接读取环境变量，不会自动加载 `.env`。只有仓库中的 E2E runner
会读取根目录 `.env`。

Lark 租户需要设置 Lark OpenAPI 域名：

```go
ch, err := channel.New(
	appID,
	appSecret,
	channel.WithDomain("https://open.larksuite.com"),
)
```

SDK 可以根据这个标准 OpenAPI 域名推导 Lark OAuth 域名。使用自定义或私有
OpenAPI 域名时，通常还需要单独传入 `WithOAuthBaseURL(...)`。

## 3. 运行 Echo Bot

可运行代码位于 [`examples/echo_bot`](../../examples/echo_bot)。

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

	ch.OnError(func(err error) {
		log.Printf("channel error: %v", err)
	})

	ch.OnMessage(func(ctx context.Context, msg *channel.NormalizedMessage) error {
		_, err := ch.Send(ctx, &channel.SendInput{
			ReceiveID: msg.ChatID,
			Text:      "echo: " + msg.Content,
		})
		return err
	})

	if err := ch.Start(context.Background()); err != nil {
		log.Fatalf("start channel: %v", err)
	}
}
```

在仓库中运行：

```bash
go run ./examples/echo_bot
```

`Start(ctx)` 会注册已配置的处理器、建立 WebSocket 连接并阻塞。事件处理器和
生命周期处理器应在调用 `Start` 前注册。

## 4. 不启动 WebSocket 直接发送

REST 能力不依赖 `Start`：

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

result, err := ch.Send(ctx, &channel.SendInput{
	ReceiveID: "oc_xxx",
	Text:      "hello from channel-sdk-go",
})
if err != nil {
	return err
}
log.Printf("sent message %s", result.MessageID)
```

应用不再使用 Channel 时调用 `Stop(ctx)`。`Stop` 可以重复调用，但已经停止的
Channel 不应重新启动；新的生命周期应重新构造 Channel。

## 5. 下一步

- [接收和归一化事件](events.md)
- [发送全部支持的消息类型](sending-messages.md)
- [构建流式回复](streaming.md)
- [查看全部构造选项和默认值](reference.md)
- [排查常见配置和权限问题](troubleshooting.md)
