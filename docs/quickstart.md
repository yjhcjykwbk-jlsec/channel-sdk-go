# Channel Quickstart

[中文](zh-CN/quickstart.md) | [Documentation index](README.md)

This guide prepares a self-built Feishu or Lark bot, starts the WebSocket long
connection, and sends a reply to an incoming message.

## 1. Prepare the App

In the Feishu or Lark developer console:

1. Create a self-built app and enable the bot capability.
2. Grant message permissions required by the bot. Common permissions include
   `im:message` and `im:message:send_as_bot`; enable additional permissions for
   reactions, chat membership, files, or document comments when those features
   are used.
3. Under event and callback settings, select long connection/WebSocket mode.
4. Subscribe to the receive-message event (`im.message.receive_v1`).
5. Publish or install the latest app version into the test tenant. Reinstall or
   republish after changing permissions or subscriptions.
6. Add the bot to a test chat, or open a private conversation with it.

For group messages, the default policy requires the bot to be mentioned. See
[Policy and safety](policy-and-safety.md) before changing that behavior.

## 2. Install the Module

```bash
go get github.com/larksuite/channel-sdk-go
```

Set credentials in the process environment:

```bash
export APP_ID=cli_xxx
export APP_SECRET=your_app_secret
```

The examples read environment variables directly. They do not load `.env`
automatically. The repository-level `.env` loader is only used by the E2E
runner.

For a Lark tenant, set the Lark OpenAPI domain:

```go
ch, err := channel.New(
	appID,
	appSecret,
	channel.WithDomain("https://open.larksuite.com"),
)
```

The SDK derives the standard Lark OAuth domain from this known OpenAPI domain.
Custom and private OpenAPI domains must normally provide
`WithOAuthBaseURL(...)` separately.

## 3. Run an Echo Bot

The runnable source is in [`examples/echo_bot`](../examples/echo_bot).

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

	ch.OnError(func(err error) {
		log.Printf("channel error: %v", err)
	})

	ch.OnMessage(func(ctx context.Context, msg *channel.NormalizedMessage) error {
		_, err := ch.Send(ctx, &channel.SendInput{
			ReceiveID: msg.ChatID,
			Text:      "echo: " + msg.Content,
		})
		return err
	})

	if err := ch.Start(context.Background()); err != nil {
		log.Fatalf("start channel: %v", err)
	}
}
```

Run it from the repository:

```bash
go run ./examples/echo_bot
```

`Start(ctx)` registers the configured handlers, opens the WebSocket connection,
and blocks. Register event and lifecycle handlers before calling it.

## 4. Send Without WebSocket

REST operations do not require `Start`:

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

result, err := ch.Send(ctx, &channel.SendInput{
	ReceiveID: "oc_xxx",
	Text:      "hello from channel-sdk-go",
})
if err != nil {
	return err
}
log.Printf("sent message %s", result.MessageID)
```

Call `Stop(ctx)` when the application is done with the Channel. `Stop` is
idempotent, but a stopped Channel should not be restarted; construct a new one
for a new lifecycle.

## 5. Next Steps

- [Receive and normalize events](events.md)
- [Send every supported message shape](sending-messages.md)
- [Build a streaming reply](streaming.md)
- [Review every constructor option and default](reference.md)
- [Diagnose common setup and permission failures](troubleshooting.md)
