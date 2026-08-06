# Channel API 参考

[English](../reference.md) | [文档索引](README.md)

本文档描述 `github.com/larksuite/channel-sdk-go` 面向业务代码的稳定公开 API。
首次接入请先阅读[快速开始](quickstart.md)。

## 入口

```go
import channel "github.com/larksuite/channel-sdk-go"

ch, err := channel.New(appID, appSecret, opts...)
```

`appID` 不能为空。除非配置了非 nil 的 `WithClientAssertionProvider`，
`appSecret` 也不能为空。`New` 会创建 REST Client、事件分发器、WebSocket
Client 和 Channel runtime。

返回的 `channel.Channel` 是高阶公开接口，同时提供
`RawClient() *lark.Client`，用于调用 Channel 未覆盖的 OpenAPI。

## 构造选项

| 选项 | 值类型 | 作用范围 | 行为 |
|---|---|---|---|
| `WithLogger` | `larkcore.Logger` | REST、WebSocket、事件分发、Channel | 使用指定 logger |
| `WithLogLevel` | `larkcore.LogLevel` | REST、WebSocket、事件分发 | 设置底层 SDK 日志级别 |
| `WithDomain` | `string` | REST 和 WebSocket | 同时设置 OpenAPI 和长连接域名 |
| `WithOpenBaseURL` | `string` | REST | 只覆盖 OpenAPI base URL |
| `WithWebSocketDomain` | `string` | WebSocket | 只覆盖长连接域名 |
| `WithOAuthBaseURL` | `string` | REST token 流程 | 覆盖 OAuth/token base URL |
| `WithHTTPClient` | `larkcore.HttpClient` | REST | 注入自定义 REST HTTP Client |
| `WithReqTimeout` | `time.Duration` | REST | 设置底层 REST 请求超时 |
| `WithTokenCache` | `larkcore.Cache` | REST token 流程 | 注入 tenant token 缓存 |
| `WithEnableTokenCache` | `bool` | REST token 流程 | 开启或关闭底层 token 缓存 |
| `WithHeaders` | `http.Header` | REST 和 WebSocket | 添加公共请求头 |
| `WithSource` | `string` | REST 和 WebSocket | 设置 SDK source 标识 |
| `WithClientAssertionProvider` | `larkcore.ClientAssertionProvider` | REST 和 WebSocket 鉴权 | 使用 client assertion，并允许 `appSecret` 为空 |
| `WithSafetyConfig` | `channel.SafetyConfig` | 入站 runtime | 替换完整安全配置 |
| `WithPolicyConfig` | `channel.PolicyConfig` | 入站 runtime | 设置消息准入策略 |
| `WithOutboundConfig` | `channel.OutboundConfig` | 发送和流式回复 | 替换完整出站配置 |
| `WithBotIdentityCacheConfig` | `channel.BotIdentityCacheConfig` | 机器人身份 | 替换身份缓存配置 |

`WithHTTPClient` 不会替换 WebSocket dialer。如果需要控制长连接域名，使用
`WithWebSocketDomain`；需要更底层的 transport 控制时，使用
`oapi-sdk-go/v3` 的 WebSocket 能力。

## 修改配置时保留默认值

`WithSafetyConfig` 和 `WithOutboundConfig` 会整块替换对应配置。只修改一个
字段时，应从 `types.DefaultChannelConfig()` 开始：

```go
import (
	channel "github.com/larksuite/channel-sdk-go"
	channeltypes "github.com/larksuite/channel-sdk-go/types"
)

defaults := channeltypes.DefaultChannelConfig()
defaults.Outbound.TextChunkLimit = 2000
defaults.Safety.Batch.DelayMs = 300 * time.Millisecond

ch, err := channel.New(
	appID,
	appSecret,
	channel.WithOutboundConfig(defaults.Outbound),
	channel.WithSafetyConfig(defaults.Safety),
)
```

不要从零值构造局部 `SafetyConfig`。特别是，过期窗口为零会把带时间戳的消息
视为过期，批处理上限为零会导致立即 flush。零值 `OutboundConfig` 还会关闭
文本分片，并使流控制器使用内部兜底节流间隔，而不是正常的 100ms 默认值。

## 生效默认值

