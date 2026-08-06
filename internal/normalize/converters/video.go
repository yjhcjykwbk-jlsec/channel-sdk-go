// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package converters

import (
	"fmt"

	"github.com/larksuite/channel-sdk-go/types"
)

func ConvertVideo(msgType string, content map[string]interface{}) (string, []types.Resource) {
	fileKey := stringValue(content, "file_key")
	fileName := stringValue(content, "file_name")
	imageKey := stringValue(content, "image_key")
	if fileKey == "" {
		return "[video]", nil
	}
	res := types.Resource{Type: "video", FileKey: fileKey, FileName: fileName, CoverImageKey: imageKey}
	attr := ""
	if fileName != "" {
		attr += fmt.Sprintf(` name="%s"`, escapeAttr(fileName))
	}
	if durFloat, ok := content["duration"].(float64); ok {
		durMs := int(durFloat)
		res.DurationMs = &durMs
		if durStr := formatDuration(durMs); durStr != "" {
			attr += fmt.Sprintf(` duration="%s"`, durStr)
		}
	}
	return fmt.Sprintf(`<video key="%s"%s/>`, fileKey, attr), []types.Resource{res}
}
