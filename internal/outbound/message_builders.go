// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package outbound

import (
	"fmt"
	"strings"

	"github.com/larksuite/channel-sdk-go/types"
)

// ComposeMentionsTextPrefix builds a text prefix that renders as real Feishu mentions when prepended
// to a text-type outbound message (the <at ...> tag form).
func ComposeMentionsTextPrefix(mentions []types.Mention) string {
	if len(mentions) == 0 {
		return ""
	}
	var parts []string
	for _, m := range mentions {
		if m.UserID == "" {
			continue
		}
		name := m.Name
		parts = append(parts, fmt.Sprintf(`<at user_id="%s">%s</at>`, m.UserID, name))
	}
	if len(parts) > 0 {
		return strings.Join(parts, " ") + " "
	}
	return ""
}
