// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package inbound

import (
	"context"
	"fmt"

	"github.com/larksuite/channel-sdk-go/internal/normalize"
	"github.com/larksuite/channel-sdk-go/internal/safety"
	"github.com/larksuite/channel-sdk-go/types"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

func (d *Dispatcher) HandleMessage(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
	handlers := d.messageHandlers()
	if len(handlers) == 0 {
		return nil
	}

	normMsg := normalize.ParseMessage(event)
	if normMsg == nil {
		return nil
	}

	if d.applyBotIdentity(ctx, normMsg) {
		return nil
	}
	if safety.IsStale(normMsg.CreateTimeMs, d.staleWindow) {
		return nil
	}
	if d.dedup != nil && d.dedup.IsDuplicate(normMsg.MessageID) {
		return nil
	}

	decision := types.PolicyDecision{Allowed: true}
	if d.policy != nil {
		decision = d.policy.Evaluate(normMsg)
	}
	if !decision.Allowed {
		d.emitReject(ctx, normMsg, decision)
		return nil
	}

	if d.lock != nil && !d.lock.Acquire(normMsg.MessageID) {
		return nil
	}

	dispatchHandler := func(ctx context.Context, batch *types.BatchedDispatch) error {
		if d.lock != nil {
			defer func() {
				for _, id := range batch.SourceIDs {
					d.lock.Release(id)
				}
			}()
		}
		for _, h := range handlers {
			if err := h(ctx, batch.Message); err != nil {
				return fmt.Errorf("message handler: %w", err)
			}
		}
		return nil
	}

	if d.pipeline == nil {
		err := dispatchHandler(ctx, &types.BatchedDispatch{
			Message:   normMsg,
			SourceIDs: []string{normMsg.MessageID},
		})
		d.emitError(ctx, err)
		return nil
	}
	d.pipeline.Push(ctx, normMsg.ChatID, normMsg, dispatchHandler)
	return nil
}

func (d *Dispatcher) applyBotIdentity(ctx context.Context, msg *types.NormalizedMessage) bool {
	if d.botIdentity == nil {
		return false
	}
	botInfo := d.botIdentity.GetBotIdentity(ctx)
	if botInfo == nil {
		return false
	}
	if msg.UserID == botInfo.OpenID {
		return true
	}
	for i := range msg.Mentions {
		m := &msg.Mentions[i]
		if m.OpenID == botInfo.OpenID || m.UserID == botInfo.OpenID || (botInfo.UserID != "" && m.UserID == botInfo.UserID) {
			msg.MentionedBot = true
			m.IsBot = true
		}
	}
	return false
}

func (d *Dispatcher) emitReject(ctx context.Context, msg *types.NormalizedMessage, decision types.PolicyDecision) {
	handlers := d.rejectHandlers()
	if len(handlers) == 0 {
		return
	}
	rejectEvent := &types.RejectEvent{
		MessageID: msg.MessageID,
		ChatID:    msg.ChatID,
		SenderID:  msg.UserID,
		Reason:    string(decision.Reason),
	}
	for _, h := range handlers {
		if err := h(ctx, rejectEvent); err != nil {
			d.emitError(ctx, fmt.Errorf("reject handler: %w", err))
		}
	}
}

func (d *Dispatcher) messageHandlers() []MessageHandler {
	if d.handlers == nil {
		return nil
	}
	return d.handlers.MessageHandlers()
}

func (d *Dispatcher) rejectHandlers() []RejectHandler {
	if d.handlers == nil {
		return nil
	}
	return d.handlers.RejectHandlers()
}
