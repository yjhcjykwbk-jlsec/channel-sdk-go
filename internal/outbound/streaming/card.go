// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package streaming

import (
	"context"
	"fmt"
	"sync"

	"github.com/larksuite/channel-sdk-go/internal/outbound"
	"github.com/larksuite/channel-sdk-go/types"
)

// CardController updates interactive card stream messages.
type CardController struct {
	driver    Driver
	config    types.ChannelConfig
	messageID string
	queue     *UpdateQueue
	mu        sync.Mutex
	isClosed  bool
}

func NewCardController(driver Driver, config types.ChannelConfig, messageID string) *CardController {
	return &CardController{
		driver:    driver,
		config:    config,
		messageID: messageID,
		queue:     NewUpdateQueue(context.Background()),
	}
}

func (c *CardController) UpdateCard(ctx context.Context, card string) error {
	c.mu.Lock()
	if c.isClosed {
		c.mu.Unlock()
		return fmt.Errorf("stream is closed")
	}
	messageID := c.messageID
	c.mu.Unlock()

	errCh := make(chan error, 1)
	if ok := c.queue.Submit(func() {
		_, err := outbound.Retry(ctx, func(attempt int) (interface{}, error) {
			return nil, c.driver.PatchMessage(ctx, messageID, card)
		}, &outbound.RetryOptions{
			MaxAttempts: c.config.Outbound.Retry.MaxAttempts,
			BaseDelay:   c.config.Outbound.Retry.BaseDelayMs,
		})
		errCh <- err
	}); !ok {
		return fmt.Errorf("stream is closed")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}

func (c *CardController) Append(ctx context.Context, text string) error {
	return fmt.Errorf("Append is not supported for CardStreamController, use UpdateCard")
}

func (c *CardController) Flush(ctx context.Context) error {
	return nil
}

func (c *CardController) Close(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.isClosed {
		return nil
	}
	c.isClosed = true
	c.queue.Stop()
	return nil
}
