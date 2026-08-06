# Troubleshooting

[中文](zh-CN/troubleshooting.md) | [Documentation index](README.md)

Start with a bounded context, an `OnError` handler, and logs that include event
or message IDs but never credentials, access tokens, full document content, or
raw sensitive payloads.

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

## `New` Fails

| Symptom | Check |
|---|---|
| `appID is required` | `APP_ID` is set in the process environment |
| `appSecret is required...` | `APP_SECRET` is set, or a non-nil client assertion provider is configured |
| Credentials appear set but are empty | The application does not automatically load `.env` |

Do not print the secret to verify it. Check only whether the environment
variable exists and has a non-zero length.

## WebSocket Does Not Become Ready

Check:

1. The app is a supported self-built bot application.
2. Long connection/WebSocket mode is enabled under events and callbacks.
3. The latest app version is published or installed.
4. REST and WebSocket domains target the same environment.
5. A custom OpenAPI domain also has the correct OAuth base URL.
6. The process can reach the configured endpoints.
7. The `Start` context has not already been canceled.

Register `OnReady`, `OnReconnecting`, `OnReconnected`, `OnDisconnected`, and
`OnError` before `Start` to distinguish startup from reconnection failures.

## Events Are Not Delivered

Check:

1. The exact event or card callback is enabled in the developer console.
2. The application has the corresponding permission.
3. The bot/app was reinstalled after permission or subscription changes.
4. The bot is present in the test chat or can access the target document.
5. The handler was registered before `Start`.
6. The test action happened after `OnReady`.
7. The event belongs to the same app and tenant as `APP_ID`.

For message events, subscribe to `im.message.receive_v1`. Comment handling uses
the custom `drive.notice.comment_add_v1` event and also requires document access.

## Events Arrive but `OnMessage` Does Not Run

The runtime can intentionally filter a message before the handler:

| Filter | Default behavior |
|---|---|
| Self-message | Messages from the current bot are ignored |
| Stale event | Messages older than 30 minutes are ignored |
| Duplicate | Recently seen message IDs are ignored |
| Group allowlist | Non-allowlisted chats are rejected when configured |
| Group mention | The bot must be mentioned |
| `@all` | Rejected by default |
| DM policy | Open by default; configurable |
| Batch debounce | Delivery can wait 600 ms, or 2 seconds for long buffered content |

Register `OnReject` to see policy reasons. When using `WithSafetyConfig`, build
from `types.DefaultChannelConfig()`; a partial zero-value safety config can
drop every timestamped message.

## Sending Fails

First classify the error:

```go
var channelErr *channel.FeishuChannelError
if errors.As(err, &channelErr) {
	log.Printf("send failed: %s", channelErr.Code)
}
```

| Code | Typical cause | Action |
|---|---|---|
| `permission_denied` | Missing send permission, invalid credential, app not installed | Grant permission, publish, and reinstall |
| `format_error` | Invalid post/card JSON or unsupported message field combination | Validate payload and send one content shape |
| `target_revoked` | Reply target recalled or unavailable | SDK attempts fallback-as-new when a target field exists |
| `rate_limited` | Upstream quota | Reduce traffic; SDK already retries eligible failures |
| `send_timeout` | Context or transport deadline | Review timeout and downstream latency |
| `unknown` | Unclassified network/API failure | Inspect sanitized upstream code and message |

Also verify that one of `ReceiveID`, `ChatID`, or `UserID` is set. Replies still
need a target for fallback. Prefer `ReceiveID` to avoid ID-type ambiguity.

## The Wrong Message Type Is Sent

Convenience content fields, not `MsgType`, determine the result. If multiple
fields are set, key/media/card/post/share/sticker fields take precedence over
Markdown, which takes precedence over text.

Create a fresh `SendInput` for each logical send. The SDK can populate key
fields while uploading, and reply fallback can clear `ReplyMessageID`; do not
reuse one input concurrently.

See the complete [content precedence](sending-messages.md#content-precedence).

## Media Upload or Download Fails

| Symptom | Check |
|---|---|
| `ssrf_blocked` | URL scheme, DNS result, private/reserved address, and exact-host allowlist |
| Body exceeds limit | `Outbound.MediaSource.MaxDownloadBytes`; default is 100 MiB |
| Path traversal rejected | Remove `..` path segments and select a trusted path |
| Audio/video duration error | Use valid Opus/OGG or MP4, or provide milliseconds explicitly |
| Upload permission error | Image/file permissions and app reinstall |
| Unsupported download type | Use `image`, `file`, `audio`, `video`, or `media` |
| Memory pressure | URL typed media and downloads are materialized in memory |

For strict redirect or content-validation requirements, fetch through an
application-controlled client and pass validated `SourceBytes`.

## Streaming Fails

| Symptom | Cause |
|---|---|
| `Append is not supported...` | Controller was created from `Card` |
| `UpdateCard is not supported...` | Controller is a Markdown stream |
| `stream is closed` | `Close` already ran |
| Initial stream creation fails | The initial `Send` failed |
| Later update fails | Message update/patch permission, payload, timeout, or retry exhaustion |
| Updates seem delayed | 100 ms throttle/coalescing, or an explicit custom interval |

Call `Flush` or `Close` and handle the error. The card controller patches normal
interactive messages; it does not expose CardKit preallocation or sequence
APIs.

## Custom Domain Problems

`WithDomain` sets REST and WebSocket domains together. Split environments can
use:

```go
ch, err := channel.New(
	appID,
	appSecret,
	channel.WithOpenBaseURL(openBaseURL),
	channel.WithWebSocketDomain(wsDomain),
	channel.WithOAuthBaseURL(oauthBaseURL),
)
```

`WithHTTPClient` affects REST calls only. It does not customize WebSocket
dialing.

## Reproduce Safely

Run local checks:

```bash
go test ./...
go test -race ./...
go test ./examples/...
CHANNEL_E2E_DRY_RUN=1 go test -tags=e2e ./e2e -run TestChannelE2E -v
```

For real-environment coverage, follow the [E2E guide](../e2e/README.md). Keep
credentials only in the ignored root `.env`; never paste secrets into issues,
logs, screenshots, fixtures, or documentation.

## Related Documents

- [Quickstart](quickstart.md)
- [API reference](reference.md)
- [Receiving events](events.md)
- [Sending messages](sending-messages.md)
- [Media](media.md)
