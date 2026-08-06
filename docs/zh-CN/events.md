# 接收事件

[English](../events.md) | [文档索引](README.md)

Channel SDK 通过 `Start(ctx)` 建立的 WebSocket 长连接接收入站事件和卡片回调。
处理器应在启动连接前注册。

## 支持的处理器

| 注册方法 | Payload | 平台事件或回调 | 串行作用域 |
|---|---|---|---|
| `OnMessage` | `*NormalizedMessage` | `im.message.receive_v1` | Chat ID |
| `OnReaction` | `*ReactionEvent` | reaction 添加/删除 | Message ID |
| `OnComment` | `*CommentEvent` | `drive.notice.comment_add_v1` | File token |
| `OnBotAdded` | `*BotAddedEvent` | 机器人加入群聊 | Chat ID |
| `OnCardAction` | `*CardActionEvent` | 卡片 action 回调 | Chat ID，缺失时用 Message ID |
| `OnReject` | `*RejectEvent` | 本地合成事件 | 与策略判断同步 |

开发者后台中的事件名称和权限名称可能因语言或租户版本不同而变化。需要开通
对应事件/回调及权限，然后发布或重新安装应用。

## 注册处理器

```go
ch.OnError(func(err error) {
	log.Printf("channel error: %v", err)
})

ch.OnMessage(func(ctx context.Context, msg *channel.NormalizedMessage) error {
	log.Printf("message %s in %s: %s", msg.MessageID, msg.ChatID, msg.Content)
	return nil
})

ch.OnReaction(func(ctx context.Context, event *channel.ReactionEvent) error {
	log.Printf("%s reaction %s", event.Action, event.ReactionType)
	return nil
})

ch.OnCardAction(func(ctx context.Context, event *channel.CardActionEvent) error {
	log.Printf("card action: %s", event.Action.Name)
	return nil
})
```

同类型的多个处理器按注册顺序执行，遇到第一个错误后停止。建议注册
`OnError` 统一观测错误，并保证处理器幂等，因为上游可能在 SDK 内存去重窗口
之外重复投递事件。

## 消息归一化

`OnMessage` 将 `P2MessageReceiveV1` 转成稳定模型：

| 字段 | 含义 |
|---|---|
| `EventID` | 原始事件 ID |
| `MessageID` | 原始消息 ID |
| `ChatID` | 会话 ID |
| `ChatType` | 通常为 `group` 或 `p2p` |
| `UserID` | 优先为发送者 open ID，否则为 user ID |
| `Content` | 原始消息类型对应的扁平文本 |
| `RawContentType` | 原始类型，如 `text`、`post`、`image`、`file`、`media` |
| `Mentions` | mention key、ID、显示名和机器人标记 |
| `MentionAll` | 是否包含 `@all` |
| `MentionedBot` | 是否 mention 当前机器人 |
| `Resources` | file key、文件名、时长、封面图等下载描述 |
| `CreateTimeMs` | Unix 毫秒时间戳 |
| `RawEvent` | 原始生成 SDK 事件 |

`Content` 适合常见机器人逻辑。需要精确线格式时，使用 `RawContentType`、
`Resources` 和 `RawEvent`。

## 消息处理顺序

消息进入 `OnMessage` 前，runtime 依次执行：

1. 归一化。
2. 根据当前机器人身份过滤自身消息。
3. 过滤过期事件。
4. 内存去重。
5. 群聊/私聊策略判断。
6. 单消息处理锁。
7. 按 chat debounce、批处理并串行执行处理器。

在第 5 步被拒绝的消息不会进入 `OnMessage`。如果注册了 `OnReject`，会收到
对应拒绝事件。

## 批处理

同一会话在 debounce 窗口内到达的消息可能被合并：

- 内容用空行连接；
- `MessageID` 等标量字段来自最后一条消息；
- `MentionAll` 和 `MentionedBot` 做 OR；
- resources 和 mentions 去重；
- 处理器只对合并消息执行一次。

