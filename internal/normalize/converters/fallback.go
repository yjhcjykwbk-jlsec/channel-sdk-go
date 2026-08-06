// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package converters

import "github.com/larksuite/channel-sdk-go/types"

func ConvertFallback(msgType string, content map[string]interface{}) (string, []types.Resource) {
	return "[unsupported message]", nil
}