| 字段 | 默认值 | 含义 |
|---|---:|---|
| `Safety.Dedup.MaxEntries` | `10000` | 内存去重 key 上限 |
| `Safety.Dedup.SweepIntervalMs` | `1h` | 去重 key 保留时间 |
| `Safety.StaleMessageWindowMs` | `30m` | 丢弃更早的带时间戳消息 |
| `Safety.Batch.DelayMs` | `600ms` | 普通 debounce 窗口 |
| `Safety.Batch.LongThresholdChars` | `1000` | 缓冲内容达到该字节数后切换长消息延迟 |
| `Safety.Batch.LongDelayMs` | `2s` | 长消息 debounce 窗口 |
| `Safety.Batch.MaxMessages` | `8` | 达到该消息数时 flush |
| `Safety.Batch.MaxChars` | `4000` | 缓冲内容达到该字节数时 flush |
| `Policy.GroupAllowlist` | 空 | 允许所有群 |
| `Policy.RequireMention` | `true` | 群消息必须 mention 机器人 |
| `Policy.RespondToMentionAll` | `false` | 拒绝包含 `@all` 的群消息 |
| `Policy.DMMode` | `"open"` | 允许私聊消息 |
| `Outbound.TextChunkLimit` | `3500` | 纯文本按 rune；Markdown 在行边界按 UTF-8 字节阈值拆分 |
| `Outbound.StreamThrottleMs` | `100ms` | Markdown 流更新最小间隔 |
| `Outbound.StreamThrottleChars` | `10` | 当前版本保留字段，尚不影响 runtime 行为 |
| `Outbound.MediaSource.MaxDownloadBytes` | `100 MiB` | URL 媒体最多读取的 body 大小 |
| `Outbound.Retry.MaxAttempts` | `3` | 发送/更新总尝试次数 |
| `Outbound.Retry.BaseDelayMs` | `500ms` | 首次重试退避 |
| `BotIdentityCache.TTL` | `30m` | 成功身份缓存时长 |
| `BotIdentityCache.MinRefreshInterval` | `1m` | 身份刷新失败后的最小重试间隔 |

`BotIdentityCache.MinRefreshInterval` 在 0 到 30 秒之间时会提升到 30 秒。媒体
限制非正数时回退到 100 MiB；重试次数和基础延迟非正数时回退到 3 次和
500ms。

## Channel 方法

| 方法 | 是否依赖 `Start` | 说明 |
|---|---:|---|
| `Send(ctx, input)` | 否 | 发送或回复一条逻辑消息 |
| `Stream(ctx, input)` | 否 | 发送初始消息并返回流控制器 |
| `DownloadFile(ctx, fileKey, mediaType)` | 否 | 将图片、文件、音频或视频下载到内存 |
| `OnMessage(handler)` | 事件投递依赖 | 注册归一化消息处理器 |
| `OnReaction(handler)` | 事件投递依赖 | 注册 reaction 添加/删除处理器 |
| `OnComment(handler)` | 事件投递依赖 | 注册文档评论处理器 |
| `OnBotAdded(handler)` | 事件投递依赖 | 注册机器人入群处理器 |
| `OnCardAction(handler)` | 事件投递依赖 | 注册长连接卡片 action 处理器 |
| `OnReject(handler)` | 事件投递依赖 | 注册本地策略拒绝处理器 |
| `OnReady(handler)` | 是 | WebSocket ready 后执行 |
| `OnError(handler)` | 否 | 观察 WebSocket 和处理器错误 |
| `OnReconnecting(handler)` | 是 | 观察开始重连 |
| `OnReconnected(handler)` | 是 | 观察重连成功 |
| `OnDisconnected(handler)` | 是 | 观察连接断开 |
| `Start(ctx)` | 不适用 | 启动并阻塞等待 WebSocket 长连接 |
| `Stop(ctx)` | 否 | 关闭 WebSocket 并释放本地 pipeline，可重复调用 |
| `UpdatePolicy(cfg)` | 否 | 局部更新运行时策略 |
| `GetPolicy()` | 否 | 返回当前策略 |
| `GetBotIdentity(ctx)` | 否 | 获取/缓存机器人身份；首次获取彻底失败时返回 nil |
| `RawClient()` | 否 | 返回底层 `*lark.Client` |

在 `Start` 前注册入站事件和生命周期处理器。Channel 应按单生命周期使用：
调用 `Stop` 后重新构造，而不是重新启动原实例。

## `SendInput`

目标选择：

| 字段 | 优先级 | 行为 |
|---|---:|---|
| `ReceiveID` | 1 | 推荐目标，根据前缀或邮箱语法推导类型 |
| `ChatID` | 2 | 使用 `receive_id_type=chat_id` |
| `UserID` | 3 | 兼容回退，使用 `receive_id_type=open_id` |
| `ReplyMessageID` | 独立 | 使用回复 API，但仍需要目标字段用于降级新发 |

内容选择：

