# Channel API Reference

[中文](zh-CN/reference.md) | [Documentation index](README.md)

This reference describes the stable application surface of
`github.com/larksuite/channel-sdk-go`. Use the [quickstart](quickstart.md) for
first-run application setup.

## Entry Point

```go
import channel "github.com/larksuite/channel-sdk-go"

ch, err := channel.New(appID, appSecret, opts...)
```

`appID` must be non-empty. `appSecret` must be non-empty unless a non-nil
`WithClientAssertionProvider` is configured. `New` creates the REST client,
event dispatcher, WebSocket client, and Channel runtime.

The returned `channel.Channel` interface is the public high-level API. It also
exposes `RawClient() *lark.Client` for OpenAPI operations not covered by the
Channel helpers.

## Constructor Options

| Option | Value | Applies to | Behavior |
|---|---|---|---|
| `WithLogger` | `larkcore.Logger` | REST, WebSocket, dispatcher, Channel | Uses the supplied logger |
| `WithLogLevel` | `larkcore.LogLevel` | REST, WebSocket, dispatcher | Sets the lower-level SDK log level |
| `WithDomain` | `string` | REST and WebSocket | Sets both OpenAPI and WebSocket domains |
| `WithOpenBaseURL` | `string` | REST | Overrides only the OpenAPI base URL |
| `WithWebSocketDomain` | `string` | WebSocket | Overrides only the long-connection domain |
| `WithOAuthBaseURL` | `string` | REST token flow | Overrides the OAuth/token base URL |
| `WithHTTPClient` | `larkcore.HttpClient` | REST | Supplies a custom REST HTTP client |
| `WithReqTimeout` | `time.Duration` | REST | Sets the lower-level REST request timeout |
| `WithTokenCache` | `larkcore.Cache` | REST token flow | Supplies a tenant-token cache |
| `WithEnableTokenCache` | `bool` | REST token flow | Enables or disables lower-level token caching |
| `WithHeaders` | `http.Header` | REST and WebSocket | Adds shared request headers |
| `WithSource` | `string` | REST and WebSocket | Sets the SDK source identifier |
| `WithClientAssertionProvider` | `larkcore.ClientAssertionProvider` | REST and WebSocket auth | Uses client assertion authentication and permits an empty `appSecret` |
| `WithSafetyConfig` | `channel.SafetyConfig` | Inbound runtime | Replaces the complete safety configuration |
| `WithPolicyConfig` | `channel.PolicyConfig` | Inbound runtime | Sets message admission policy |
| `WithOutboundConfig` | `channel.OutboundConfig` | Sending and streaming | Replaces the complete outbound configuration |
| `WithBotIdentityCacheConfig` | `channel.BotIdentityCacheConfig` | Bot identity | Replaces bot identity cache settings |

`WithHTTPClient` does not replace WebSocket dialing. Use
`WithWebSocketDomain`, or lower-level WebSocket facilities from
`oapi-sdk-go/v3`, when transport control beyond the Channel API is required.

## Preserve Defaults When Customizing

`WithSafetyConfig` and `WithOutboundConfig` replace their complete nested
configuration. Start from `types.DefaultChannelConfig()` when changing one
field:

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

Do not build a partial `SafetyConfig` from its zero value. In particular, a
zero stale-message window treats timestamped messages as stale, and zero batch
limits cause immediate flushing. A zero-value `OutboundConfig` also disables
text chunking and uses the stream controller's internal fallback throttle
instead of the normal 100 ms default.

## Effective Defaults

