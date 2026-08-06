// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package inbound

import (
	"context"
	"time"
)

type Dispatcher struct {
	handlers    HandlerRegistry
	errors      ErrorSink
	botIdentity BotIdentityProvider
	dedup       Deduper
	lock        ProcessingLocker
	pipeline    PipelineManager
	policy      PolicyEvaluator
	staleWindow time.Duration
}

func NewDispatcher(deps Dependencies) *Dispatcher {
	return &Dispatcher{
		handlers:    deps.Handlers,
		errors:      deps.Errors,
		botIdentity: deps.BotIdentity,
		dedup:       deps.Dedup,
		lock:        deps.Lock,
		pipeline:    deps.Pipeline,
		policy:      deps.Policy,
		staleWindow: deps.StaleWindow,
	}
}

func (d *Dispatcher) emitError(ctx context.Context, err error) {
	if err == nil || d.errors == nil {
		return
	}
	d.errors.EmitError(ctx, err)
}
