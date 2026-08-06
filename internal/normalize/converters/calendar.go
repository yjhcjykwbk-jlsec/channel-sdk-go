// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package converters

import (
	"fmt"
	"strings"

	"github.com/larksuite/channel-sdk-go/types"
)

func ConvertCalendar(msgType string, content map[string]interface{}) (string, []types.Resource) {
	summary := stringValue(content, "summary")
	startTimeStr := numericStringValue(content, "start_time")
	endTimeStr := numericStringValue(content, "end_time")

	lines := []string{}
	if summary != "" {
		lines = append(lines, "📅 "+summary)
	}
	start := millisToDatetime(startTimeStr)
	end := millisToDatetime(endTimeStr)
	if start != "" && end != "" {
		lines = append(lines, fmt.Sprintf("🕙 %s ~ %s", start, end))
	} else if start != "" {
		lines = append(lines, fmt.Sprintf("🕙 %s", start))
	}
	inner := "[calendar event]"
	if len(lines) > 0 {
		inner = strings.Join(lines, "\n")
	}

	tag := "calendar"
	if msgType == "calendar" {
		tag = "calendar_invite"
	} else if msgType == "share_calendar_event" {
		tag = "calendar_share"
	}
	return fmt.Sprintf("<%s>\n%s\n</%s>", tag, inner, tag), nil
}
