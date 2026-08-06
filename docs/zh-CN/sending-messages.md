# 发送消息

[English](../sending-messages.md) | [文档索引](README.md)

`Channel.Send` 接收一个 `SendInput`，解析目标，生成一种逻辑内容，对符合条件的
失败执行重试，并返回 `SendResult`。它使用 REST API，不依赖 WebSocket
`Start`。

## 基础发送

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

result, err := ch.Send(ctx, &channel.SendInput{
	ReceiveID: "oc_xxx",
	Text:      "hello",
})
if err != nil {
	return err
}
log.Printf("message_id=%s chat_id=%s", result.MessageID, result.ChatID)
```

网络调用应始终使用符合业务要求的 deadline。

## 目标解析

至少需要设置一个目标字段。

| 字段 | 优先级 | 解析方式 |
|---|---:|---|
| `ReceiveID` | 1 | 根据值推导 ID 类型 |
| `ChatID` | 2 | 使用 `receive_id_type=chat_id` |
| `UserID` | 3 | 兼容回退，使用 `receive_id_type=open_id` |

`ReceiveID` 推导规则：

| 值形态 | `receive_id_type` |
|---|---|
| `oc_...` | `chat_id` |
| `ou_...` | `open_id` |
| `on_...` | `union_id` |
| 包含 `@` | `email` |
| 其它非空值 | `user_id` |

新代码推荐使用 `ReceiveID`。`UserID` 是兼容字段，其值会按 open ID 发送；如果
手里是真正的 `user_id`，应使用 `ReceiveID` 让 SDK 推导类型。

## 回复消息

设置 `ReplyMessageID` 后使用回复 API：

```go
result, err := ch.Send(ctx, &channel.SendInput{
	ReceiveID:      msg.ChatID,
	ReplyMessageID: msg.MessageID,
	Markdown:       "**received:** " + msg.Content,
})
```

回复仍然必须提供目标字段。如果回复目标已撤回或不可用，SDK 会清空
`ReplyMessageID`，向已解析目标新发一条消息。

## 内容优先级

每次只设置一种逻辑内容。如果多个字段同时非空，SDK 按以下顺序选择第一个：

1. `ImageKey`，或上传后的 `ImagePath`/图片 `Media`
2. `AudioKey`，或上传后的音频 `Media`
3. `VideoKey`，或上传后的视频 `Media`
4. `FileKey`，或上传后的 `FilePath`/文件 `Media`
5. `Card`
6. `Post`
7. `ShareChatID`
8. `ShareUserID`
9. `StickerFileKey`
10. `Markdown`
11. `Text`

`MsgType` 仅为兼容保留，不能单独提供消息内容，也不会覆盖上述优先级。建议设置
明确内容字段并让 `MsgType` 为空。

## 文本和 Markdown

纯文本：

```go
_, err := ch.Send(ctx, &channel.SendInput{
	ReceiveID: "oc_xxx",
	Text:      "plain text",
})
```

Markdown 会包装成包含原生 `md` 元素的飞书/Lark post：

```go
_, err := ch.Send(ctx, &channel.SendInput{
	ReceiveID: "oc_xxx",
	Title:     "Build result",
	Markdown:  "Status: **passed**\n\n```go\nfmt.Println(\"ok\")\n```",
})
```

纯文本超过 `Outbound.TextChunkLimit` 时按 rune 拆分。Markdown 在行边界按
UTF-8 字节阈值拆分，并保持 fenced code block 在各分片中闭合。默认阈值为
3500；单个 Markdown 行超过阈值时，不会在行内继续拆分。

`SendResult.MessageID` 是第一条消息 ID；发生拆分时，`ChunkIDs` 按顺序包含
全部消息 ID。

## Mention

```go
_, err := ch.Send(ctx, &channel.SendInput{
	ReceiveID: "oc_xxx",
	Text:      "please review",
	Mentions: []channel.Mention{
		{
			UserID: "ou_xxx",
			Name:   "Alice",
		},
	},
})
```

出站 mention 必须设置 `Mention.UserID`。SDK 会把 mention 前置到文本或第一个
Markdown 分片。`OpenID` 会用于入站归一化，但当前出站 mention builder 不读取
这个字段。

## 富文本和交互卡片

`Post` 接收已经序列化的 post payload：

```go
post := `{"zh_cn":{"title":"Status","content":[[{"tag":"text","text":"ready"}]]}}`

_, err := ch.Send(ctx, &channel.SendInput{
	ReceiveID: "oc_xxx",
	Post:      post,
})
```

`Card` 接收已经序列化的交互卡片 payload：

```go
card := `{
  "schema": "2.0",
  "body": {
    "elements": [
      {"tag": "markdown", "content": "**Choose an action**"}
    ]
  }
}`

_, err := ch.Send(ctx, &channel.SendInput{
	ReceiveID: "oc_xxx",
	Card:      card,
})
```

接收卡片 action 需要注册 `OnCardAction` 并启动 WebSocket。完整代码见
[`card_action`](../../examples/card_action)。

## 已有媒体 Key

```go
_, err := ch.Send(ctx, &channel.SendInput{
	ReceiveID: "oc_xxx",
	ImageKey:  "img_xxx",
})

_, err = ch.Send(ctx, &channel.SendInput{
	ReceiveID: "oc_xxx",
	FileKey:   "file_xxx",
})

_, err = ch.Send(ctx, &channel.SendInput{
	ReceiveID: "oc_xxx",
	AudioKey:  "file_xxx",
})

_, err = ch.Send(ctx, &channel.SendInput{
	ReceiveID: "oc_xxx",
	VideoKey:  "file_xxx",
})
```

本地文件、字节和 URL 发送方式见[媒体](media.md)。

## 群名片、用户名片和贴纸

```go
_, err := ch.Send(ctx, &channel.SendInput{
	ReceiveID:  "oc_xxx",
	ShareChatID: "oc_shared_chat",
})

_, err = ch.Send(ctx, &channel.SendInput{
	ReceiveID:  "oc_xxx",
	ShareUserID: "ou_shared_user",
})

_, err = ch.Send(ctx, &channel.SendInput{
	ReceiveID:      "oc_xxx",
	StickerFileKey: "file_v3_xxx",
})
```

应用必须有权访问被分享目标或贴纸 key。

## 重试与降级

默认发送/更新最多尝试 3 次。可重试的 `rate_limited` 和 `unknown` 错误使用
从 500ms 开始的指数退避，`context.Context` 可以取消退避等待。

重试后：

- 回复目标失效会降级为新发消息；
- 富内容出现 `format_error` 时，只有存在 `Markdown` 或 `Text` 才能降级为
  纯文本；
- 权限、格式、SSRF、目标和超时错误不会继续执行内置重试。

SDK 会返回分类后的 `*FeishuChannelError`，不会把最终错误隐藏在
`SendResult.Error` 中。

## 相关文档

- [API 参考](reference.md)
- [流式回复](streaming.md)
- [媒体](media.md)
- [发送问题排查](troubleshooting.md)
