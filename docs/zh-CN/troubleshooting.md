# 问题排查

[English](../troubleshooting.md) | [文档索引](README.md)

排查前先准备带边界的 context、`OnError` 处理器和安全日志。日志可以记录 event
或 message ID，但禁止记录凭证、access token、完整文档内容或包含敏感数据的
原始 payload。

```go
ch.OnError(func(err error) {
	var channelErr *channel.FeishuChannelError
	if errors.As(err, &channelErr) {
		log.Printf("channel error code=%s", channelErr.Code)
		return
	}
	log.Printf("channel error: %v", err)
})
```

## `New` 失败

| 现象 | 检查项 |
|---|---|
| `appID is required` | 进程环境中是否设置 `APP_ID` |
| `appSecret is required...` | 是否设置 `APP_SECRET`，或配置了非 nil client assertion provider |
| 看起来已配置但读取为空 | 应用不会自动加载 `.env` |

不要通过打印 secret 来验证。只检查环境变量是否存在以及长度是否非零。

## WebSocket 一直没有 Ready

检查：

1. 应用是否为支持的自建机器人应用。
2. 事件与回调中是否开启长连接/WebSocket 模式。
3. 最新应用版本是否已经发布或安装。
4. REST 和 WebSocket 域名是否指向同一环境。
5. 自定义 OpenAPI 域名是否同时配置正确 OAuth base URL。
6. 进程是否能访问目标 endpoint。
7. `Start` context 是否已经取消。

在 `Start` 前注册 `OnReady`、`OnReconnecting`、`OnReconnected`、
`OnDisconnected` 和 `OnError`，可以区分首次连接和重连失败。

## 收不到事件

检查：

1. 开发者后台是否开启了对应事件或卡片回调。
2. 应用是否有对应权限。
3. 修改权限或订阅后是否重新安装应用。
4. 机器人是否在测试群中，或者应用能否访问目标文档。
5. 是否在 `Start` 前注册处理器。
6. 测试动作是否发生在 `OnReady` 之后。
7. 事件是否属于 `APP_ID` 对应的同一应用和租户。

消息事件使用 `im.message.receive_v1`。评论处理使用自定义事件
`drive.notice.comment_add_v1`，同时要求应用能够访问文档。

## 收到事件但 `OnMessage` 不执行

runtime 可能在调用处理器前主动过滤消息：

| 过滤项 | 默认行为 |
|---|---|
| 自身消息 | 忽略当前机器人发送的消息 |
| 过期事件 | 忽略超过 30 分钟的消息 |
| 重复事件 | 忽略最近处理过的 message ID |
| 群 allowlist | 配置后拒绝不在列表中的群 |
| 群 mention | 必须 mention 机器人 |
| `@all` | 默认拒绝 |
| 私聊策略 | 默认 open，可配置 |
| 批处理 debounce | 可能等待 600ms；长缓冲内容可能等待 2 秒 |

注册 `OnReject` 可以看到策略原因。使用 `WithSafetyConfig` 时必须从
`types.DefaultChannelConfig()` 开始；局部零值配置可能丢弃所有带时间戳消息。

## 发送失败

先对错误分类：

```go
var channelErr *channel.FeishuChannelError
if errors.As(err, &channelErr) {
	log.Printf("send failed: %s", channelErr.Code)
}
```

| 错误码 | 常见原因 | 处理方式 |
|---|---|---|
| `permission_denied` | 缺少发送权限、凭证无效、应用未安装 | 开通权限、发布并重新安装 |
| `format_error` | post/card JSON 无效，或字段组合不支持 | 校验 payload，每次只发送一种内容 |
| `target_revoked` | 回复目标已撤回或不可用 | 有目标字段时 SDK 会尝试降级新发 |
| `rate_limited` | 上游配额限制 | 降低流量；SDK 已对可重试错误执行重试 |
| `send_timeout` | Context 或 transport 超时 | 检查超时和下游延迟 |
| `unknown` | 未分类网络/API 错误 | 检查脱敏后的上游 code 和 message |

同时确认 `ReceiveID`、`ChatID` 或 `UserID` 至少设置一个。回复也需要目标用于
降级。推荐使用 `ReceiveID` 避免 ID 类型歧义。

## 发送了错误的消息类型

实际类型由内容字段决定，而不是 `MsgType`。设置多个字段时，key/媒体/卡片/
post/名片/贴纸优先于 Markdown，Markdown 又优先于 text。

每次逻辑发送都应创建新的 `SendInput`。上传过程可能写入 key，回复降级可能清空
`ReplyMessageID`，因此不能并发复用同一个 input。

完整顺序见[内容优先级](sending-messages.md)。

## 媒体上传或下载失败

| 现象 | 检查项 |
|---|---|
| `ssrf_blocked` | URL scheme、DNS、私有/保留地址和精确 host allowlist |
| Body 超限 | `Outbound.MediaSource.MaxDownloadBytes`，默认 100 MiB |
| 路径穿越被拒绝 | 移除 `..` segment，并使用可信路径 |
| 音视频时长错误 | 使用有效 Opus/OGG 或 MP4，或显式传入毫秒值 |
| 上传权限错误 | 图片/文件权限以及是否重新安装应用 |
| 下载类型不支持 | 使用 `image`、`file`、`audio`、`video` 或 `media` |
| 内存压力 | URL 类型化媒体和下载都会整体进入内存 |

需要严格控制重定向或内容校验时，使用业务自己的 HTTP Client 获取并验证数据，
再传入 `SourceBytes`。

## 流式回复失败

| 现象 | 原因 |
|---|---|
| `Append is not supported...` | 控制器由 `Card` 创建 |
| `UpdateCard is not supported...` | 当前是 Markdown 流 |
| `stream is closed` | 已经调用 `Close` |
| 创建流失败 | 初始 `Send` 失败 |
| 后续更新失败 | 消息 update/patch 权限、payload、超时或重试耗尽 |
| 更新看起来有延迟 | 100ms 节流/合并，或自定义了更长间隔 |

调用 `Flush` 或 `Close` 并处理错误。卡片控制器 patch 普通交互消息，不提供
CardKit 预创建或 sequence API。

## 自定义域名问题

`WithDomain` 同时设置 REST 和 WebSocket 域名。分离环境可以使用：

```go
ch, err := channel.New(
	appID,
	appSecret,
	channel.WithOpenBaseURL(openBaseURL),
	channel.WithWebSocketDomain(wsDomain),
	channel.WithOAuthBaseURL(oauthBaseURL),
)
```

`WithHTTPClient` 只影响 REST，不会自定义 WebSocket dialer。

## 安全复现

运行本地检查：

```bash
go test ./...
go test -race ./...
go test ./examples/...
CHANNEL_E2E_DRY_RUN=1 go test -tags=e2e ./e2e -run TestChannelE2E -v
```

真实环境覆盖见 [E2E 指南](../../e2e/README.md)。凭证只能放在被忽略的根目录
`.env`，不能粘贴到 issue、日志、截图、fixture 或文档中。

## 相关文档

- [快速开始](quickstart.md)
- [API 参考](reference.md)
- [接收事件](events.md)
- [发送消息](sending-messages.md)
- [媒体](media.md)
