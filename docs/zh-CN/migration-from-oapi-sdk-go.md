# 从 `oapi-sdk-go/v3/channel` 迁移

[English](../migration-from-oapi-sdk-go.md) | [文档索引](README.md)

独立模块保留高阶 Channel 模型和方法，同时将业务代码导入路径和构造入口迁移到
`github.com/larksuite/channel-sdk-go`。

## 导入路径映射

```diff
- "github.com/larksuite/oapi-sdk-go/v3/channel"
+ channel "github.com/larksuite/channel-sdk-go"

- "github.com/larksuite/oapi-sdk-go/v3/channel/types"
+ "github.com/larksuite/channel-sdk-go/types"
```

大部分业务类型也由根包重新导出，因此通常可以直接使用
`channel.SendInput`、`channel.NormalizedMessage` 和
`channel.PolicyConfig`，不必额外导入 `types`。

## 构造方式变化

历史方式需要先创建底层 Client：

```go
client := lark.NewClient(appID, appSecret)
wsClient := larkws.NewClient(appID, appSecret, wsOptions...)
ch := oldchannel.NewChannel(client, wsClient, channelOptions...)
```

独立包内部持有这些 Client：

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

现在由 `New` 校验凭证参数。

## Option 映射

| 历史输入 | 独立包对应项 |
|---|---|
| `lark.WithOpenBaseUrl` | `channel.WithDomain` 或 `channel.WithOpenBaseURL` |
| `lark.WithOAuthBaseUrl` | `channel.WithOAuthBaseURL` |
| `lark.WithHttpClient` | `channel.WithHTTPClient` |
| `lark.WithReqTimeout` | `channel.WithReqTimeout` |
| `lark.WithTokenCache` | `channel.WithTokenCache` |
| `lark.WithEnableTokenCache` | `channel.WithEnableTokenCache` |
| `larkws.WithDomain` | `channel.WithDomain` 或 `channel.WithWebSocketDomain` |
| 底层 logger/log level 选项 | `channel.WithLogger` / `channel.WithLogLevel` |
| `types.WithSafetyConfig` | `channel.WithSafetyConfig` |
| `types.WithPolicyConfig` | `channel.WithPolicyConfig` |
| `types.WithOutboundConfig` | `channel.WithOutboundConfig` |
| `types.WithBotIdentityCacheConfig` | `channel.WithBotIdentityCacheConfig` |

`WithDomain` 同时设置 REST 和 WebSocket 域名。私有环境或分离环境应显式配置
`WithOpenBaseURL`、`WithWebSocketDomain` 和 `WithOAuthBaseURL`。

## 公开方法兼容性

主要工作流保持不变：

```go
ch.OnMessage(handler)
ch.OnReaction(reactionHandler)
ch.OnCardAction(cardHandler)

if err := ch.Start(ctx); err != nil {
	return err
}
```

独立 `Channel` 保留 `Send`、`Stream`、事件注册、生命周期 hook、策略、媒体
下载、机器人身份、`Start` 和 `Stop`。

## 访问底层 Client

偶尔需要高阶 API 未覆盖的 OpenAPI 时，使用 `ch.RawClient()`：

```go
raw := ch.RawClient()
```

独立构造器不提供 `NewWithClient`、`NewWithClients` 或 `NewChannel` 注入入口。
如果应用需要完全自定义的生成 Client、每次请求手动传 access token，或完整
控制 WebSocket transport，应继续直接使用主 `oapi-sdk-go/v3`。

两个模块可以同时存在于一个 `go.mod`：Channel 流程使用独立模块，更广的
OpenAPI 能力使用主 SDK。

## 配置检查

迁移时：

1. 修改 safety 或 outbound 字段前，从 `types.DefaultChannelConfig()` 开始。
2. 自定义环境分别确认 REST、OAuth 和 WebSocket 域名。
3. 明确保留或修改群 mention 与私聊策略。
4. 在 `Start` 前注册事件处理器。
5. 每次 REST 调用使用带 deadline 的 context。
6. 确认应用权限并重新安装应用。
7. 运行单测、race、示例和 E2E dry-run。

```bash
go test ./...
go test -race ./...
go test ./examples/...
CHANNEL_E2E_DRY_RUN=1 go test -tags=e2e ./e2e -run TestChannelE2E -v
```

## 需要重新确认的行为

- 默认群策略要求 mention 机器人，并拒绝 `@all`。
- 入站消息可能在 600ms 窗口内合并。
- 推荐使用 `ReceiveID`，因为它会推导目标 ID 类型。
- 回复仍需要目标字段，用于目标失效时降级新发。
- `Stop` 可以重复调用，但已停止的 Channel 不应重启。
- 当前版本只通过 WebSocket 接收入站事件。

## 相关文档

- [快速开始](quickstart.md)
- [API 参考](reference.md)
- [问题排查](troubleshooting.md)