| Field | Default | Meaning |
|---|---:|---|
| `Safety.Dedup.MaxEntries` | `10000` | Maximum in-memory duplicate keys |
| `Safety.Dedup.SweepIntervalMs` | `1h` | Duplicate-key retention window |
| `Safety.StaleMessageWindowMs` | `30m` | Drop older timestamped messages |
| `Safety.Batch.DelayMs` | `600ms` | Normal debounce window |
| `Safety.Batch.LongThresholdChars` | `1000` | Switch to the long-message delay at this buffered byte size |
| `Safety.Batch.LongDelayMs` | `2s` | Debounce window for a long buffered message |
| `Safety.Batch.MaxMessages` | `8` | Flush when the batch reaches this count |
| `Safety.Batch.MaxChars` | `4000` | Flush when buffered content reaches this byte size |
| `Policy.GroupAllowlist` | empty | Allow all groups |
| `Policy.RequireMention` | `true` | Require bot mention in groups |
| `Policy.RespondToMentionAll` | `false` | Reject `@all` group messages |
| `Policy.DMMode` | `"open"` | Allow private messages |
| `Outbound.TextChunkLimit` | `3500` | Plain text uses runes; Markdown uses a UTF-8 byte threshold at line boundaries |
| `Outbound.StreamThrottleMs` | `100ms` | Minimum interval between Markdown stream updates |
| `Outbound.StreamThrottleChars` | `10` | Reserved in the current release; it does not change runtime behavior |
| `Outbound.MediaSource.MaxDownloadBytes` | `100 MiB` | Maximum body read from a URL media source |
| `Outbound.Retry.MaxAttempts` | `3` | Total send/update attempts |
| `Outbound.Retry.BaseDelayMs` | `500ms` | Initial retry backoff |
| `BotIdentityCache.TTL` | `30m` | Successful identity cache lifetime |
| `BotIdentityCache.MinRefreshInterval` | `1m` | Minimum interval after an identity refresh failure |

`BotIdentityCache.MinRefreshInterval` values between zero and 30 seconds are
raised to 30 seconds. Non-positive media limits fall back to 100 MiB. Retry
attempts and base delay also fall back to 3 and 500 ms when non-positive.

## Channel Methods

| Method | Requires `Start` | Description |
|---|---:|---|
| `Send(ctx, input)` | no | Send or reply to one logical message |
| `Stream(ctx, input)` | no | Send an initial message and return a stream controller |
| `DownloadFile(ctx, fileKey, mediaType)` | no | Download an image, file, audio, or video into memory |
| `OnMessage(handler)` | for delivery | Register a normalized message handler |
| `OnReaction(handler)` | for delivery | Register reaction-created and reaction-deleted handlers |
| `OnComment(handler)` | for delivery | Register a document-comment handler |
| `OnBotAdded(handler)` | for delivery | Register a bot-added-to-chat handler |
| `OnCardAction(handler)` | for delivery | Register a long-connection card action handler |
| `OnReject(handler)` | for delivery | Register a local policy-rejection handler |
| `OnReady(handler)` | yes | Run after the WebSocket connection becomes ready |
| `OnError(handler)` | no | Observe WebSocket and handler errors |
| `OnReconnecting(handler)` | yes | Observe reconnection start |
| `OnReconnected(handler)` | yes | Observe successful reconnection |
| `OnDisconnected(handler)` | yes | Observe disconnection |
| `Start(ctx)` | n/a | Start and block on the WebSocket long connection |
| `Stop(ctx)` | no | Close WebSocket and dispose local pipelines; idempotent |
| `UpdatePolicy(cfg)` | no | Apply a partial runtime policy update |
| `GetPolicy()` | no | Return the current policy configuration |
| `GetBotIdentity(ctx)` | no | Fetch/cache bot identity; returns `nil` on an unrecoverable fetch failure |
| `RawClient()` | no | Return the lower-level `*lark.Client` |

Register inbound and lifecycle handlers before `Start`. Treat a Channel as
single-lifecycle: after `Stop`, create a new Channel rather than restarting the
old one.

## `SendInput`

Target selection:

| Field | Priority | Behavior |
|---|---:|---|
| `ReceiveID` | 1 | Preferred target; type inferred from prefix or email syntax |
| `ChatID` | 2 | Sent with `receive_id_type=chat_id` |
| `UserID` | 3 | Legacy fallback sent with `receive_id_type=open_id` |
| `ReplyMessageID` | separate | Uses the reply API, but a target field is still required for fallback |

Content selection:

