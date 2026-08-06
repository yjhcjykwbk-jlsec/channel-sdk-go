// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package streaming

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestThrottleControllerTriggerFlushAndClose(t *testing.T) {
	var execCount int32
	tc := newThrottleController(50*time.Millisecond, func(ctx context.Context) error {
		atomic.AddInt32(&execCount, 1)
		return nil
	}, nil)

	if err := tc.Trigger(context.Background()); err != nil {
		t.Fatalf("trigger: %v", err)
	}
	if got := atomic.LoadInt32(&execCount); got != 1 {
		t.Fatalf("exec count after first trigger = %d, want 1", got)
	}

	if err := tc.Trigger(context.Background()); err != nil {
		t.Fatalf("second trigger: %v", err)
	}
	if got := atomic.LoadInt32(&execCount); got != 1 {
		t.Fatalf("exec count before throttle interval = %d, want 1", got)
	}

	if err := tc.Flush(context.Background()); err != nil {
		t.Fatalf("flush: %v", err)
	}
	if got := atomic.LoadInt32(&execCount); got != 2 {
		t.Fatalf("exec count after flush = %d, want 2", got)
	}

	if err := tc.Close(context.Background()); err != nil {
		t.Fatalf("close: %v", err)
	}
	err := tc.Trigger(context.Background())
	if err == nil || !strings.Contains(err.Error(), "stream is closed") {
		t.Fatalf("expected closed stream error, got %v", err)
	}
}

func TestThrottleControllerCoalescesPendingTrigger(t *testing.T) {
	var execCount int32
	tc := newThrottleController(20*time.Millisecond, func(ctx context.Context) error {
		atomic.AddInt32(&execCount, 1)
		return nil
	}, nil)

	if err := tc.Trigger(context.Background()); err != nil {
		t.Fatalf("trigger: %v", err)
	}
	if err := tc.Trigger(context.Background()); err != nil {
		t.Fatalf("second trigger: %v", err)
	}

	time.Sleep(60 * time.Millisecond)
	if got := atomic.LoadInt32(&execCount); got != 2 {
		t.Fatalf("exec count after pending trigger = %d, want 2", got)
	}
}

func TestThrottleControllerReportsAsyncErrors(t *testing.T) {
	errCh := make(chan error, 1)
	tc := newThrottleController(20*time.Millisecond, func(ctx context.Context) error {
		return context.Canceled
	}, func(err error) {
		errCh <- err
	})

	if err := tc.Trigger(context.Background()); err != context.Canceled {
		t.Fatalf("first trigger error = %v, want context.Canceled", err)
	}
	if err := tc.Trigger(context.Background()); err != nil {
		t.Fatalf("second trigger: %v", err)
	}

	select {
	case err := <-errCh:
		if err != context.Canceled {
			t.Fatalf("async error = %v, want context.Canceled", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("timed out waiting for async error")
	}
}
