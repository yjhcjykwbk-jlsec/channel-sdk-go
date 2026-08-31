// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package converters

import (
	"fmt"
	"strings"

	"github.com/larksuite/channel-sdk-go/types"
)

func ConvertPost(msgType string, content map[string]interface{}) (string, []types.Resource) {
	var resources []types.Resource

	var bodyMap map[string]interface{}
	for _, v := range content {
		if m, ok := v.(map[string]interface{}); ok {
			bodyMap = m
			break
		}
	}
	if bodyMap == nil {
		return "[rich text message]", nil
	}

	var sourceParagraphs []interface{}
	if cv2, ok := bodyMap["content_v2"].([]interface{}); ok && len(cv2) > 0 {
		sourceParagraphs = cv2
	} else if cl, ok := bodyMap["content"].([]interface{}); ok {
		sourceParagraphs = cl
	}
	if len(sourceParagraphs) == 0 {
		return "[rich text message]", nil
	}

	lines := []string{}
	if title, _ := bodyMap["title"].(string); title != "" {
		lines = append(lines, fmt.Sprintf("**%s**", title))
		lines = append(lines, "")
	}

	for _, paragraphInterface := range sourceParagraphs {
		paragraph, ok := paragraphInterface.([]interface{})
		if !ok {
			continue
		}
		var lineParts []string
		for _, elInterface := range paragraph {
			el, ok := elInterface.(map[string]interface{})
			if !ok {
				continue
			}
			tag, _ := el["tag"].(string)
			text, _ := el["text"].(string)

			switch tag {
			case "md":
				mdText, mdRes := processMdText(text)
				lineParts = append(lineParts, mdText)
				resources = append(resources, mdRes...)
			case "text":
				lineParts = append(lineParts, text)
			case "a":
				href, _ := el["href"].(string)
				label := text
				if label == "" {
					label = href
				}
				if href != "" {
					lineParts = append(lineParts, fmt.Sprintf("[%s](%s)", label, href))
				} else {
					lineParts = append(lineParts, label)
				}
			case "at":
				userID, _ := el["user_id"].(string)
				userName, _ := el["user_name"].(string)
				if userID == "all" || userID == "all_members" {
					lineParts = append(lineParts, "@all")
				} else if userName != "" {
					lineParts = append(lineParts, "@"+userName)
				} else {
					lineParts = append(lineParts, "@"+userID)
				}
			case "img":
				imageKey, _ := el["image_key"].(string)
				if imageKey != "" {
					resources = append(resources, types.Resource{Type: "image", FileKey: imageKey})
					lineParts = append(lineParts, fmt.Sprintf("![image](%s)", imageKey))
				}
			case "media":
				fileKey, _ := el["file_key"].(string)
				if fileKey != "" {
					resources = append(resources, types.Resource{Type: "file", FileKey: fileKey})
					lineParts = append(lineParts, fmt.Sprintf(`<file key="%s"/>`, fileKey))
				}
			case "code_block":
				lang, _ := el["language"].(string)
				lineParts = append(lineParts, fmt.Sprintf("\n```%s\n%s\n```\n", lang, text))
			case "hr":
				lineParts = append(lineParts, "\n---\n")
			default:
				lineParts = append(lineParts, text)
			}
		}
		lines = append(lines, strings.Join(lineParts, ""))
	}

	// Attachment zone: the top-level `files` array of a post message, outside
	// any locale document. Files render as <file .../> (same tag style as the
	// standalone file converter) and are surfaced as downloadable resources;
	// folders render as <folder .../> tags only (mirrors the standalone folder
	// converter, resources=nil). Both key and name are escaped: downstream
	// parses these tags as structured info, so a quote inside a key must not
	// be able to forge an extra attribute.
	if files, ok := content["files"].([]interface{}); ok {
		for _, fInterface := range files {
			f, ok := fInterface.(map[string]interface{})
			if !ok {
				continue
			}
			fileKey := stringValue(f, "file_key")
			if fileKey == "" {
				continue
			}
			fileName := stringValue(f, "file_name")
			isFolder, _ := f["is_folder"].(bool)
			if isFolder {
				if fileName != "" {
					lines = append(lines, fmt.Sprintf(`<folder key="%s" name="%s"/>`, escapeAttr(fileKey), escapeAttr(fileName)))
				} else {
					lines = append(lines, fmt.Sprintf(`<folder key="%s"/>`, escapeAttr(fileKey)))
				}
			} else {
				if fileName != "" {
					lines = append(lines, fmt.Sprintf(`<file key="%s" name="%s"/>`, escapeAttr(fileKey), escapeAttr(fileName)))
				} else {
					lines = append(lines, fmt.Sprintf(`<file key="%s"/>`, escapeAttr(fileKey)))
				}
				resources = append(resources, types.Resource{Type: "file", FileKey: fileKey, FileName: fileName})
			}
		}
	}

	contentStr := strings.TrimSpace(strings.Join(lines, "\n"))
	if contentStr == "" {
		contentStr = "[rich text message]"
	}
	return contentStr, resources
}
