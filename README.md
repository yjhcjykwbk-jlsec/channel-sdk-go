# Lark Channel SDK for Go

[中文文档](README.zh.md)

`channel-sdk-go` is a Go package for building Feishu and Lark conversational
bots. It provides one high-level Channel entry point for WebSocket event
listening, normalized inbound events, outbound messaging, media handling,
interactive card callbacks, policy controls, and streaming replies.

Requires Go 1.18 or later.

## Install

```bash
go get github.com/larksuite/channel-sdk-go
```

## Minimal Example

```go
package main

import (
	"context"
	"log"
	"os"

	channel "github.com/larksuite/channel-sdk-go"
)

func main() {
	ch, err := channel.New(os.Getenv("APP_ID"), os.Getenv("APP_SECRET"))
	if err != nil {
		log.Fatalf("create channel: %v", err)
	}

	ch.OnMessage(func(ctx context.Context, msg *channel.NormalizedMessage) error {
		_, err := ch.Send(ctx, &channel.SendInput{
			ReceiveID: msg.ChatID,
			Text:      "received: " + msg.Content,
		})
		return err
	})

	if err := ch.Start(context.Background()); err != nil {
		log.Fatalf("start channel: %v", err)
	}
}
```

`Start(ctx)` opens the WebSocket connection and blocks until the connection
ends. `Send` and `Stream` use REST APIs and can be called without `Start`.

## Documentation

| Topic | Description |
|---|---|
| [Documentation index](docs/README.md) | All user guides and examples |
| [Quickstart](docs/quickstart.md) | Prepare an app, configure events, and run an echo bot |
| [API reference](docs/reference.md) | Constructor options, defaults, methods, models, and errors |
| [Receiving events](docs/events.md) | Message normalization, callbacks, ordering, and concurrency |
| [Sending messages](docs/sending-messages.md) | `SendInput` parameters, target detection, message types, and replies |
| [Streaming replies](docs/streaming.md) | Markdown and interactive-card stream controllers |
| [Media](docs/media.md) | Upload from keys, files, bytes, or URLs and download resources |
| [Policy and safety](docs/policy-and-safety.md) | Admission policy, batching, deduplication, and stale-event handling |
| [Migration guide](docs/migration-from-oapi-sdk-go.md) | Move from `oapi-sdk-go/v3/channel` |
| [Troubleshooting](docs/troubleshooting.md) | Events, permissions, sending, media, and streaming diagnostics |
| [E2E testing](e2e/README.md) | Run real Feishu/Lark end-to-end tests |

## Examples

- [Send a message](examples/send_message)
- [Echo bot](examples/echo_bot)
- [Streaming reply](examples/stream_reply)
- [Card action callback](examples/card_action)
- [Custom domains](examples/custom_domain)

## Package Boundary

Application code should normally import only the root package:

```go
import channel "github.com/larksuite/channel-sdk-go"
```

The public [`types`](types) package is available when explicit type imports or
default configuration values are useful. Packages under `internal` are
implementation details and are not part of the compatibility contract.

This release receives events and callbacks through WebSocket long connections.
It does not expose an HTTP webhook adapter. Use `RawClient()` or the main
`oapi-sdk-go/v3` module when an integration needs OpenAPI capabilities outside
the Channel surface.

## Local Development

```bash
go test ./...
go test -race ./...
go test ./examples/...
```

Real E2E tests use the `e2e` build tag and are not included in normal unit
tests:

```bash
CHANNEL_E2E_DRY_RUN=1 go test -tags=e2e ./e2e -run TestChannelE2E -v
```

## Contributing

Please read [CONTRIBUTING.md](CONTRIBUTING.md) before submitting a pull
request. All contributors must follow our
[Code of Conduct](CODE_OF_CONDUCT.md).

## Security

Report potential vulnerabilities according to [SECURITY.md](SECURITY.md).
Please do not disclose security vulnerabilities through public GitHub issues.

## License

This project is licensed under the [MIT License](LICENSE).
