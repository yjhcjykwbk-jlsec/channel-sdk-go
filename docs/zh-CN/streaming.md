# 流式回复

[English](../streaming.md) | [文档索引](README.md)

`Channel.Stream` 会先发送一条初始消息，然后返回 `StreamController`。当前实现
通过更新普通消息实现流式效果：Markdown 使用消息 update/reply API，卡片使用
消息 patch API；它不是低层 CardKit 预创建接口。

`Stream` 属于 REST 能力，不依赖 `Start`。

## Markdown 流

```go
stream, err := ch.Stream(ctx, &channel.SendInput{
	ReceiveID: msg.ChatID,
	Title:     "Assistant",
	Markdown:  "Thinking...",
})
if err != nil {
	return err
}

if err := stream.Append(ctx, "\n\nFirst result."); err != nil {
	return err
}
if err := stream.Append(ctx, "\n\nSecond result."); err != nil {
	return err
}
if err := stream.Flush(ctx); err != nil {
	return err
}
return stream.Close(ctx)
```

如果 `Markdown` 和 `Text` 都为空，初始 Markdown 为 `"..."`。Markdown 流推荐
使用 `Markdown`；虽然初始消息可以是文本，但后续流更新会使用 post/Markdown。

## Markdown 控制器

| 方法 | 行为 |
|---|---|
| `Append(ctx, text)` | 追加文本并触发节流更新 |
| `Flush(ctx)` | 立即发送当前累计内容 |
| `Close(ctx)` | flush 尚未执行的节流更新，然后关闭控制器 |
| `UpdateCard(ctx, card)` | 返回不支持操作错误 |

正常节流间隔为 100ms，间隔内调用会合并。立即执行更新的 `Append` 可以返回网络
错误；被合并的延迟更新是异步执行，因此结束时应调用 `Flush` 或 `Close` 并处理
错误。

累计 Markdown 超过 `TextChunkLimit` 后，控制器会回复一个新的 post 分片，后续
继续更新最新分片。Fenced code block 会保持闭合。

## 交互卡片流

使用一个完整的交互卡片作为初始值：

```go
initialCard := `{
  "schema": "2.0",
  "body": {
    "elements": [
      {
        "tag": "markdown",
        "element_id": "content",
        "content": "Starting..."
      }
    ]
  }
}`

stream, err := ch.Stream(ctx, &channel.SendInput{
	ReceiveID: "oc_xxx",
	Card:      initialCard,
})
if err != nil {
	return err
}

nextCard := `{
  "schema": "2.0",
  "body": {
    "elements": [
      {
        "tag": "markdown",
        "element_id": "content",
        "content": "Finished"
      }
    ]
  }
}`

if err := stream.UpdateCard(ctx, nextCard); err != nil {
	return err
}
return stream.Close(ctx)
```

## 卡片控制器

| 方法 | 行为 |
|---|---|
| `UpdateCard(ctx, card)` | 串行发送完整卡片 message patch |
| `Flush(ctx)` | 无操作；每次已接收更新都会等待 patch 结果 |
| `Close(ctx)` | 停止接收新更新并关闭内部队列 |
| `Append(ctx, text)` | 返回不支持操作错误 |

`Close` 不会调用 CardKit 的“结束流式”接口，传入的卡片 JSON 本身必须是有效的
最终卡片。

## 生命周期与错误规则

- 初始发送在 `Stream` 内完成；失败时不会返回控制器。
- `Close` 后不能再调用 `Append` 或 `UpdateCard`。
- `Close` 可以重复调用。
- 使用与普通 REST 请求相同的超时和取消策略。
- 更新使用出站重试配置。
- `Append` 和 `UpdateCard` 属于不同模式；错误调用会返回错误，不会切换模式。

完整机器人示例见
[`examples/stream_reply`](../../examples/stream_reply)。

## 相关文档

- [发送消息](sending-messages.md)
- [出站配置](reference.md)
- [流式回复问题排查](troubleshooting.md)
