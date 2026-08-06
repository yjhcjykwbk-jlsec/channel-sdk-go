# Receiving Events

[中文](zh-CN/events.md) | [Documentation index](README.md)

The Channel SDK receives inbound events and card callbacks through the
WebSocket long connection created by `Start(ctx)`. Register handlers before
starting the connection.

## Supported Handlers

| Registration | Payload | Platform event or callback | Serialization scope |
|---|---|---|---|
| `OnMessage` | `*NormalizedMessage` | `im.message.receive_v1` | Chat ID |
| `OnReaction` | `*ReactionEvent` | reaction created/deleted | Message ID |
| `OnComment` | `*CommentEvent` | `drive.notice.comment_add_v1` | File token |
| `OnBotAdded` | `*BotAddedEvent` | bot added to chat | Chat ID |
| `OnCardAction` | `*CardActionEvent` | card action callback | Chat ID, then message ID |
| `OnReject` | `*RejectEvent` | Local synthetic event | Synchronous with policy evaluation |

Event names and permission labels can differ by console locale or tenant
version. Enable the matching event/callback and its required permission, then
publish or reinstall the app.

## Register Handlers

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

Multiple handlers of the same type run in registration order. Processing stops
at the first error for that event. Register `OnError` for centralized
observability and make handlers idempotent because upstream delivery can be
repeated outside the SDK's in-memory deduplication window.

## Message Normalization

`OnMessage` converts `P2MessageReceiveV1` into one stable model:

| Field | Meaning |
|---|---|
| `EventID` | Original event ID |
| `MessageID` | Original message ID |
| `ChatID` | Conversation ID |
| `ChatType` | Usually `group` or `p2p` |
| `UserID` | Sender open ID when available, otherwise sender user ID |
| `Content` | Flattened text representation for the original message type |
| `RawContentType` | Original type such as `text`, `post`, `image`, `file`, or `media` |
| `Mentions` | Mention key, IDs, display name, and bot marker |
| `MentionAll` | Whether `@all` was present |
| `MentionedBot` | Whether the current bot identity was mentioned |
| `Resources` | Download descriptors such as file key, filename, duration, and cover image |
| `CreateTimeMs` | Unix timestamp in milliseconds |
| `RawEvent` | Original generated SDK event |

`Content` is intended for common bot logic. Use `RawContentType`, `Resources`,
and `RawEvent` when exact wire-level details matter.

## Message Processing Order

Before a message reaches `OnMessage`, the runtime applies:

1. Normalization.
2. Self-message filtering using the current bot identity.
3. Stale-event filtering.
4. In-memory duplicate suppression.
5. Group/DM policy evaluation.
6. Per-message processing lock.
7. Per-chat debounce, batching, and serialized handler execution.

Messages rejected at step 5 do not reach `OnMessage`. They produce
`OnReject` when that handler is registered.

## Batching

Messages arriving in the same chat inside the configured debounce window can
be merged:

- content is joined with a blank line;
- the last message supplies scalar fields such as `MessageID`;
- `MentionAll` and `MentionedBot` are ORed;
- resources and mentions are deduplicated;
- handlers run once for the merged message.

The default window is 600 ms, or 2 seconds after buffered content reaches 1000
UTF-8 bytes. A batch flushes at 8 messages or 4000 bytes. See
[Policy and safety](policy-and-safety.md) before customizing these values.

## Ordering and Concurrency

Work in the same serialization scope runs in order. Different chats, message
IDs, or file tokens can be processed concurrently. A card action for a chat is
serialized with work using the same chat pipeline.

Handlers must therefore:

- avoid unsynchronized writes to shared maps or slices;
- respect `ctx` cancellation and set timeouts on downstream calls;
- return errors instead of only logging and continuing;
- remain safe when separate chats are handled concurrently.

## Event Payloads

### `ReactionEvent`

| Field | Meaning |
|---|---|
| `EventID` | Event ID |
| `MessageID` | Reacted-to message |
| `ReactionType` | Emoji type such as `SMILE` |
| `OperatorType` | Raw operator kind |
| `UserID` | User or app identifier |
| `Action` | `added` or `removed` |
| `CreateTimeMs` | Event time |
| `RawEvent` | Original created/deleted event |

### `CommentEvent`

| Field | Meaning |
|---|---|
| `EventID` | Event ID |
| `CommentID` | Comment ID |
| `FileToken` / `FileType` | Commented document target |
| `Operator` | Open ID, user ID, and union ID when supplied |
| `ReplyID` | Reply ID for comment replies |
| `MentionedBot` | Whether the event marks the bot as mentioned |
| `Timestamp` | Event timestamp |
| `RawEvent` | Original custom event |

Document comment delivery requires the document event subscription, suitable
Drive permissions, and access to the target document.

### `BotAddedEvent`

| Field | Meaning |
|---|---|
| `EventID` | Event ID |
| `ChatID` / `ChatName` | Target chat |
| `UserID` | User who added the bot when supplied |
| `CreateTimeMs` | Event time |
| `RawEvent` | Original event |

### `CardActionEvent`

| Field | Meaning |
|---|---|
| `EventID` | Callback event ID |
| `MessageID` / `ChatID` | Source message and chat |
| `Operator` | User and tenant identifiers |
| `Action` | Tag, name, option, value, form values, input, options, and checked state |
| `Context` | Callback URL, preview token, open message ID, and open chat ID |
| `Token`, `Host`, `DeliveryType` | Callback metadata when supplied |
| `RawEvent` | Original card callback |

### `RejectEvent`

| Field | Meaning |
|---|---|
| `MessageID` | Rejected message |
| `ChatID` | Source conversation |
| `SenderID` | Sender identifier |
| `Reason` | Stable policy reason |

Reasons are `group_not_allowed`, `no_mention`, `mention_all_blocked`,
`dm_disabled`, and `sender_not_allowed`.

## Lifecycle Events

| Handler | Timing |
|---|---|
| `OnReady` | WebSocket became ready; bot identity resolution is attempted first |
| `OnError` | WebSocket or Channel handler reported an error |
| `OnReconnecting` | Reconnection began |
| `OnReconnected` | Reconnection succeeded |
| `OnDisconnected` | Connection closed |

`GetBotIdentity(ctx)` can also be called directly. It returns cached data when
fresh, stale cached data after a refresh failure, or `nil` when no identity has
ever been resolved.

## Related Documents

- [Quickstart](quickstart.md)
- [API reference](reference.md)
- [Policy and safety](policy-and-safety.md)
- [Troubleshooting events](troubleshooting.md#events-are-not-delivered)