| Field | Resulting type | Notes |
|---|---|---|
| `Text` | `text` | Chunked; mentions are prepended |
| `Markdown` | `post` | Rendered in a `tag=md` post; chunked |
| `Title` | n/a | Post title for Markdown and Markdown streaming |
| `ImageKey` | `image` | Existing uploaded image |
| `FileKey` | `file` | Existing uploaded file |
| `AudioKey` | `audio` | Existing uploaded audio |
| `VideoKey` | `media` | Existing uploaded video |
| `Card` | `interactive` | Stringified card JSON |
| `Post` | `post` | Stringified Feishu/Lark post JSON |
| `ShareChatID` | `share_chat` | Chat business-card target |
| `ShareUserID` | `share_user` | User business-card target |
| `StickerFileKey` | `sticker` | Existing sticker file key |
| `ImagePath` | `image` | Uploads a local image when `ImageKey` is empty |
| `FilePath` | `file` | Uploads a local file when `FileKey` is empty |
| `Media` | based on `Media.Kind` | Upload from bytes, local path, or URL |
| `Mentions` | n/a | Prepended to text or the first Markdown chunk |
| `MsgType` | compatibility hint | Convenience content fields determine the actual message type |

Provide one logical content shape. If multiple shapes are set, the SDK uses the
precedence documented in [Sending messages](sending-messages.md#content-precedence).

## `SendResult`

| Field | Description |
|---|---|
| `MessageID` | First message ID for the logical send |
| `ChunkIDs` | All message IDs when long text/Markdown was split; omitted for one chunk |
| `ChatID` | Chat ID returned by the send API |
| `Error` | Reserved result field; current `Send` failures are returned as the Go `error` |

## `NormalizedMessage`

| Field | Description |
|---|---|
| `EventID` | Original event ID |
| `MessageID` | Message ID; for a merged batch, this comes from the last message |
| `ChatID` | Conversation ID |
| `ChatType` | Usually `group` or `p2p` |
| `UserID` | Sender open ID when available, otherwise sender user ID |
| `Content` | Flattened normalized content |
| `RawContentType` | Original message type |
| `Mentions` | Normalized mentions |
| `MentionAll` | Whether the message includes `@all` |
| `MentionedBot` | Whether the current bot was mentioned |
| `Resources` | Image, file, audio, video, sticker, and other resource descriptors |
| `CreateTimeMs` | Event creation time in Unix milliseconds |
| `RawEvent` | Original `P2MessageReceiveV1` payload |

See [Receiving events](events.md) for event payload tables and dispatch
semantics.

## `UploadInput`

| Field | Description |
|---|---|
| `Kind` | `MediaKindImage`, `MediaKindFile`, `MediaKindAudio`, or `MediaKindVideo` |
| `SourceBytes` | Highest-priority source when non-nil |
| `SourcePath` | Local path used when bytes are nil |
| `SourceURL` | HTTP(S) URL used when bytes and path are absent |
| `FileName` | Upload filename; a kind-specific default is used when empty |
| `Duration` | Audio/video duration in milliseconds; auto-detected for supported Opus/OGG and MP4 data when omitted |

See [Media](media.md) for source precedence and safety rules.

## Error Model

`Send` and streaming updates classify lower-level failures as
`*channel.FeishuChannelError`.

| Code | Meaning | Retryable by the built-in retry loop |
|---|---|---:|
| `target_revoked` | Reply target was recalled, unavailable, or mismatched | no; replies fall back to a new message |
| `permission_denied` | Credentials or permissions are insufficient | no |
| `format_error` | Message payload was rejected | no; rich content may fall back to text |
| `rate_limited` | Upstream rate limit | yes |
| `ssrf_blocked` | URL media source failed safety validation | no |
| `send_timeout` | Request timed out | no |
| `unknown` | Uncategorized transport or upstream error | yes |

Use `errors.As` or the helpers:

```go
var channelErr *channel.FeishuChannelError
if errors.As(err, &channelErr) {
	log.Printf("channel error code: %s", channelErr.Code)
}

if channel.IsRetryable(err) {
	// The SDK has already exhausted its configured retry loop.
}
```

The root package exposes `ClassifyError`, `IsRetryable`, and `IsFormatError`.
`types.IsReplyTargetGone` is available from the public `types` package.

## Related Documents

- [Receiving events](events.md)
- [Sending messages](sending-messages.md)
- [Streaming replies](streaming.md)
- [Media](media.md)
- [Policy and safety](policy-and-safety.md)
