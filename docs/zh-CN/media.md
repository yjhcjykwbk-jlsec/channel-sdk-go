# 媒体上传与下载

[English](../media.md) | [文档索引](README.md)

Channel SDK 可以发送已有媒体 key，上传本地图片/文件路径，也可以从字节、本地
路径或 HTTP(S) URL 上传图片、文件、音频和视频。

## 选择最简单的输入

| 输入 | 适用场景 |
|---|---|
| `ImageKey`、`FileKey`、`AudioKey`、`VideoKey` | 媒体已经上传 |
| `ImagePath` | 将本地图片作为消息图片上传 |
| `FilePath` | 将本地文件作为通用 stream 文件上传 |
| `Media` | 需要字节、URL、音视频时长处理或显式媒体类型 |

已有 key 可以避免重复上传，并且优先于自动 `Media` 上传。

## 本地图片或文件

```go
_, err := ch.Send(ctx, &channel.SendInput{
	ReceiveID: "oc_xxx",
	ImagePath: "testdata/e2e/image.png",
})

_, err = ch.Send(ctx, &channel.SendInput{
	ReceiveID: "oc_xxx",
	FilePath:  "testdata/e2e/document.pdf",
})
```

设置 `ImageKey` 后忽略 `ImagePath`；设置 `FileKey` 后忽略 `FilePath`。

## 类型化媒体上传

从字节上传：

```go
_, err := ch.Send(ctx, &channel.SendInput{
	ReceiveID: "oc_xxx",
	Media: &channel.UploadInput{
		Kind:        channel.MediaKindImage,
		SourceBytes: imageBytes,
		FileName:   "result.png",
	},
})
```

从本地文件上传：

```go
_, err := ch.Send(ctx, &channel.SendInput{
	ReceiveID: "oc_xxx",
	Media: &channel.UploadInput{
		Kind:       channel.MediaKindVideo,
		SourcePath: "testdata/e2e/video.mp4",
		FileName:  "demo.mp4",
	},
})
```

从 URL 上传：

```go
_, err := ch.Send(ctx, &channel.SendInput{
	ReceiveID: "oc_xxx",
	Media: &channel.UploadInput{
		Kind:      channel.MediaKindImage,
		SourceURL: "https://cdn.example.com/image.png",
		FileName: "image.png",
	},
})
```

## 来源优先级

同时设置多个 `UploadInput` 来源时：

1. 非 nil 的 `SourceBytes`；
2. `SourcePath`；
3. `SourceURL`。

长度为零但非 nil 的 `SourceBytes` 仍会被选中，并可能在上传时失败。每次只应
提供一个来源。

## 媒体类型

| 类型 | 上传行为 | 默认文件名 |
|---|---|---|
| `MediaKindImage` | Image API，`image_type=message` | `image.png` |
| `MediaKindFile` | File API，`file_type=stream` | `upload.bin` |
| `MediaKindAudio` | File API，`file_type=opus` | `voice.opus` |
| `MediaKindVideo` | File API，`file_type=mp4` | `video.mp4` |

`Duration` 单位是毫秒。音频或视频的 `Duration` 为 nil 或非正数时，SDK 会尝试
从 Opus/OGG 或 MP4 数据解析时长。如果输入能被平台接受、但无法被这两个解析器
识别，应显式传入时长。

## URL 安全

获取 `SourceURL` 前，SDK 会：

- 只允许 `http` 和 `https`；
- 解析初始 hostname；
- 拒绝 loopback、私有、link-local、multicast、reserved 等非公网 IP；
- 默认最多读取 100 MiB body；
- 使用 15 秒 HTTP Client 超时；
- 使用调用方 context 取消请求。

从默认值开始配置精确 hostname allowlist 和下载限制：

```go
defaults := channeltypes.DefaultChannelConfig()
defaults.Outbound.MediaSource.URLAllowlist = []string{"media.internal.example"}
defaults.Outbound.MediaSource.MaxDownloadBytes = 20 * 1024 * 1024

ch, err := channel.New(
	appID,
	appSecret,
	channel.WithOutboundConfig(defaults.Outbound),
)
```

Allowlist 中的 hostname 会绕过公网 IP 拒绝，因此它是安全边界，只能加入受控
域名。URL 重定向由 Go 标准 HTTP Client 处理；对重定向策略有更严格要求时，
应由应用自行下载和校验，再传入 `SourceBytes`。

## 本地路径安全

本地来源路径包含 `..` segment 时会被拒绝。路径应由可信业务逻辑选择。类型化
媒体会先整体读入内存再上传，因此对不可信本地输入还应增加应用层文件大小限制。

## 下载资源

```go
data, err := ch.DownloadFile(ctx, resource.FileKey, resource.Type)
if err != nil {
	return err
}
```

支持的 `mediaType`：

| 值 | 使用的 API |
|---|---|
| `image` | 图片下载 |
| `file`、`audio`、`video`、`media` | 文件下载 |

`DownloadFile` 将完整 body 作为 `[]byte` 返回，不写入文件，也不使用出站 URL
下载上限。加载不可信的大资源前应增加应用层限制。这个 helper 不支持贴纸下载。

## 相关文档

- [发送消息](sending-messages.md)
- [出站配置](reference.md)
- [媒体问题排查](troubleshooting.md)
- [E2E 媒体素材](../../e2e/README.md)
