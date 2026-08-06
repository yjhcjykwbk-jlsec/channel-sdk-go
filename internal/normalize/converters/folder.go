// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package converters

import (
	"fmt"

	"github.com/larksuite/channel-sdk-go/types"
)

func ConvertFolder(msgType string, content map[string]interface{}) (string, []types.Resource) {
	fileKey := stringValue(content, "file_key")
	fileName := stringValue(content, "file_name")
	if fileKey == "" {
		return "[folder]", nil
	}
	attr := ""
	if fileName != "" {
		attr = fmt.Sprintf(` name="%s"`, escapeAttr(fileName))
	}
	return fmt.Sprintf(`<folder key="%s"%s/>`, fileKey, attr), nil
}
