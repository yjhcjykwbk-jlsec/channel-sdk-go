// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package converters

import (
	"fmt"

	"github.com/larksuite/channel-sdk-go/types"
)

func ConvertHongbao(msgType string, content map[string]interface{}) (string, []types.Resource) {
	text := stringValue(content, "text")
	attr := ""
	if text != "" {
		attr = fmt.Sprintf(` text="%s"`, escapeAttr(text))
	}
	return fmt.Sprintf(`<hongbao%s/>`, attr), nil
}
