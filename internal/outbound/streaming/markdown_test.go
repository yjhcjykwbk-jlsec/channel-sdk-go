// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package streaming

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/larksuite/channel-sdk-go/types"
)

func TestMarkdownControllerAppendTriggersUpdate(t *testing.T) {
	driver := &fakeDriver{}
	cfg := types.DefaultChannelConfig()
	cfg.Outbound.StreamThrottleMs = time.Hour
	cfg.Outbound.Retry.MaxAttempts = 1
	ctrl := NewMarkdownController(driver, cfg, "om_1", "hello", "Title")

	if err := ctrl.Append(context.Background(), " world"); err != nil {
		t.Fatalf("append: %v", err)
	}

	if got := driver.updateCount(); got != 1 {
		t.Fatalf("update count = %d, want 1", got)
	}
	driver.mu.Lock()
	update := driver.updates[0]
	driver.mu.Unlock()
	if update.messageID != "om_1" || update.msgType != "post" {
		t.Fatalf("unexpected update op: %+v", update)
	}
	if !strings.Contains(update.content, "hello world") {
		t.Fatalf("expected updated content to contain appended text, got %s", update.content)
	}
}

func TestMarkdownControllerFlushTriggersUpdate(t *testing.T) {
	driver := &fakeDriver{}
	cfg := types.DefaultChannelConfig()
	cfg.Outbound.Retry.MaxAttempts = 1
	ctrl := NewMarkdownController(driver, cfg, "om_1", "hello", "Title")

	if err := ctrl.Flush(context.Background()); err != nil {
		t.Fatalf("flush: %v", err)
	}
	if got := driver.updateCount(); got != 1 {
		t.Fatalf("update count = %d, want 1", got)
	}
}

func TestMarkdownControllerRepliesWhenContentExceedsChunkLimit(t *testing.T) {
	driver := &fakeDriver{nextReplyIDs: []string{"om_2"}}
	cfg := types.DefaultChannelConfig()
	cfg.Outbound.TextChunkLimit = 5
	cfg.Outbound.StreamThrottleMs = time.Hour
	cfg.Outbound.Retry.MaxAttempts = 1
	ctrl := NewMarkdownController(driver, cfg, "om_1", "aaa", "Title")

	if err := ctrl.Append(context.Background(), "\nbbb\nccc"); err != nil {
		t.Fatalf("append: %v", err)
	}
	if got := driver.replyCount(); got != 1 {
		t.Fatalf("reply count = %d, want 1", got)
	}

	driver.mu.Lock()
	reply := driver.replies[0]
	driver.mu.Unlock()
	if reply.messageID != "om_1" || reply.msgType != "post" {
		t.Fatalf("unexpected reply op: %+v", reply)
	}

	if err := ctrl.Flush(context.Background()); err != nil {
		t.Fatalf("flush after reply: %v", err)
	}
	driver.mu.Lock()
	lastUpdate := driver.updates[len(driver.updates)-1]
	driver.mu.Unlock()
	if lastUpdate.messageID != "om_2" {
		t.Fatalf("update after reply messageID = %q, want om_2", lastUpdate.messageID)
	}
}

func TestMarkdownControllerRejectsAppendAfterClose(t *testing.T) {
	driver := &fakeDriver{}
	cfg := types.DefaultChannelConfig()
	cfg.Outbound.Retry.MaxAttempts = 1
	ctrl := NewMarkdownController(driver, cfg, "om_1", "hello", "Title")

	if err := ctrl.Close(context.Background()); err != nil {
		t.Fatalf("close: %v", err)
	}
	err := ctrl.Append(context.Background(), " world")
	if err == nil || !strings.Contains(err.Error(), "stream is closed") {
		t.Fatalf("expected closed stream error, got %v", err)
	}
}
