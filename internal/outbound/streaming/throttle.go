// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package streaming

import (
	"context"
	"fmt"
	"sync"
	"time"
)

const defaultThrottleInterval = 500 * time.Millisecond

type throttleController struct {
	mu           sync.Mutex
	interval     time.Duration
	lastExecTime time.Time
	timer        *time.Timer
	isClosed     bool
	exec         func(ctx context.Context) error
	onAsyncError func(error)
}

func newThrottleController(interval time.Duration, exec func(ctx context.Context) error, onAsyncError func(error)) *throttleController {
	if interval <= 0 {
		interval = defaultThrottleInterval
	}
	return &throttleController{
		interval:     interval,
		exec:         exec,
		onAsyncError: onAsyncError,
	}
}

func (t *throttleController) Trigger(ctx context.Context) error {
	t.mu.Lock()
	if t.isClosed {
		t.mu.Unlock()
		return fmt.Errorf("stream is closed")
	}

	now := time.Now()
	if now.Sub(t.lastExecTime) >= t.interval {
		if t.timer != nil {
			t.timer.Stop()
			t.timer = nil
		}
		t.lastExecTime = now
		t.mu.Unlock()

		return t.exec(ctx)
	}

	if t.timer == nil {
		waitDur := t.interval - now.Sub(t.lastExecTime)
		t.timer = time.AfterFunc(waitDur, func() {
			t.mu.Lock()
			if t.isClosed {
				t.mu.Unlock()
				return
			}
			t.lastExecTime = time.Now()
			t.timer = nil
			t.mu.Unlock()

			if err := t.exec(context.Background()); err != nil && t.onAsyncError != nil {
				t.onAsyncError(err)
			}
		})
	}
	t.mu.Unlock()
	return nil
}

func (t *throttleController) Flush(ctx context.Context) error {
	t.mu.Lock()
	if t.isClosed {
		t.mu.Unlock()
		return fmt.Errorf("stream is closed")
	}

	if t.timer != nil {
		t.timer.Stop()
		t.timer = nil
	}
	t.lastExecTime = time.Now()
	t.mu.Unlock()

	return t.exec(ctx)
}

func (t *throttleController) Close(ctx context.Context) error {
	t.mu.Lock()
	if t.isClosed {
		t.mu.Unlock()
		return nil
	}

	if t.timer != nil {
		t.timer.Stop()
		t.timer = nil
		t.mu.Unlock()

		err := t.exec(ctx)
		t.mu.Lock()
		t.isClosed = true
		t.mu.Unlock()
		return err
	}

	t.isClosed = true
	t.mu.Unlock()
	return nil
}
