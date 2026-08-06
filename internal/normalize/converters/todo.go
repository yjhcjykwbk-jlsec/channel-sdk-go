// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package converters

import (
	"fmt"
	"strings"

	"github.com/larksuite/channel-sdk-go/types"
)

func ConvertTodo(msgType string, content map[string]interface{}) (string, []types.Resource) {
	summaryMap, ok := content["summary"].(map[string]interface{})
	if !ok {
		return "<todo>\n[todo]\n</todo>", nil
	}
	lines := []string{}
	if title, _ := summaryMap["title"].(string); title != "" {
		lines = append(lines, title)
	}

	if contentList, ok := summaryMap["content"].([]interface{}); ok {
		bodyText := extractPostPlainText(contentList)
		if bodyText != "" {
			lines = append(lines, bodyText)
		}
	}

	dueTimeStr := numericStringValue(content, "due_time")
	if due := millisToDatetime(dueTimeStr); due != "" {
		lines = append(lines, "Due: "+due)
	}

	if len(lines) == 0 {
		return "<todo>\n[todo]\n</todo>", nil
	}
	return fmt.Sprintf("<todo>\n%s\n</todo>", strings.Join(lines, "\n")), nil
}
