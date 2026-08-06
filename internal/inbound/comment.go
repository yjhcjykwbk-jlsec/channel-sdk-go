// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package inbound

import (
	"context"
	"fmt"

	"github.com/larksuite/channel-sdk-go/internal/normalize"
	larkevent "github.com/larksuite/oapi-sdk-go/v3/event"
)

func (d *Dispatcher) HandleComment(ctx context.Context, event *larkevent.EventReq) error {
	handlers := d.commentHandlers()
	if len(handlers) == 0 {
		return nil
	}
	commentEvent := normalize.ParseComment(event)
	if commentEvent == nil || commentEvent.CommentID == "" {
		return nil
	}

	dedupKey := commentDedupKey(commentEvent)
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
			if err := h(ctx, commentEvent); err != nil {
				return fmt.Errorf("comment handler: %w", err)
			}
		}
		return nil
	}

	var err error
	if d.pipeline != nil {
		err = d.pipeline.Run(ctx, commentEvent.FileToken, run)
	} else {
		err = run()
	}
	d.emitError(ctx, err)
	return nil
}

func (d *Dispatcher) commentHandlers() []CommentHandler {
	if d.handlers == nil {
		return nil
	}
	return d.handlers.CommentHandlers()
}
