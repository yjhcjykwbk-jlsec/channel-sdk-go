# Channel SDK for Go Documentation

[中文](zh-CN/README.md) | [Project README](../README.md)

Use this index to move from first setup to the detailed behavior of each public
Channel capability.

## Start Here

| Document | Use it when |
|---|---|
| [Quickstart](quickstart.md) | You are creating the first bot or validating app setup |
| [Migration from `oapi-sdk-go`](migration-from-oapi-sdk-go.md) | Existing code imports `oapi-sdk-go/v3/channel` |
| [Troubleshooting](troubleshooting.md) | A connection, event, send, upload, or stream does not behave as expected |

## Core Capabilities

| Document | Covers |
|---|---|
| [API reference](reference.md) | `New`, all constructor options, defaults, `Channel` methods, public models, and error codes |
| [Receiving events](events.md) | Message, reaction, comment, bot-added, card-action, reject, and lifecycle events |
| [Sending messages](sending-messages.md) | Text, Markdown, post, card, media, share, sticker, mention, and reply messages |
| [Streaming replies](streaming.md) | Incremental Markdown updates and interactive-card patches |
| [Media](media.md) | Existing media keys, automatic upload, URL safety, and resource download |
| [Policy and safety](policy-and-safety.md) | Group/DM admission, batching, deduplication, processing locks, and stale events |

## Runnable Examples

| Example | Purpose |
|---|---|
| [send_message](../examples/send_message) | Send with REST APIs without starting WebSocket |
| [echo_bot](../examples/echo_bot) | Receive and reply to normalized messages |
| [stream_reply](../examples/stream_reply) | Incrementally update a Markdown reply |
| [card_action](../examples/card_action) | Handle an interactive card action |
| [custom_domain](../examples/custom_domain) | Configure OpenAPI, OAuth, and WebSocket domains |

## Scope

The stable application entry point is the root
`github.com/larksuite/channel-sdk-go` package. The `types` package is public;
packages below `internal` are not.

This release supports WebSocket long connections for inbound events and
callbacks. It does not provide an HTTP webhook adapter. The SDK delegates
lower-level REST, WebSocket, event-dispatch, and generated OpenAPI models to
`github.com/larksuite/oapi-sdk-go/v3`.

## Test Documentation Changes

```bash
go test ./...
go test -race ./...
go test ./examples/...
CHANNEL_E2E_DRY_RUN=1 go test -tags=e2e ./e2e -run TestChannelE2E -v
```

Real-environment setup is documented in the [E2E guide](../e2e/README.md).
