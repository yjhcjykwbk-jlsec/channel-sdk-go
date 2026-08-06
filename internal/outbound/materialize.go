// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package outbound

import (
	"encoding/json"
	"fmt"

	"github.com/larksuite/channel-sdk-go/types"
)

func buildStaticContent(input *types.SendInput) (string, string, bool, error) {
	msgType := input.MsgType

	if input.ImageKey != "" {
		content, err := marshalContent(map[string]string{"image_key": input.ImageKey})
		return "image", content, true, err
	}
	if input.AudioKey != "" {
		content, err := marshalContent(map[string]string{"file_key": input.AudioKey})
		return "audio", content, true, err
	}
	if input.VideoKey != "" {
		content, err := marshalContent(map[string]string{"file_key": input.VideoKey})
		return "media", content, true, err
	}
	if input.FileKey != "" {
		content, err := marshalContent(map[string]string{"file_key": input.FileKey})
		return "file", content, true, err
	}
	if input.Card != "" {
		return "interactive", input.Card, true, nil
	}
	if input.Post != "" {
		return "post", input.Post, true, nil
	}
	if input.ShareChatID != "" {
		content, err := marshalContent(map[string]string{"chat_id": input.ShareChatID})
		return "share_chat", content, true, err
	}
	if input.ShareUserID != "" {
		content, err := marshalContent(map[string]string{"user_id": input.ShareUserID})
		return "share_user", content, true, err
	}
	if input.StickerFileKey != "" {
		content, err := marshalContent(map[string]string{"file_key": input.StickerFileKey})
		return "sticker", content, true, err
	}
	if msgType != "" {
		return msgType, "", false, nil
	}
	return "", "", false, nil
}

func marshalContent(content map[string]string) (string, error) {
	bytes, err := json.Marshal(content)
	if err != nil {
		return "", fmt.Errorf("marshal message content: %w", err)
	}
	return string(bytes), nil
}

func splitPlain(text string, limit int) []string {
	if limit <= 0 || len(text) <= limit {
		return []string{text}
	}
	var out []string
	runes := []rune(text)
	for i := 0; i < len(runes); i += limit {
		end := i + limit
		if end > len(runes) {
			end = len(runes)
		}
		out = append(out, string(runes[i:end]))
	}
	return out
}
