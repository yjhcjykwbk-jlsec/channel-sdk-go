// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package converters

import (
	"fmt"
	"strings"

	"github.com/larksuite/channel-sdk-go/types"
)

func ConvertVideoChat(msgType string, content map[string]interface{}) (string, []types.Resource) {
	topic := stringValue(content, "topic")
	startTimeStr := numericStringValue(content, "start_time")
	lines := []string{}
	if topic != "" {
		lines = append(lines, "📹 "+topic)
	}
	if dt := millisToDatetime(startTimeStr); dt != "" {
		lines = append(lines, "🕙 "+dt)
	}
	inner := "[video chat]"
	if len(lines) > 0 {
		inner = strings.Join(lines, "\n")
	}
	return fmt.Sprintf("<meeting>\n%s\n</meeting>", inner), nil
}
