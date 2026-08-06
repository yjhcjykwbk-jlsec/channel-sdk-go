// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package runtime

import (
	"github.com/larksuite/channel-sdk-go/internal/inbound"
)

func (ch *Client) MessageHandlers() []inbound.MessageHandler {
	return ch.onMessageHandlers
}

func (ch *Client) CommentHandlers() []inbound.CommentHandler {
	return ch.onCommentHandlers
}

func (ch *Client) ReactionHandlers() []inbound.ReactionHandler {
	return ch.onReactionHandlers
}

func (ch *Client) BotAddedHandlers() []inbound.BotAddedHandler {
	return ch.onBotAddedHandlers
}

func (ch *Client) CardActionHandlers() []inbound.CardActionHandler {
	return ch.onCardActionHandlers
}

func (ch *Client) RejectHandlers() []inbound.RejectHandler {
	return ch.onRejectHandlers
}
