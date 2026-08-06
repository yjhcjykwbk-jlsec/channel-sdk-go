// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package runtime

import (
	"context"
	"fmt"

	"github.com/larksuite/channel-sdk-go/internal/outbound/streaming"
	"github.com/larksuite/channel-sdk-go/types"
)

// Stream initiates a streaming message session. It returns a StreamController to append and flush content.
func (ch *Client) Stream(ctx context.Context, input *types.SendInput) (types.StreamController, error) {
	if input == nil {
		return nil, fmt.Errorf("input cannot be nil")
	}

	driver := streaming.NewLarkDriver(ch.client)
	if input.Card != "" {
		res, err := ch.Send(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to send initial card message: %w", err)
		}
		return streaming.NewCardController(driver, ch.config, res.MessageID), nil
	}

	if input.Markdown == "" && input.Text == "" {
		input.Markdown = "..."
	}

	res, err := ch.Send(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to send initial message for streaming: %w", err)
	}

	return streaming.NewMarkdownController(driver, ch.config, res.MessageID, input.Markdown, input.Title), nil
}
