# Channel SDK Go E2E 测试

本目录放真实飞书 / Lark 环境的端到端测试入口。测试文件带 `e2e`
build tag，普通 `go test ./...` 不会触发真实网络调用。

## 前置条件

在仓库根目录创建 `.env`，可以从 `.env.example` 复制。真实密钥只放在
`.env`，不要提交。

最小必填项：

- `APP_ID`
- `APP_SECRET`
- `CHANNEL_E2E_RECEIVE_OPEN_ID`

推荐补齐项：

- `CHANNEL_E2E_GROUP_CHAT_ID`：群聊发送、mention、策略用例。
- `CHANNEL_E2E_MENTION_USER_OPEN_ID`：mention 目标用户。
- `CHANNEL_E2E_DOC_TOKEN`：文档评论事件。
- `CHANNEL_E2E_SHARE_CHAT_ID`：群名片消息。
- `CHANNEL_E2E_ALLOWED_GROUP_CHAT_ID`、`CHANNEL_E2E_BLOCKED_GROUP_CHAT_ID`、`CHANNEL_E2E_ALLOWED_USER_OPEN_ID`：策略用例。
- `CHANNEL_E2E_BLOCKED_USER_OPEN_ID`：可选；为空时只跳过 blocked-user 策略用例。
- `CHANNEL_E2E_STICKER_FILE_KEY`：可选；为空时跳过贴纸消息用例。

媒体素材默认指向 `testdata/e2e`。如果线上平台对音频或视频格式做更严格校验，
可以把 `.env` 中的音视频路径改成真实测试素材。

## 运行命令

只校验配置、素材路径和用例计划，不调用飞书接口：

```bash
CHANNEL_E2E_DRY_RUN=1 go test -tags=e2e ./e2e -run TestChannelE2E -v
```

运行自动发送类用例：

```bash
go test -tags=e2e ./e2e -run TestChannelE2E -v -timeout 10m
```

运行需要人工触发的事件类用例：

```bash
CHANNEL_E2E_MANUAL=1 CHANNEL_E2E_SKIP_AUTO=1 go test -tags=e2e ./e2e -run TestChannelE2E -v -timeout 15m
```

运行策略用例：

```bash
CHANNEL_E2E_MANUAL=1 CHANNEL_E2E_SKIP_AUTO=1 CHANNEL_E2E_ENABLE_POLICY=1 go test -tags=e2e ./e2e -run TestChannelE2E -v -timeout 15m
```

机器人入群事件会改变群成员关系，默认不跑。需要覆盖时显式开启：

```bash
CHANNEL_E2E_MANUAL=1 CHANNEL_E2E_SKIP_AUTO=1 CHANNEL_E2E_ENABLE_BOT_ADDED=1 go test -tags=e2e ./e2e -run TestChannelE2E -v -timeout 15m
```

## 人工触发项

`CHANNEL_E2E_MANUAL=1` 时，runner 会先启动长连接并发送必要的目标消息或卡片，
然后在测试日志中输出要执行的动作：

- 给机器人发送日志里指定的 trace 文本。
- 对 runner 发出的 reaction target 消息添加并删除 reaction。
- 点击 runner 发出的 `E2E OK` 卡片按钮。
- 在 `.env` 指定文档里添加测试评论，按用例需要 mention 机器人。
- 策略用例开启时，分别在 allowed / blocked 群或用户会话里发送测试消息。

runner 只记录用例 ID、跳过原因、trace 和脱敏后的文档 token，不打印
`APP_SECRET` 或 access token。

## 端点配置

默认飞书环境可以不配端点。自定义环境可按需配置：

- `CHANNEL_E2E_DOMAIN`：兼容入口，同时设置 REST 和 WebSocket 域名。
- `CHANNEL_E2E_OPEN_BASE_URL`：只设置 REST OpenAPI base URL。
- `CHANNEL_E2E_WS_BASE_URL`：只设置长连接域名。

如果同时配置，`OPEN_BASE_URL` 和 `WS_BASE_URL` 会分别覆盖 REST 与长连接端点。
