# Migration from `oapi-sdk-go/v3/channel`

[中文](zh-CN/migration-from-oapi-sdk-go.md) | [Documentation index](README.md)

The standalone module keeps the high-level Channel models and methods while
moving application imports and construction to
`github.com/larksuite/channel-sdk-go`.

## Import Mapping

```diff
- "github.com/larksuite/oapi-sdk-go/v3/channel"
+ channel "github.com/larksuite/channel-sdk-go"

- "github.com/larksuite/oapi-sdk-go/v3/channel/types"
+ "github.com/larksuite/channel-sdk-go/types"
```

Most application-facing types are also re-exported by the root package, so
applications can often use `channel.SendInput`, `channel.NormalizedMessage`,
and `channel.PolicyConfig` without a separate `types` import.

## Constructor Change

Historical construction created lower-level clients first:

```go
client := lark.NewClient(appID, appSecret)
wsClient := larkws.NewClient(appID, appSecret, wsOptions...)
ch := oldchannel.NewChannel(client, wsClient, channelOptions...)
```

The standalone package owns those clients:

```go
ch, err := channel.New(
	appID,
	appSecret,
	channel.WithDomain("https://open.feishu.cn"),
	channel.WithPolicyConfig(policy),
)
if err != nil {
	return err
}
```

Credential validation now happens in `New`.

## Option Mapping

| Historical input | Standalone equivalent |
|---|---|
| `lark.WithOpenBaseUrl` | `channel.WithDomain` or `channel.WithOpenBaseURL` |
| `lark.WithOAuthBaseUrl` | `channel.WithOAuthBaseURL` |
| `lark.WithHttpClient` | `channel.WithHTTPClient` |
| `lark.WithReqTimeout` | `channel.WithReqTimeout` |
| `lark.WithTokenCache` | `channel.WithTokenCache` |
| `lark.WithEnableTokenCache` | `channel.WithEnableTokenCache` |
| `larkws.WithDomain` | `channel.WithDomain` or `channel.WithWebSocketDomain` |
| lower-level logger/log level options | `channel.WithLogger` / `channel.WithLogLevel` |
| `types.WithSafetyConfig` | `channel.WithSafetyConfig` |
| `types.WithPolicyConfig` | `channel.WithPolicyConfig` |
| `types.WithOutboundConfig` | `channel.WithOutboundConfig` |
| `types.WithBotIdentityCacheConfig` | `channel.WithBotIdentityCacheConfig` |

`WithDomain` configures both REST and WebSocket domains. For private or split
environments, configure `WithOpenBaseURL`, `WithWebSocketDomain`, and
`WithOAuthBaseURL` explicitly.

## Public Method Compatibility

The main workflow remains:

```go
ch.OnMessage(handler)
ch.OnReaction(reactionHandler)
ch.OnCardAction(cardHandler)

if err := ch.Start(ctx); err != nil {
	return err
}
```

The standalone `Channel` retains `Send`, `Stream`, event registrations,
lifecycle hooks, policy access, media download, bot identity, `Start`, and
`Stop`.

## Lower-Level Client Access

Use `ch.RawClient()` for occasional OpenAPI calls outside the high-level
surface:

```go
raw := ch.RawClient()
```

The standalone constructor does not expose `NewWithClient`, `NewWithClients`,
or `NewChannel` injection variants. Keep using the main `oapi-sdk-go/v3`
clients directly when the application requires a fully custom generated
client, per-request access tokens, or complete WebSocket transport ownership.

Both modules can be present in one `go.mod`: use the standalone module for
Channel workflows and the main SDK for the broader OpenAPI surface.

## Configuration Review

When migrating:

1. Start from `types.DefaultChannelConfig()` before modifying safety or outbound
   fields.
2. Confirm REST, OAuth, and WebSocket domains separately in custom
   environments.
3. Preserve group mention and DM policy intentionally.
4. Keep event handlers registered before `Start`.
5. Add a deadline to each REST call context.
6. Verify app permissions and reinstall the app.
7. Run unit, race, example, and E2E dry-run checks.

```bash
go test ./...
go test -race ./...
go test ./examples/...
CHANNEL_E2E_DRY_RUN=1 go test -tags=e2e ./e2e -run TestChannelE2E -v
```

## Behavior Worth Rechecking

- The default group policy requires a bot mention and rejects `@all`.
- Inbound messages can be batched inside a 600 ms window.
- `ReceiveID` is preferred because it infers the target ID type.
- Replies require a target field for fallback-as-new behavior.
- `Stop` is idempotent, but a stopped Channel should not be restarted.
- Inbound transport is WebSocket-only in the current release.

## Related Documents

- [Quickstart](quickstart.md)
- [API reference](reference.md)
- [Troubleshooting](troubleshooting.md)
