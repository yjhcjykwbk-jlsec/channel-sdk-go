// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package converters

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/larksuite/channel-sdk-go/types"
)

var (
	atMentionRe = regexp.MustCompile(`<at(\s+)user_id(\s*)=(\s*)"(.*?)">(.*?)</at>`)
	imageKeyRe  = regexp.MustCompile(`!\[(.*?)\]\(([^)]+)\)`)
)

func escapeAttr(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}

func formatDuration(ms int) string {
	if ms <= 0 {
		return ""
	}
	sec := ms / 1000
	if sec < 60 {
		return fmt.Sprintf("0:%02d", sec)
	}
	min := sec / 60
	sec = sec % 60
	if min < 60 {
		return fmt.Sprintf("%d:%02d", min, sec)
	}
	hr := min / 60
	min = min % 60
	return fmt.Sprintf("%d:%02d:%02d", hr, min, sec)
}

func millisToDatetime(msStr string) string {
	var ms int64
	if msStr == "" {
		return ""
	}
	_, err := fmt.Sscanf(msStr, "%d", &ms)
	if err != nil || ms <= 0 {
		return ""
	}
	t := time.Unix(ms/1000, (ms%1000)*1000000)
	loc := time.FixedZone("UTC+8", 8*3600)
	t = t.In(loc)
	return t.Format("2006-01-02 15:04:05")
}

func stringValue(content map[string]interface{}, key string) string {
	value, _ := content[key].(string)
	return value
}

func numericStringValue(content map[string]interface{}, key string) string {
	if value, ok := content[key].(string); ok {
		return value
	}
	if value, ok := content[key].(float64); ok {
		return fmt.Sprintf("%d", int64(value))
	}
	return ""
}

func extractPostPlainText(blocks []interface{}) string {
	var lines []string
	for _, paragraphInterface := range blocks {
		paragraph, ok := paragraphInterface.([]interface{})
		if !ok {
			continue
		}
		var parts []string
		for _, elInterface := range paragraph {
			el, ok := elInterface.(map[string]interface{})
			if !ok {
				continue
			}
			tag, _ := el["tag"].(string)
			text, _ := el["text"].(string)
			if tag == "text" && text != "" {
				parts = append(parts, text)
			} else if tag == "a" && text != "" {
				parts = append(parts, text)
			}
		}
		if len(parts) > 0 {
			lines = append(lines, strings.Join(parts, ""))
		}
	}
	return strings.Join(lines, "\n")
}

func processMdText(text string) (string, []types.Resource) {
	var resources []types.Resource
	parts := strings.Split(text, "```")
	total := len(parts)
	for i, part := range parts {
		isInside := i%2 == 1
		if isInside && total%2 == 0 && i == total-1 {
			isInside = false
		}
		if !isInside {
			part = atMentionRe.ReplaceAllStringFunc(part, func(match string) string {
				submatch := atMentionRe.FindStringSubmatch(match)
				if len(submatch) >= 6 {
					userID := submatch[4]
					name := submatch[5]
					if userID == "all" || userID == "all_members" {
						return "@all"
					}
					if name != "" {
						return "@" + name
					}
					return "@" + userID
				}
				return match
			})

			for _, imgMatch := range imageKeyRe.FindAllStringSubmatch(part, -1) {
				if len(imgMatch) >= 3 && imgMatch[2] != "" {
					resources = append(resources, types.Resource{Type: "image", FileKey: imgMatch[2]})
				}
			}
		}
		parts[i] = part
	}
	return strings.Join(parts, "```"), resources
}
