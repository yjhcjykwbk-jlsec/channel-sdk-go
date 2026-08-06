// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package runtime

import (
	"bytes"
	"context"
	"fmt"

	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

// DownloadFile downloads media by key and type (e.g., "image", "file").
func (ch *Client) DownloadFile(ctx context.Context, fileKey string, mediaType string) ([]byte, error) {
	if fileKey == "" {
		return nil, fmt.Errorf("fileKey cannot be empty")
	}

	if mediaType == "image" {
		req := larkim.NewGetImageReqBuilder().
			ImageKey(fileKey).
			Build()
		resp, err := ch.client.Im.V1.Image.Get(ctx, req)
		if err != nil {
			return nil, err
		}
		if !resp.Success() {
			return nil, fmt.Errorf("download image API error: %d - %s", resp.Code, resp.Msg)
		}
		var buf bytes.Buffer
		_, err = buf.ReadFrom(resp.File)
		if err != nil {
			return nil, fmt.Errorf("failed to read image stream: %w", err)
		}
		return buf.Bytes(), nil

	} else if mediaType == "file" || mediaType == "audio" || mediaType == "video" || mediaType == "media" {
		req := larkim.NewGetFileReqBuilder().
			FileKey(fileKey).
			Build()
		resp, err := ch.client.Im.V1.File.Get(ctx, req)
		if err != nil {
			return nil, err
		}
		if !resp.Success() {
			return nil, fmt.Errorf("download file API error: %d - %s", resp.Code, resp.Msg)
		}
		var buf bytes.Buffer
		_, err = buf.ReadFrom(resp.File)
		if err != nil {
			return nil, fmt.Errorf("failed to read file stream: %w", err)
		}
		return buf.Bytes(), nil
	}

	return nil, fmt.Errorf("unsupported mediaType: %s", mediaType)
}