| 字段 | 结果类型 | 说明 |
|---|---|---|
| `Text` | `text` | 自动分片，前置 mention |
| `Markdown` | `post` | 生成包含 `tag=md` 的 post，并自动分片 |
| `Title` | 无 | Markdown 和 Markdown 流的 post 标题 |
| `ImageKey` | `image` | 已上传图片 |
| `FileKey` | `file` | 已上传文件 |
| `AudioKey` | `audio` | 已上传音频 |
| `VideoKey` | `media` | 已上传视频 |
| `Card` | `interactive` | 序列化后的卡片 JSON |
| `Post` | `post` | 序列化后的飞书/Lark post JSON |
| `ShareChatID` | `share_chat` | 群名片目标 |
| `ShareUserID` | `share_user` | 用户名片目标 |
| `StickerFileKey` | `sticker` | 已有贴纸 file key |
| `ImagePath` | `image` | `ImageKey` 为空时上传本地图片 |
| `FilePath` | `file` | `FileKey` 为空时上传本地文件 |
| `Media` | 由 `Media.Kind` 决定 | 从字节、本地路径或 URL 上传 |
| `Mentions` | 无 | 前置到文本或第一个 Markdown 分片 |
| `MsgType` | 兼容提示 | 实际消息类型由内容字段决定 |

每次只提供一种逻辑内容。设置多个内容字段时，SDK 按
[发送消息](sending-messages.md)中记录的优先级处理。

## `SendResult`

| 字段 | 说明 |
|---|---|
| `MessageID` | 逻辑发送产生的第一条消息 ID |
| `ChunkIDs` | 长文本/Markdown 分片后的全部消息 ID；只有一片时省略 |
| `ChatID` | 发送 API 返回的群 ID |
| `Error` | 保留字段；当前 `Send` 失败通过 Go `error` 返回 |

## `NormalizedMessage`

| 字段 | 说明 |
|---|---|
| `EventID` | 原始事件 ID |
| `MessageID` | 消息 ID；合并批次时来自最后一条消息 |
| `ChatID` | 会话 ID |
| `ChatType` | 通常为 `group` 或 `p2p` |
| `UserID` | 优先为发送者 open ID，否则为 user ID |
| `Content` | 扁平化后的归一化内容 |
| `RawContentType` | 原始消息类型 |
| `Mentions` | 归一化 mention |
| `MentionAll` | 是否包含 `@all` |
| `MentionedBot` | 是否 mention 当前机器人 |
| `Resources` | 图片、文件、音频、视频、贴纸等资源描述 |
| `CreateTimeMs` | Unix 毫秒时间戳 |
| `RawEvent` | 原始 `P2MessageReceiveV1` |

事件 payload 和分发语义见[接收事件](events.md)。

## `UploadInput`

| 字段 | 说明 |
|---|---|
| `Kind` | `MediaKindImage`、`MediaKindFile`、`MediaKindAudio` 或 `MediaKindVideo` |
| `SourceBytes` | 非 nil 时为最高优先级来源 |
| `SourcePath` | bytes 为 nil 时使用的本地路径 |
| `SourceURL` | bytes 和 path 都未设置时使用的 HTTP(S) URL |
| `FileName` | 上传文件名；为空时使用对应类型默认名 |
| `Duration` | 音视频时长，单位毫秒；省略时尝试从 Opus/OGG 或 MP4 中解析 |

来源优先级和安全规则见[媒体](media.md)。

## 错误模型

`Send` 和流式更新会把底层失败分类成 `*channel.FeishuChannelError`。

| 错误码 | 含义 | 内置重试 |
|---|---|---:|
| `target_revoked` | 回复目标已撤回、不可用或不匹配 | 否；回复会尝试降级为新消息 |
| `permission_denied` | 凭证或权限不足 | 否 |
| `format_error` | 消息 payload 被拒绝 | 否；有文本来源时富内容可降级 |
| `rate_limited` | 上游限流 | 是 |
| `ssrf_blocked` | URL 媒体未通过安全检查 | 否 |
| `send_timeout` | 请求超时 | 否 |
| `unknown` | 未分类 transport 或上游错误 | 是 |

使用 `errors.As` 或辅助函数：

```go
var channelErr *channel.FeishuChannelError
if errors.As(err, &channelErr) {
	log.Printf("channel error code: %s", channelErr.Code)
}

if channel.IsRetryable(err) {
	// SDK 已经耗尽配置的内置重试。
}
```

根包公开 `ClassifyError`、`IsRetryable` 和 `IsFormatError`。
`types.IsReplyTargetGone` 位于公开 `types` 包。

## 相关文档

- [接收事件](events.md)
- [发送消息](sending-messages.md)
- [流式回复](streaming.md)
- [媒体](media.md)
- [策略与安全](policy-and-safety.md)
