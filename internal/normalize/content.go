// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package normalize

import (
	"encoding/json"

	"github.com/larksuite/channel-sdk-go/types"
)

// ParseContent parses the message content based on the message type.
func ParseContent(msgType string, content string) (string, []types.Resource) {
	if msgType == "merge_forward" {
		return content, nil
	}

	var contentMap map[string]interface{}
	if err := json.Unmarshal([]byte(content), &contentMap); err != nil {
		return "[unsupported message]", nil
	}

	return defaultContentRegistry.Convert(msgType, contentMap)
}
