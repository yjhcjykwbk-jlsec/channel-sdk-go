// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package converters

import (
	"fmt"
	"strings"

	"github.com/larksuite/channel-sdk-go/types"
)

func ConvertVote(msgType string, content map[string]interface{}) (string, []types.Resource) {
	topic := stringValue(content, "topic")
	optionsInterface, _ := content["options"].([]interface{})
	var options []string
	for _, opt := range optionsInterface {
		if s, ok := opt.(string); ok {
			options = append(options, s)
		}
	}
	if topic == "" && len(options) == 0 {
		return "<vote>\n[vote]\n</vote>", nil
	}
	lines := []string{}
	if topic != "" {
		lines = append(lines, topic)
	}
	for _, opt := range options {
		lines = append(lines, "• "+opt)
	}
	return fmt.Sprintf("<vote>\n%s\n</vote>", strings.Join(lines, "\n")), nil
}
