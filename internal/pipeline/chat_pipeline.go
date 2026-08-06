// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package pipeline

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/larksuite/channel-sdk-go/types"
)

type FlushHandler func(context.Context, *types.BatchedDispatch) error
type ErrorHandler func(context.Context, error)

var errChatPipelineClosed = errors.New("chat pipeline is closed")

// ChatPipeline manages debouncing, batching, and serializing tasks for a single scope.
type ChatPipeline struct {
	mu             sync.Mutex
	config         types.BatchConfig
	serialOnly     bool
	buffer         []*types.NormalizedMessage
	bufferChars    int
	timer          *time.Timer
	pendingHandler FlushHandler
	onError        ErrorHandler

	// Task queue for serialization
	tasks chan func()

	stopCh chan struct{}
	closed bool
}

func NewChatPipeline(config types.BatchConfig, serialOnly bool) *ChatPipeline {
	cp := &ChatPipeline{
		config:     config,
		serialOnly: serialOnly,
		tasks:      make(chan func(), 100),
		stopCh:     make(chan struct{}),
	}
	go cp.worker()
	return cp
}

func (cp *ChatPipeline) SetErrorHandler(handler ErrorHandler) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.onError = handler
}

func (cp *ChatPipeline) worker() {
	for {
		select {
		case task, ok := <-cp.tasks:
			if !ok {
				return
			}
			task()
		case <-cp.stopCh:
			return
		}
	}
}

func (cp *ChatPipeline) Push(ctx context.Context, msg *types.NormalizedMessage, handler FlushHandler) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	if cp.closed {
		return
	}

	cp.buffer = append(cp.buffer, msg)
	cp.bufferChars += len(msg.Content)
	if cp.pendingHandler == nil {
		cp.pendingHandler = handler
	}

	if len(cp.buffer) >= cp.config.MaxMessages || cp.bufferChars >= cp.config.MaxChars {
		cp.clearTimer()
		cp.enqueueFlush(ctx)
		return
	}

	if cp.config.DelayMs <= 0 || cp.serialOnly {
		cp.clearTimer()
		cp.enqueueFlush(ctx)
		return
	}

	cp.clearTimer()
	delay := cp.config.DelayMs
	if cp.bufferChars >= cp.config.LongThresholdChars {
		delay = cp.config.LongDelayMs
	}

	cp.timer = time.AfterFunc(delay, func() {
		cp.mu.Lock()
		defer cp.mu.Unlock()
		cp.timer = nil
		cp.enqueueFlush(ctx)
	})
}

func (cp *ChatPipeline) Run(ctx context.Context, task func() error) error {
	cp.mu.Lock()
	if cp.closed {
		cp.mu.Unlock()
		return errChatPipelineClosed
	}
	if len(cp.buffer) > 0 {
		cp.clearTimer()
		cp.enqueueFlush(ctx)
	}

	errCh := make(chan error, 1)
	cp.tasks <- func() {
		errCh <- task()
	}
	cp.mu.Unlock()

	return <-errCh
}

func (cp *ChatPipeline) FlushNow(ctx context.Context) {
	cp.mu.Lock()
	if cp.closed {
		cp.mu.Unlock()
		return
	}
	if len(cp.buffer) > 0 {
		cp.clearTimer()
		cp.enqueueFlush(ctx)
	}

	// Wait for queue to drain
	done := make(chan struct{})
	cp.tasks <- func() {
		close(done)
	}
	cp.mu.Unlock()
	<-done
}

func (cp *ChatPipeline) IsIdle() bool {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	return len(cp.buffer) == 0 && cp.timer == nil
}

func (cp *ChatPipeline) Dispose() {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	if cp.closed {
		return
	}
	cp.closed = true
	cp.clearTimer()
	close(cp.stopCh)
	close(cp.tasks)
}

func (cp *ChatPipeline) clearTimer() {
	if cp.timer != nil {
		cp.timer.Stop()
		cp.timer = nil
	}
}

func (cp *ChatPipeline) enqueueFlush(ctx context.Context) {
	if len(cp.buffer) == 0 {
		return
	}

	batch := cp.buffer
	handler := cp.pendingHandler

	cp.buffer = nil
	cp.bufferChars = 0
	cp.pendingHandler = nil

	if handler == nil {
		return
	}

	dispatch := &types.BatchedDispatch{
		Message:   mergeBatch(batch),
		SourceIDs: extractSourceIDs(batch),
	}

	cp.tasks <- func() {
		if err := handler(ctx, dispatch); err != nil {
			cp.reportError(ctx, err)
		}
	}
}

func (cp *ChatPipeline) reportError(ctx context.Context, err error) {
	if err == nil {
		return
	}
	cp.mu.Lock()
	handler := cp.onError
	cp.mu.Unlock()
	if handler != nil {
		handler(ctx, err)
	}
}
