// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package streaming

import (
	"context"
	"fmt"
	"sync"

	"github.com/larksuite/channel-sdk-go/internal/outbound"
	outboundmarkdown "github.com/larksuite/channel-sdk-go/internal/outbound/markdown"
	"github.com/larksuite/channel-sdk-go/types"
)

// MarkdownController updates markdown stream messages.
type MarkdownController struct {
	driver     Driver
	config     types.ChannelConfig
	messageID  string
	title      string
	content    string
	chunkIndex int

	mu       sync.Mutex
	throttle *throttleController
}

func NewMarkdownController(driver Driver, config types.ChannelConfig, messageID, initialContent, title string) *MarkdownController {
	m := &MarkdownController{
		driver:     driver,
		config:     config,
		messageID:  messageID,
		title:      title,
		content:    initialContent,
		chunkIndex: 0,
	}

	m.throttle = newThrottleController(config.Outbound.StreamThrottleMs, m.doUpdate, nil)
	return m
}

func (m *MarkdownController) Append(ctx context.Context, text string) error {
	m.mu.Lock()
	m.content += text
	m.mu.Unlock()

	return m.throttle.Trigger(ctx)
}

func (m *MarkdownController) UpdateCard(ctx context.Context, card string) error {
	return fmt.Errorf("UpdateCard is not supported for MarkdownStreamController, use Append")
}

func (m *MarkdownController) Flush(ctx context.Context) error {
	return m.throttle.Flush(ctx)
}

func (m *MarkdownController) Close(ctx context.Context) error {
	return m.throttle.Close(ctx)
}

func (m *MarkdownController) doUpdate(ctx context.Context) error {
	m.mu.Lock()
	currentContent := m.content
	currentIndex := m.chunkIndex
	currentMessageID := m.messageID
	m.mu.Unlock()

	chunks := outboundmarkdown.SplitWithCodeFences(currentContent, m.config.Outbound.TextChunkLimit)
	if len(chunks) == 0 {
		return nil
	}

	targetIndex := len(chunks) - 1
	targetChunk := chunks[targetIndex]
	contentStr, err := outboundmarkdown.SimpleMarkdownToPost(m.title, targetChunk, nil)
	if err != nil {
		return fmt.Errorf("failed to marshal post content: %w", err)
	}

	if targetIndex > currentIndex {
		res, err := outbound.Retry(ctx, func(attempt int) (interface{}, error) {
			return m.driver.ReplyMessage(ctx, currentMessageID, "post", contentStr)
		}, &outbound.RetryOptions{
			MaxAttempts: m.config.Outbound.Retry.MaxAttempts,
			BaseDelay:   m.config.Outbound.Retry.BaseDelayMs,
		})
		if err != nil {
			return err
		}

		reply, ok := res.(*outbound.SendResponse)
		if !ok || reply == nil || reply.MessageID == "" {
			return fmt.Errorf("reply message failed: empty response data")
		}

		m.mu.Lock()
		m.messageID = reply.MessageID
		m.chunkIndex = targetIndex
		m.mu.Unlock()
		return nil
	}

	_, err = outbound.Retry(ctx, func(attempt int) (interface{}, error) {
		return nil, m.driver.UpdateMessage(ctx, currentMessageID, "post", contentStr)
	}, &outbound.RetryOptions{
		MaxAttempts: m.config.Outbound.Retry.MaxAttempts,
		BaseDelay:   m.config.Outbound.Retry.BaseDelayMs,
	})
	return err
}
