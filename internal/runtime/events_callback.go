// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package runtime

import (
	"context"

	"github.com/larksuite/channel-sdk-go/types"
	larkevent "github.com/larksuite/oapi-sdk-go/v3/event"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher/callback"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

// OnComment registers a handler for CommentEvent.
func (ch *Client) OnComment(handler func(ctx context.Context, event *types.CommentEvent) error) {
	ch.onCommentHandlers = append(ch.onCommentHandlers, handler)
	if ch.commentHandlerReg || ch.wsClient == nil {
		return
	}
	ch.commentHandlerReg = true
	dispatcher := ch.wsClient.EventHandler()
	if dispatcher != nil {
		dispatcher.OnCustomizedEvent("drive.notice.comment_add_v1", func(ctx context.Context, event *larkevent.EventReq) error {
			return ch.inbound.HandleComment(ctx, event)
		})
	}
}

// OnBotAdded registers a handler for BotAddedEvent.
func (ch *Client) OnBotAdded(handler func(ctx context.Context, event *types.BotAddedEvent) error) {
	ch.onBotAddedHandlers = append(ch.onBotAddedHandlers, handler)
	if ch.botAddedHandlerReg || ch.wsClient == nil {
		return
	}
	ch.botAddedHandlerReg = true
	dispatcher := ch.wsClient.EventHandler()
	if dispatcher != nil {
		dispatcher.OnP2ChatMemberBotAddedV1(func(ctx context.Context, event *larkim.P2ChatMemberBotAddedV1) error {
			return ch.inbound.HandleBotAdded(ctx, event)
		})
	}
}

// OnReaction registers a handler for ReactionEvent.
func (ch *Client) OnReaction(handler func(ctx context.Context, event *types.ReactionEvent) error) {
	ch.onReactionHandlers = append(ch.onReactionHandlers, handler)
	if ch.reactionHandlerReg || ch.wsClient == nil {
		return
	}
	ch.reactionHandlerReg = true
	dispatcher := ch.wsClient.EventHandler()
	if dispatcher != nil {
		dispatcher.OnP2MessageReactionCreatedV1(func(ctx context.Context, event *larkim.P2MessageReactionCreatedV1) error {
			return ch.inbound.HandleReaction(ctx, event)
		})
		dispatcher.OnP2MessageReactionDeletedV1(func(ctx context.Context, event *larkim.P2MessageReactionDeletedV1) error {
			return ch.inbound.HandleReaction(ctx, event)
		})
	}
}

// OnCardAction registers a handler for CardActionEvent events.
func (ch *Client) OnCardAction(handler func(ctx context.Context, event *types.CardActionEvent) error) {
	ch.onCardActionHandlers = append(ch.onCardActionHandlers, handler)
	if ch.wsClient == nil {
		return
	}
	dispatcher := ch.wsClient.EventHandler()
	if dispatcher != nil {
		dispatcher.OnP2CardActionTrigger(func(ctx context.Context, event *callback.CardActionTriggerEvent) (*callback.CardActionTriggerResponse, error) {
			return ch.inbound.HandleCardAction(ctx, event)
		})
	}
}

// OnReject registers a handler for messages rejected by safety policies.
func (ch *Client) OnReject(handler func(ctx context.Context, event *types.RejectEvent) error) {
	ch.onRejectHandlers = append(ch.onRejectHandlers, handler)
}
