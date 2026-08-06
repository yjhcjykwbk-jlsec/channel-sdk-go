// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package streaming

import (
	"context"
	"sync"
)

type UpdateQueue struct {
	ch     chan func()
	done   chan struct{}
	cancel context.CancelFunc

	mu     sync.Mutex
	closed bool
	wg     sync.WaitGroup
}

func NewUpdateQueue(ctx context.Context) *UpdateQueue {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithCancel(ctx)
	q := &UpdateQueue{
		ch:     make(chan func(), 100),
		done:   make(chan struct{}),
		cancel: cancel,
	}
	q.wg.Add(1)
	go q.loop()
	go func() {
		select {
		case <-ctx.Done():
			q.Stop()
		case <-q.done:
		}
	}()
	return q
}

func (q *UpdateQueue) loop() {
	defer q.wg.Done()
	for f := range q.ch {
		if f != nil {
			f()
		}
	}
}

func (q *UpdateQueue) Submit(f func()) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed {
		return false
	}
	select {
	case q.ch <- f:
		return true
	default:
		return false
	}
}

func (q *UpdateQueue) Stop() {
	q.mu.Lock()
	if q.closed {
		q.mu.Unlock()
		q.wg.Wait()
		return
	}
	q.closed = true
	close(q.done)
	close(q.ch)
	q.cancel()
	q.mu.Unlock()

	q.wg.Wait()
}
