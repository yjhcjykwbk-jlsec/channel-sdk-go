// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package runtime

import (
	"context"

	"github.com/larksuite/channel-sdk-go/types"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

// OnMessage registers a handler for NormalizedMessage events.
func (ch *Client) OnMessage(handler func(ctx context.Context, msg *types.NormalizedMessage) error) {
	ch.onMessageHandlers = append(ch.onMessageHandlers, handler)
	ch.ensureMessageHandler()
}

func (ch *Client) ensureMessageHandler() {
	if ch.messageHandlerReg || ch.wsClient == nil {
		return
	}
	ch.messageHandlerReg = true
	dispatcher := ch.wsClient.EventHandler()
	if dispatcher != nil {
		dispatcher.OnP2MessageReceiveV1(func(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
			return ch.inbound.HandleMessage(ctx, event)
		})
	}
}