默认窗口为 600ms；缓冲内容达到 1000 个 UTF-8 字节后使用 2 秒窗口；达到
8 条消息或 4000 字节时立即 flush。修改前请阅读
[策略与安全](policy-and-safety.md)。

## 顺序与并发

同一个串行作用域中的任务按顺序执行；不同 chat、message ID 或 file token
可能并发处理。卡片 action 会与同一 chat pipeline 中的任务串行。

因此处理器需要：

- 对共享 map 或 slice 的并发写加锁；
- 响应 `ctx` 取消，并为下游调用设置超时；
- 返回错误，而不是只记录日志后继续；
- 能够安全处理不同会话的并发调用。

## 事件 Payload

### `ReactionEvent`

| 字段 | 含义 |
|---|---|
| `EventID` | 事件 ID |
| `MessageID` | 被 reaction 的消息 |
| `ReactionType` | emoji 类型，如 `SMILE` |
| `OperatorType` | 原始操作者类型 |
| `UserID` | 用户或应用标识 |
| `Action` | `added` 或 `removed` |
| `CreateTimeMs` | 事件时间 |
| `RawEvent` | 原始添加/删除事件 |

### `CommentEvent`

| 字段 | 含义 |
|---|---|
| `EventID` | 事件 ID |
| `CommentID` | 评论 ID |
| `FileToken` / `FileType` | 被评论文档目标 |
| `Operator` | 事件中提供的 open ID、user ID 和 union ID |
| `ReplyID` | 评论回复 ID |
| `MentionedBot` | 事件是否标记 mention 机器人 |
| `Timestamp` | 事件时间戳 |
| `RawEvent` | 原始自定义事件 |

文档评论投递需要订阅评论事件、开通合适的云空间权限，并确保应用能够访问目标
文档。

### `BotAddedEvent`

| 字段 | 含义 |
|---|---|
| `EventID` | 事件 ID |
| `ChatID` / `ChatName` | 目标群聊 |
| `UserID` | 事件提供时，表示添加机器人的用户 |
| `CreateTimeMs` | 事件时间 |
| `RawEvent` | 原始事件 |

### `CardActionEvent`

| 字段 | 含义 |
|---|---|
| `EventID` | 回调事件 ID |
| `MessageID` / `ChatID` | 来源消息和群聊 |
| `Operator` | 用户和租户标识 |
| `Action` | tag、name、option、value、表单值、输入、选项和 checked 状态 |
| `Context` | 回调 URL、preview token、open message ID 和 open chat ID |
| `Token`、`Host`、`DeliveryType` | 事件提供时的回调元数据 |
| `RawEvent` | 原始卡片回调 |

### `RejectEvent`

| 字段 | 含义 |
|---|---|
| `MessageID` | 被拒绝消息 |
| `ChatID` | 来源会话 |
| `SenderID` | 发送者标识 |
| `Reason` | 稳定策略原因 |

原因包括 `group_not_allowed`、`no_mention`、`mention_all_blocked`、
`dm_disabled` 和 `sender_not_allowed`。

## 生命周期事件

| 处理器 | 时机 |
|---|---|
| `OnReady` | WebSocket ready；此前会尝试解析机器人身份 |
| `OnError` | WebSocket 或 Channel 处理器报告错误 |
| `OnReconnecting` | 开始重连 |
| `OnReconnected` | 重连成功 |
| `OnDisconnected` | 连接关闭 |

也可以直接调用 `GetBotIdentity(ctx)`。缓存有效时返回缓存；刷新失败但有旧缓存时
返回旧缓存；从未成功解析身份且请求失败时返回 nil。

## 相关文档

- [快速开始](quickstart.md)
- [API 参考](reference.md)
- [策略与安全](policy-and-safety.md)
- [事件问题排查](troubleshooting.md)
