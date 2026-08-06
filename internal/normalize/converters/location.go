// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package converters

import (
	"fmt"

	"github.com/larksuite/channel-sdk-go/types"
)

func ConvertLocation(msgType string, content map[string]interface{}) (string, []types.Resource) {
	name := stringValue(content, "name")
	lat := stringValue(content, "latitude")
	lng := stringValue(content, "longitude")
	attr := ""
	if name != "" {
		attr += fmt.Sprintf(` name="%s"`, escapeAttr(name))
	}
	if lat != "" && lng != "" {
		attr += fmt.Sprintf(` coords="lat:%s,lng:%s"`, lat, lng)
	}
	return fmt.Sprintf(`<location%s/>`, attr), nil
}
