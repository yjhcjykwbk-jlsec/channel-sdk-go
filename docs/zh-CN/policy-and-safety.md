# 策略与安全

[English](../policy-and-safety.md) | [文档索引](README.md)

入站 runtime 在调用业务消息处理器前，会执行机器人自身消息过滤、过期事件
检查、去重、准入策略、处理锁以及按 chat 批处理。

## 默认准入策略

| 会话 | 默认行为 |
|---|---|
| 群聊 | 允许所有群 |
| 群聊 mention | 必须 mention 机器人 |
| 群聊 `@all` | 拒绝 |
| 私聊 | 允许 |

被拒绝的消息不会进入 `OnMessage`。注册 `OnReject` 可以观察策略结果。

## 构造时配置策略

指针布尔值用于区分显式 `false` 和“使用默认值”：

```go
func boolPtr(value bool) *bool {
	return &value
}

policy := channel.PolicyConfig{
	GroupAllowlist:      []string{"oc_allowed"},
	RequireMention:      boolPtr(true),
	RespondToMentionAll: boolPtr(false),
	DMMode:              "allowlist",
	DMAllowlist:         []string{"ou_allowed"},
}

ch, err := channel.New(
	appID,
	appSecret,
	channel.WithPolicyConfig(policy),
)
```

## `PolicyConfig`

| 字段 | 零值/默认行为 | 含义 |
|---|---|---|
| `GroupAllowlist` | 空表示允许所有群 | 允许的 chat ID |
| `RequireMention` | nil 表示 `true` | 是否必须 mention 当前机器人 |
| `RespondToMentionAll` | nil 表示 `false` | 是否允许包含 `@all` 的消息 |
| `DMMode` | 空表示 `"open"` | `"open"`、`"disabled"` 或 `"allowlist"` |
| `DMAllowlist` | 空 | DM 为 allowlist 时允许的发送者 ID |

`DMMode` 除 `disabled` 和 `allowlist` 之外的值目前表现为 open。请只使用文档
列出的值。

## 运行时更新策略

`UpdatePolicy` 执行局部更新：

```go
ch.UpdatePolicy(channel.PolicyConfig{
	RequireMention: boolPtr(false),
})
```

更新规则：

| 字段值 | 效果 |
|---|---|
| nil slice | 保留当前 allowlist |
| 非 nil 空 slice | 清空当前 allowlist |
| 非空 slice | 替换当前 allowlist |
| nil 布尔指针 | 保留当前值 |
| 非 nil 布尔指针 | 替换当前值 |
| 空 `DMMode` | 保留当前模式 |
| 非空 `DMMode` | 替换当前模式 |

`GetPolicy` 返回当前配置。方法内部使用锁，但返回的 slice 必须视为只读，不能
并发修改。

## 拒绝原因

| 原因 | 触发条件 |
|---|---|
| `group_not_allowed` | Chat 不在 `GroupAllowlist` 中 |
| `no_mention` | 群消息没有 mention 机器人，也没有 `@all` |
| `mention_all_blocked` | 消息包含 `@all` 且该行为未开启 |
| `dm_disabled` | 私聊已关闭 |
| `sender_not_allowed` | 发送者不在 `DMAllowlist` 中 |

## Safety 配置

修改单个字段前先复制默认配置：

```go
defaults := channeltypes.DefaultChannelConfig()
defaults.Safety.StaleMessageWindowMs = 10 * time.Minute
defaults.Safety.Batch.DelayMs = 300 * time.Millisecond
defaults.Safety.Batch.MaxMessages = 4

ch, err := channel.New(
	appID,
	appSecret,
	channel.WithSafetyConfig(defaults.Safety),
)
```

`WithSafetyConfig` 会替换完整配置。使用局部零值 literal 不安全：过期窗口为零会
丢弃所有带时间戳消息，批处理上限为零会让每条消息立即 flush。

## 过期事件

`CreateTimeMs` 为正且早于 `StaleMessageWindowMs` 的消息会被丢弃。没有可解析
时间戳的事件不视为过期。默认窗口为 30 分钟。

## 去重与处理锁

SDK 使用内存 LRU 保存最近处理的 key：

- 消息：message ID；
- reaction：message、actor、reaction、action 和时间戳；
- 评论：file token 和 comment ID；
- 机器人入群：event ID；
- 卡片 action：event ID，缺失时使用稳定的 payload 派生值。

默认最多保留 10000 个 key，保留窗口为 1 小时。处理锁避免同一个 key 被并发
处理，内部生命周期为 5 分钟。

这些保护只在当前进程内有效，不能跨副本去重，也不会在重启后保留。需要持久化
exactly-once 副作用时，应用必须使用自己的幂等 key 和持久化唯一约束。

## 按 Chat 批处理

默认批处理配置：

| 字段 | 默认值 |
|---|---:|
| `DelayMs` | `600ms` |
| `LongThresholdChars` | `1000` |
| `LongDelayMs` | `2s` |
| `MaxMessages` | `8` |
| `MaxChars` | `4000` |

同一 chat 的消息可能在 debounce 窗口内合并。内容用空行连接，标量元数据来自
最后一条消息，resources/mentions 去重。同一 chat 内串行执行，不同 chat 可以
并发。

虽然字段名包含 `Chars`，当前实现通过 Go `len(string)` 计算
`LongThresholdChars` 和 `MaxChars`，因此实际阈值单位是 UTF-8 字节。

处理器应使用较短超时。慢处理器会阻塞同一 chat 的后续任务。

## 机器人身份缓存

SDK 请求 `/open-apis/bot/v3/info`，用于过滤自身消息和识别 mention。成功结果
缓存 30 分钟；刷新失败后至少等待 1 分钟再试；有旧缓存时返回旧值。

首次获取失败时 `GetBotIdentity` 返回 nil 而不是 error。业务依赖身份时，应关注
启动日志。

## 出站 URL 安全

URL 媒体使用独立 SSRF 检查和 body 大小限制，见[媒体](media.md)。准入策略不会
授权媒体 host，媒体 allowlist 也不会授权 chat 或发送者。

## 相关文档

- [接收事件](events.md)
- [API 默认值](reference.md)
- [媒体安全](media.md)
- [被过滤事件排查](troubleshooting.md)
