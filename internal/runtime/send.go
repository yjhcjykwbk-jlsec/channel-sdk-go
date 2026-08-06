// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package runtime

import (
	"context"
	"fmt"

	"github.com/larksuite/channel-sdk-go/types"
)

func (c *Client) Send(ctx context.Context, input *types.SendInput) (*types.SendResult, error) {
	if c.sender == nil {
		return nil, fmt.Errorf("sender is not initialized")
	}
	return c.sender.Send(ctx, input)
}
