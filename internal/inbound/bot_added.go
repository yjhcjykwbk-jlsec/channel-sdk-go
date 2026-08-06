// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package inbound

import (
	"context"

	"github.com/larksuite/channel-sdk-go/internal/normalize"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

func (d *Dispatcher) HandleBotAdded(ctx context.Context, event *larkim.P2ChatMemberBotAddedV1) error {
	handlers := d.botAddedHandlers()
	if len(handlers) == 0 {
		return nil
	}
	botAddedEvent := normalize.ParseBotAdded(event)
	if botAddedEvent == nil {
		return nil
	}
	if d.dedup != nil && d.dedup.IsDuplicate(botAddedEvent.EventID) {
		return nil
	}
	if d.lock != nil {
		if !d.lock.Acquire(botAddedEvent.EventID) {
			return nil
		}
		defer d.lock.Release(botAddedEvent.EventID)
	}

	run := func() error {
		for _, h := range handlers {
			if err := h(ctx, botAddedEvent); err != nil {
				return err
			}
		}
		return nil
	}
	if d.pipeline != nil {
		return d.pipeline.Run(ctx, botAddedEvent.ChatID, run)
	}
	return run()
}

func (d *Dispatcher) botAddedHandlers() []BotAddedHandler {
	if d.handlers == nil {
		return nil
	}
	return d.handlers.BotAddedHandlers()
}
