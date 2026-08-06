// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package streaming

import (
	"context"
	"sync"

	"github.com/larksuite/channel-sdk-go/internal/outbound"
)

type streamOp struct {
	messageID string
	msgType   string
	content   string
}

type fakeDriver struct {
	mu           sync.Mutex
	updates      []streamOp
	replies      []streamOp
	patches      []streamOp
	updateErr    error
	replyErr     error
	patchErr     error
	nextReplyIDs []string
}

func (f *fakeDriver) UpdateMessage(ctx context.Context, messageID, msgType, content string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.updates = append(f.updates, streamOp{
		messageID: messageID,
		msgType:   msgType,
		content:   content,
	})
	return f.updateErr
}

func (f *fakeDriver) ReplyMessage(ctx context.Context, messageID, msgType, content string) (*outbound.SendResponse, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.replies = append(f.replies, streamOp{
		messageID: messageID,
		msgType:   msgType,
		content:   content,
	})
	if f.replyErr != nil {
		return nil, f.replyErr
	}
	messageID = "om_reply"
	if len(f.nextReplyIDs) > 0 {
		messageID = f.nextReplyIDs[0]
		f.nextReplyIDs = f.nextReplyIDs[1:]
	}
	return &outbound.SendResponse{MessageID: messageID}, nil
}

func (f *fakeDriver) PatchMessage(ctx context.Context, messageID, content string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.patches = append(f.patches, streamOp{
		messageID: messageID,
		content:   content,
	})
	return f.patchErr
}

func (f *fakeDriver) updateCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.updates)
}

func (f *fakeDriver) replyCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.replies)
}

func (f *fakeDriver) patchCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.patches)
}
