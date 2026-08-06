// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package converters

import "github.com/larksuite/channel-sdk-go/types"

func ConvertText(msgType string, content map[string]interface{}) (string, []types.Resource) {
	if text, ok := content["text"].(string); ok {
		return text, nil
	}
	return "", nil
}
