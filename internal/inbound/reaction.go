// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package inbound

import (
	"context"

	"github.com/larksuite/channel-sdk-go/internal/normalize"
	"github.com/larksuite/channel-sdk-go/types"
)

func (d *Dispatcher) HandleReaction(ctx context.Context, event interface{}) error {
	handlers := d.reactionHandlers()
	if len(handlers) == 0 {
		return nil
	}
	reactionEvent := normalize.ParseReaction(event)
	if reactionEvent == nil {
		return nil
	}
	return d.handleReactionEvent(ctx, reactionEvent, handlers)
}

func (d *Dispatcher) handleReactionEvent(ctx context.Context, event *types.ReactionEvent, handlers []ReactionHandler) error {
	dedupKey := reactionDedupKey(event)
	if d.dedup != nil && d.dedup.IsDuplicate(dedupKey) {
		return nil
	}
	if d.lock != nil {
		if !d.lock.Acquire(dedupKey) {
			return nil
		}
		defer d.lock.Release(dedupKey)
	}

	run := func() error {
		for _, h := range handlers {
			if err := h(ctx, event); err != nil {
				return err
			}
		}
		return nil
	}

	var err error
	if d.pipeline != nil {
		err = d.pipeline.Run(ctx, event.MessageID, run)
	} else {
		err = run()
	}
	d.emitError(ctx, err)
	return err
}

func (d *Dispatcher) reactionHandlers() []ReactionHandler {
	if d.handlers == nil {
		return nil
	}
	return d.handlers.ReactionHandlers()
}
