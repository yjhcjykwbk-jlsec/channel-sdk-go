// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package inbound

import (
	"context"

	"github.com/larksuite/channel-sdk-go/internal/normalize"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher/callback"
)

func (d *Dispatcher) HandleCardAction(ctx context.Context, event *callback.CardActionTriggerEvent) (*callback.CardActionTriggerResponse, error) {
	handlers := d.cardActionHandlers()
	if len(handlers) == 0 {
		return nil, nil
	}
	cardActionEvent := normalize.ParseCardAction(event)
	if cardActionEvent == nil {
		return nil, nil
	}

	dedupKey := cardActionDedupKey(cardActionEvent)
	if d.dedup != nil && d.dedup.IsDuplicate(dedupKey) {
		return nil, nil
	}
	if d.lock != nil {
		if !d.lock.Acquire(dedupKey) {
			return nil, nil
		}
		defer d.lock.Release(dedupKey)
	}

	scope := cardActionEvent.ChatID
	if scope == "" {
		scope = cardActionEvent.MessageID
	}
	run := func() error {
		for _, h := range handlers {
			if err := h(ctx, cardActionEvent); err != nil {
				return err
			}
		}
		return nil
	}

	var err error
	if d.pipeline != nil {
		err = d.pipeline.Run(ctx, scope, run)
	} else {
		err = run()
	}
	d.emitError(ctx, err)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (d *Dispatcher) cardActionHandlers() []CardActionHandler {
	if d.handlers == nil {
		return nil
	}
	return d.handlers.CardActionHandlers()
}
