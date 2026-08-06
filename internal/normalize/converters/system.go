// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package converters

import (
	"fmt"
	"strings"

	"github.com/larksuite/channel-sdk-go/types"
)

func ConvertSystem(msgType string, content map[string]interface{}) (string, []types.Resource) {
	template, ok := content["template"].(string)
	if !ok || template == "" {
		return "[system message]", nil
	}
	out := template
	for k, v := range content {
		if k == "template" {
			continue
		}
		placeholder := fmt.Sprintf("{%s}", k)
		if !strings.Contains(out, placeholder) {
			continue
		}
		var strVal string
		switch val := v.(type) {
		case string:
			strVal = val
		case []interface{}:
			var strVals []string
			for _, item := range val {
				strVals = append(strVals, fmt.Sprintf("%v", item))
			}
			strVal = strings.Join(strVals, ", ")
		default:
			strVal = fmt.Sprintf("%v", val)
		}
		out = strings.ReplaceAll(out, placeholder, strVal)
	}
	out = strings.TrimSpace(out)
	if out == "" {
		out = "[system message]"
	}
	return out, nil
}
