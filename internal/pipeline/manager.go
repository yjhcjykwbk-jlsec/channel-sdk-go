// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package pipeline

import (
	"context"
	"sync"

	"github.com/larksuite/channel-sdk-go/types"
)

// ChatPipelineManager manages multiple chat pipelines by scope.
type ChatPipelineManager struct {
	mu        sync.RWMutex
	config    types.BatchConfig
	pipelines map[string]*ChatPipeline
	onError   ErrorHandler
	closed    bool
}

func NewChatPipelineManager(config types.BatchConfig) *ChatPipelineManager {
	return &ChatPipelineManager{
		config:    config,
		pipelines: make(map[string]*ChatPipeline),
	}
}

func (cpm *ChatPipelineManager) getOrCreate(scope string, serialOnly bool) *ChatPipeline {
	cpm.mu.RLock()
	if cpm.closed {
		cpm.mu.RUnlock()
		return nil
	}
	p, ok := cpm.pipelines[scope]
	cpm.mu.RUnlock()

	if ok {
		return p
	}

	cpm.mu.Lock()
	defer cpm.mu.Unlock()
	if cpm.closed {
		return nil
	}

	// Double check
	p, ok = cpm.pipelines[scope]
	if ok {
		return p
	}

	p = NewChatPipeline(cpm.config, serialOnly)
	p.SetErrorHandler(cpm.onError)
	cpm.pipelines[scope] = p
	return p
}

func (cpm *ChatPipelineManager) Push(ctx context.Context, scope string, msg *types.NormalizedMessage, handler FlushHandler) {
	p := cpm.getOrCreate(scope, false)
	if p == nil {
		return
	}
	p.Push(ctx, msg, handler)
}

func (cpm *ChatPipelineManager) Run(ctx context.Context, scope string, task func() error) error {
	p := cpm.getOrCreate(scope, true)
	if p == nil {
		return errChatPipelineClosed
	}
	return p.Run(ctx, task)
}

func (cpm *ChatPipelineManager) SetErrorHandler(handler ErrorHandler) {
	cpm.mu.Lock()
	defer cpm.mu.Unlock()
	cpm.onError = handler
	for _, p := range cpm.pipelines {
		p.SetErrorHandler(handler)
	}
}

func (cpm *ChatPipelineManager) FlushAll(ctx context.Context) {
	cpm.mu.RLock()
	pipelines := make([]*ChatPipeline, 0, len(cpm.pipelines))
	for _, p := range cpm.pipelines {
		pipelines = append(pipelines, p)
	}
	cpm.mu.RUnlock()

	var wg sync.WaitGroup
	for _, p := range pipelines {
		wg.Add(1)
		go func(pipeline *ChatPipeline) {
			defer wg.Done()
			pipeline.FlushNow(ctx)
		}(p)
	}
	wg.Wait()
}

func (cpm *ChatPipelineManager) Dispose() {
	cpm.mu.Lock()
	if cpm.closed {
		cpm.mu.Unlock()
		return
	}
	cpm.closed = true
	pipelines := make([]*ChatPipeline, 0, len(cpm.pipelines))
	for _, p := range cpm.pipelines {
		pipelines = append(pipelines, p)
	}
	cpm.pipelines = make(map[string]*ChatPipeline)
	cpm.mu.Unlock()

	for _, p := range pipelines {
		p.FlushNow(context.Background())
		p.Dispose()
	}
}
