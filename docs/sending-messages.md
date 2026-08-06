# Sending Messages

[中文](zh-CN/sending-messages.md) | [Documentation index](README.md)

`Channel.Send` accepts one `SendInput`, resolves a target, materializes one
logical content shape, retries eligible failures, and returns a `SendResult`.
It uses REST APIs and does not require the WebSocket connection to be started.

## Basic Send

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

Always give network operations a deadline appropriate for the application.

## Target Resolution

At least one target field is required.

| Field | Priority | Resolution |
|---|---:|---|
| `ReceiveID` | 1 | Infers the ID type from the value |
| `ChatID` | 2 | Uses `receive_id_type=chat_id` |
| `UserID` | 3 | Compatibility fallback using `receive_id_type=open_id` |

`ReceiveID` inference:

| Shape | `receive_id_type` |
|---|---|
| `oc_...` | `chat_id` |
| `ou_...` | `open_id` |
| `on_...` | `union_id` |
| Contains `@` | `email` |
| Any other non-empty value | `user_id` |

Prefer `ReceiveID` for new code because its type is inferred. `UserID` is a
legacy convenience field whose value is sent as an open ID; use `ReceiveID`
when the value is a real `user_id`.

## Replies

Set `ReplyMessageID` to use the reply API:

```go
result, err := ch.Send(ctx, &channel.SendInput{
	ReceiveID:      msg.ChatID,
	ReplyMessageID: msg.MessageID,
	Markdown:       "**received:** " + msg.Content,
})
```

A target field is still required. If the reply target is recalled or otherwise
unavailable, the SDK clears `ReplyMessageID` and sends a new message to the
resolved target.

## Content Precedence

Set one logical content shape. If several are non-empty, the first one in this
order wins:

1. `ImageKey` or an uploaded `ImagePath`/image `Media`
2. `AudioKey` or uploaded audio `Media`
3. `VideoKey` or uploaded video `Media`
4. `FileKey` or an uploaded `FilePath`/file `Media`
5. `Card`
6. `Post`
7. `ShareChatID`
8. `ShareUserID`
9. `StickerFileKey`
10. `Markdown`
11. `Text`

`MsgType` is retained for compatibility, but it does not by itself supply
message content and it does not override this precedence. Prefer setting the
typed convenience field and leaving `MsgType` empty.

## Text and Markdown

Plain text:

```go
_, err := ch.Send(ctx, &channel.SendInput{
	ReceiveID: "oc_xxx",
	Text:      "plain text",
})
```

Markdown is wrapped in a Feishu/Lark post containing a native `md` element:

```go
_, err := ch.Send(ctx, &channel.SendInput{
	ReceiveID: "oc_xxx",
	Title:     "Build result",
	Markdown:  "Status: **passed**\n\n```go\nfmt.Println(\"ok\")\n```",
})
```

Plain text longer than `Outbound.TextChunkLimit` is split by runes. Markdown is
split at line boundaries using a UTF-8 byte threshold and preserves fenced code
blocks across chunks. The default threshold is 3500. One Markdown line longer
than the threshold is not split inside that line.

`SendResult.MessageID` is the first message ID. When splitting occurs,
`ChunkIDs` contains all generated message IDs in order.

## Mentions

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

For outbound mentions, `Mention.UserID` must be populated. The SDK prepends
mention elements to text or to the first Markdown chunk. `OpenID` is normalized
for inbound data but is not used by the current outbound mention builder.

## Post and Interactive Card

Use `Post` for an already serialized post payload:

```go
post := `{"zh_cn":{"title":"Status","content":[[{"tag":"text","text":"ready"}]]}}`

_, err := ch.Send(ctx, &channel.SendInput{
	ReceiveID: "oc_xxx",
	Post:      post,
})
```

Use `Card` for a serialized interactive-card payload:

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

Register `OnCardAction` and start the WebSocket connection to receive card
actions. See the runnable [`card_action`](../examples/card_action) example.

## Existing Media Keys

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

For local files, bytes, and URLs, see [Media](media.md).

## Chat, User, and Sticker Shares

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
	ReceiveID:     "oc_xxx",
	StickerFileKey: "file_v3_xxx",
})
```

The application must have access to the shared target or sticker key.

## Retries and Fallbacks

The default send/update loop makes at most three attempts. Retryable
`rate_limited` and `unknown` errors use exponential delays starting at 500 ms.
The `context.Context` can cancel the backoff.

After retries:

- a revoked reply target falls back to a new message;
- a rich-content `format_error` falls back to plain text only when `Markdown`
  or `Text` is available;
- permission, format, SSRF, target, and timeout errors are returned without
  further built-in retry.

The SDK returns a classified `*FeishuChannelError`; it does not hide the final
failure in `SendResult.Error`.

## Related Documents

- [API reference](reference.md#sendinput)
- [Streaming replies](streaming.md)
- [Media](media.md)
- [Troubleshooting sending](troubleshooting.md#sending-fails)
