# Policy and Safety

[中文](zh-CN/policy-and-safety.md) | [Documentation index](README.md)

The inbound runtime applies bot self-filtering, stale-event checks,
deduplication, admission policy, processing locks, and per-chat batching before
calling application message handlers.

## Default Admission Policy

| Conversation | Default |
|---|---|
| Group | All groups are allowed |
| Group mention | The bot must be mentioned |
| Group `@all` | Rejected |
| Direct message | Allowed |

Rejected messages do not reach `OnMessage`. Register `OnReject` to observe the
decision.

## Configure Policy at Construction

Pointer booleans distinguish an explicit `false` from "use the default":

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

| Field | Zero/default behavior | Meaning |
|---|---|---|
| `GroupAllowlist` | Empty allows all groups | Allowed chat IDs |
| `RequireMention` | nil means `true` | Require the current bot to be mentioned |
| `RespondToMentionAll` | nil means `false` | Permit messages containing `@all` |
| `DMMode` | Empty means `"open"` | `"open"`, `"disabled"`, or `"allowlist"` |
| `DMAllowlist` | Empty | Allowed sender IDs when DM mode is `allowlist` |

`DMMode` values other than `disabled` and `allowlist` currently behave as open.
Use only the documented values.

## Runtime Policy Updates

`UpdatePolicy` applies a partial update:

```go
ch.UpdatePolicy(channel.PolicyConfig{
	RequireMention: boolPtr(false),
})
```

Update rules:

| Field value | Effect |
|---|---|
| nil slice | Keep the current allowlist |
| non-nil empty slice | Clear the current allowlist |
| non-empty slice | Replace the current allowlist |
| nil boolean pointer | Keep the current value |
| non-nil boolean pointer | Replace the current value |
| empty `DMMode` | Keep the current mode |
| non-empty `DMMode` | Replace the current mode |

`GetPolicy` returns the current configuration. Calls use an internal lock, but
the returned slices must be treated as read-only and must not be mutated
concurrently.

## Rejection Reasons

| Reason | Trigger |
|---|---|
| `group_not_allowed` | Chat is outside `GroupAllowlist` |
| `no_mention` | Group message did not mention the bot or `@all` |
| `mention_all_blocked` | Message contains `@all` and that behavior is disabled |
| `dm_disabled` | Direct messages are disabled |
| `sender_not_allowed` | Sender is outside `DMAllowlist` |

## Safety Configuration

Start from the defaults before changing one field:

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

`WithSafetyConfig` replaces the whole safety configuration. A partial zero-value
literal is unsafe: a zero stale window drops timestamped messages and zero
batch limits make each message flush immediately.

## Stale Events

Messages with a positive `CreateTimeMs` older than
`StaleMessageWindowMs` are dropped. Events without a parseable timestamp are
not considered stale. The default window is 30 minutes.

## Deduplication and Processing Locks

The SDK keeps an in-memory LRU set of recently seen keys:

- message: message ID;
- reaction: message, actor, reaction, action, and timestamp;
- comment: file token and comment ID;
- bot-added: event ID;
- card action: event ID, with a stable payload-derived fallback.

The default maximum is 10,000 keys and the retention window is one hour.
Processing locks prevent the same key from being handled concurrently and use
a five-minute internal lifetime.

This protection is process-local. It does not deduplicate across replicas or
survive restart. Applications that require durable exactly-once side effects
must use their own idempotency key and persistent uniqueness constraint.

## Per-Chat Batching

The default batch settings are:

| Field | Default |
|---|---:|
| `DelayMs` | `600ms` |
| `LongThresholdChars` | `1000` |
| `LongDelayMs` | `2s` |
| `MaxMessages` | `8` |
| `MaxChars` | `4000` |

Messages in the same chat can be merged during the debounce window. Content is
joined with a blank line, the last message supplies scalar metadata, and
resources/mentions are deduplicated. Work for one chat is serialized; different
chats can execute concurrently.

Despite the field names, the current batching implementation measures
`LongThresholdChars` and `MaxChars` with Go's `len(string)`, so these thresholds
are UTF-8 byte counts.

Use short handler timeouts. A slow handler blocks later work in the same chat
scope.

## Bot Identity Cache

The SDK fetches `/open-apis/bot/v3/info` to filter self messages and recognize
mentions. Successful results are cached for 30 minutes. After a refresh
failure, another refresh is delayed for at least one minute; stale cached data
is returned when available.

`GetBotIdentity` returns `nil` instead of an error when the first fetch fails,
so observe startup logs when identity is required by the application.

## Outbound URL Safety

URL media has a separate SSRF guard and body-size limit. See
[Media](media.md#url-safety). Admission policy does not authorize media hosts,
and the media allowlist does not authorize chats or senders.

## Related Documents

- [Receiving events](events.md)
- [API defaults](reference.md#effective-defaults)
- [Media safety](media.md#url-safety)
- [Troubleshooting filtered events](troubleshooting.md#events-arrive-but-onmessage-does-not-run)
