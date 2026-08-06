// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package inbound

import (
	"context"
	"time"

	"github.com/larksuite/channel-sdk-go/internal/pipeline"
	"github.com/larksuite/channel-sdk-go/types"
)

type MessageHandler = func(ctx context.Context, msg *types.NormalizedMessage) error
type CommentHandler = func(ctx context.Context, event *types.CommentEvent) error
type ReactionHandler = func(ctx context.Context, event *types.ReactionEvent) error
type BotAddedHandler = func(ctx context.Context, event *types.BotAddedEvent) error
type CardActionHandler = func(ctx context.Context, event *types.CardActionEvent) error
type RejectHandler = func(ctx context.Context, event *types.RejectEvent) error

type HandlerRegistry interface {
	MessageHandlers() []MessageHandler
	CommentHandlers() []CommentHandler
	ReactionHandlers() []ReactionHandler
	BotAddedHandlers() []BotAddedHandler
	CardActionHandlers() []CardActionHandler
	RejectHandlers() []RejectHandler
}

type ErrorSink interface {
	EmitError(ctx context.Context, err error)
}

type ErrorSinkFunc func(ctx context.Context, err error)

func (f ErrorSinkFunc) EmitError(ctx context.Context, err error) {
	if f != nil {
		f(ctx, err)
	}
}

type BotIdentityProvider interface {
	GetBotIdentity(ctx context.Context) *types.BotIdentity
}

type Deduper interface {
	IsDuplicate(key string) bool
}

type ProcessingLocker interface {
	Acquire(id string) bool
	Release(id string)
}

type PipelineManager interface {
	Push(ctx context.Context, scope string, msg *types.NormalizedMessage, handler pipeline.FlushHandler)
	Run(ctx context.Context, scope string, task func() error) error
}

type PolicyEvaluator interface {
	Evaluate(msg *types.NormalizedMessage) types.PolicyDecision
}

type Dependencies struct {
	Handlers    HandlerRegistry
	Errors      ErrorSink
	BotIdentity BotIdentityProvider
	Dedup       Deduper
	Lock        ProcessingLocker
	Pipeline    PipelineManager
	Policy      PolicyEvaluator
	StaleWindow time.Duration
}
