# Media Upload and Download

[中文](zh-CN/media.md) | [Documentation index](README.md)

The Channel SDK can send an existing media key, upload a local image/file path,
or upload image, file, audio, and video data from bytes, a local path, or an
HTTP(S) URL.

## Choose the Simplest Input

| Input | Use it when |
|---|---|
| `ImageKey`, `FileKey`, `AudioKey`, `VideoKey` | The media was already uploaded |
| `ImagePath` | A local image should be uploaded as a message image |
| `FilePath` | A local file should be uploaded as a generic stream file |
| `Media` | You need bytes, URL input, audio/video duration handling, or an explicit media kind |

Existing keys avoid another upload and take precedence over automatic `Media`
upload.

## Local Image or File

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

`ImagePath` is ignored when `ImageKey` is already set. `FilePath` is ignored
when `FileKey` is already set.

## Typed Media Upload

From bytes:

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

From a local file:

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

From a URL:

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

## Source Precedence

When several `UploadInput` sources are populated:

1. non-nil `SourceBytes`;
2. `SourcePath`;
3. `SourceURL`.

An empty but non-nil `SourceBytes` is still selected and may fail during upload.
Provide exactly one source.

## Media Kinds

| Kind | Upload behavior | Default filename |
|---|---|---|
| `MediaKindImage` | Image API with `image_type=message` | `image.png` |
| `MediaKindFile` | File API with `file_type=stream` | `upload.bin` |
| `MediaKindAudio` | File API with `file_type=opus` | `voice.opus` |
| `MediaKindVideo` | File API with `file_type=mp4` | `video.mp4` |

`Duration` is measured in milliseconds. For audio and video, the SDK attempts
to read duration from Opus/OGG or MP4 data when `Duration` is nil or
non-positive. Supply it explicitly when the input is valid for the platform but
cannot be parsed by those readers.

## URL Safety

Before fetching `SourceURL`, the SDK:

- accepts only `http` and `https`;
- resolves the initial hostname;
- rejects loopback, private, link-local, multicast, reserved, and similar
  non-public IP ranges;
- limits the response body to 100 MiB by default;
- applies a 15-second HTTP client timeout;
- uses the caller's context for cancellation.

Configure an exact-host allowlist and download limit from a copy of the
defaults:

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

An allowlisted hostname bypasses the public-IP rejection, so treat this list as
a security boundary. Add only controlled hosts. URL redirects are handled by
the standard Go HTTP client; applications with stricter redirect policy should
download and validate the data themselves, then pass `SourceBytes`.

## Local Path Safety

Local source paths containing a `..` segment are rejected. Use a path selected
by trusted application logic. The SDK reads typed media files into memory
before upload; apply an application-level size limit to untrusted local input.

## Download Resources

```go
data, err := ch.DownloadFile(ctx, resource.FileKey, resource.Type)
if err != nil {
	return err
}
```

Supported `mediaType` values:

| Value | API used |
|---|---|
| `image` | Image download |
| `file`, `audio`, `video`, `media` | File download |

`DownloadFile` returns the complete body as `[]byte`; it does not write files
or impose the outbound URL download limit. Avoid loading untrusted, very large
resources without an application-level bound. Sticker download is not exposed
by this helper.

## Related Documents

- [Sending messages](sending-messages.md)
- [Outbound options](reference.md#effective-defaults)
- [Media troubleshooting](troubleshooting.md#media-upload-or-download-fails)
- [E2E media fixtures](../e2e/README.md)
