# Streaming Replies

[中文](zh-CN/streaming.md) | [Documentation index](README.md)

`Channel.Stream` sends an initial message and returns a `StreamController`.
The current implementation streams by updating normal messages: Markdown uses
the message update/reply APIs, and cards use the message patch API. It is not a
low-level CardKit preallocation API.

`Stream` is a REST capability and does not require `Start`.

## Markdown Stream

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

When neither `Markdown` nor `Text` is supplied, the initial Markdown content is
`"..."`. Prefer `Markdown` for a Markdown stream. Although a text initial
message is accepted, subsequent stream updates are post/Markdown updates.

## Markdown Controller

| Method | Behavior |
|---|---|
| `Append(ctx, text)` | Appends text and triggers a throttled update |
| `Flush(ctx)` | Immediately sends the accumulated content |
| `Close(ctx)` | Flushes a pending throttled update, then closes the controller |
| `UpdateCard(ctx, card)` | Returns an unsupported-operation error |

The normal throttle interval is 100 ms. Calls inside that interval are
coalesced. An `Append` that performs an immediate update can return its network
error; a coalesced delayed update runs asynchronously, so always finish with
`Flush` or `Close` and handle the returned error.

When accumulated Markdown exceeds `TextChunkLimit`, the controller replies with
a new post chunk and continues updating the newest chunk. Fenced code blocks
are kept balanced across chunks.

## Interactive Card Stream

Start with a complete interactive card:

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

## Card Controller

| Method | Behavior |
|---|---|
| `UpdateCard(ctx, card)` | Serializes and sends a whole-card message patch |
| `Flush(ctx)` | No-op; each accepted card update waits for its patch result |
| `Close(ctx)` | Stops accepting updates and closes the internal queue |
| `Append(ctx, text)` | Returns an unsupported-operation error |

`Close` does not call a CardKit "finish streaming" API. The supplied card JSON
must itself be a valid final card payload.

## Lifecycle and Error Rules

- The initial send occurs inside `Stream`; no controller is returned if it
  fails.
- Do not call `Append` or `UpdateCard` after `Close`.
- `Close` is idempotent.
- Use the same timeout and cancellation policy as normal REST calls.
- Updates use the configured outbound retry policy.
- `Append` and `UpdateCard` belong to different controller modes; unsupported
  method calls return an error instead of changing modes.

For a complete bot, see [`examples/stream_reply`](../examples/stream_reply).

## Related Documents

- [Sending messages](sending-messages.md)
- [Outbound configuration](reference.md#effective-defaults)
- [Troubleshooting streams](troubleshooting.md#streaming-fails)
