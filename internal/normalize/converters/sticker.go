// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package converters

import (
	"fmt"

	"github.com/larksuite/channel-sdk-go/types"
)

func ConvertSticker(msgType string, content map[string]interface{}) (string, []types.Resource) {
	fileKey := stringValue(content, "file_key")
	if fileKey == "" {
		return "[sticker]", nil
	}
	return fmt.Sprintf(`<sticker key="%s"/>`, fileKey), []types.Resource{{Type: "sticker", FileKey: fileKey}}
}
