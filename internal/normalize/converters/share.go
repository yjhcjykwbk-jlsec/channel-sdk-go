// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package converters

import (
	"fmt"

	"github.com/larksuite/channel-sdk-go/types"
)

func ConvertShareChat(msgType string, content map[string]interface{}) (string, []types.Resource) {
	return fmt.Sprintf(`<group_card id="%s"/>`, stringValue(content, "chat_id")), nil
}

func ConvertShareUser(msgType string, content map[string]interface{}) (string, []types.Resource) {
	return fmt.Sprintf(`<contact_card id="%s"/>`, stringValue(content, "user_id")), nil
}
