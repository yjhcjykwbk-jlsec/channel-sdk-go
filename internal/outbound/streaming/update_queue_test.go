// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package streaming

import (
	"context"
	"testing"
	"time"
)

func TestUpdateQueueExecutesSubmittedWork(t *testing.T) {
	q := NewUpdateQueue(context.Background())
	defer q.Stop()

	done := make(chan struct{})
	if ok := q.Submit(func() {
		close(done)
	}); !ok {
		t.Fatalf("submit returned false")
	}

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("timed out waiting for submitted work")
	}
}

func TestUpdateQueueStopDrainsQueuedWork(t *testing.T) {
	q := NewUpdateQueue(context.Background())
	done := make(chan struct{})

	if ok := q.Submit(func() {
		close(done)
	}); !ok {
		t.Fatalf("submit returned false")
	}
	q.Stop()

	select {
	case <-done:
	default:
		t.Fatalf("queued work was not drained before stop returned")
	}
}

func TestUpdateQueueSubmitAfterStopDoesNotBlock(t *testing.T) {
	q := NewUpdateQueue(context.Background())
	q.Stop()

	done := make(chan bool, 1)
	go func() {
		done <- q.Submit(func() {})
	}()

	select {
	case ok := <-done:
		if ok {
			t.Fatalf("submit after stop returned true")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("submit after stop blocked")
	}
}
