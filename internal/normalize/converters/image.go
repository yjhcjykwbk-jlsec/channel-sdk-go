// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package converters

import (
	"fmt"

	"github.com/larksuite/channel-sdk-go/types"
)

func ConvertImage(msgType string, content map[string]interface{}) (string, []types.Resource) {
	imageKey := stringValue(content, "image_key")
	if imageKey == "" {
		return "[image]", nil
	}
	return fmt.Sprintf("![image](%s)", imageKey), []types.Resource{{Type: "image", FileKey: imageKey}}
}
