// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package converters

import (
	"fmt"

	"github.com/larksuite/channel-sdk-go/types"
)

func ConvertAudio(msgType string, content map[string]interface{}) (string, []types.Resource) {
	fileKey := stringValue(content, "file_key")
	if fileKey == "" {
		return "[audio]", nil
	}
	res := types.Resource{Type: "audio", FileKey: fileKey}
	attr := ""
	if durFloat, ok := content["duration"].(float64); ok {
		durMs := int(durFloat)
		res.DurationMs = &durMs
		if durStr := formatDuration(durMs); durStr != "" {
			attr = fmt.Sprintf(` duration="%s"`, durStr)
		}
	}
	return fmt.Sprintf(`<audio key="%s"%s/>`, fileKey, attr), []types.Resource{res}
}
