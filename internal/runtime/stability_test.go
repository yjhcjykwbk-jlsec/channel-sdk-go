// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package runtime

import (
	"context"
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/larksuite/channel-sdk-go/types"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"
)

type captureLogger struct {
	mu      sync.Mutex
	entries []string
}

func (l *captureLogger) Debug(ctx context.Context, args ...interface{}) {
	l.append(args...)
}

func (l *captureLogger) Info(ctx context.Context, args ...interface{}) {
	l.append(args...)
}

func (l *captureLogger) Warn(ctx context.Context, args ...interface{}) {
	l.append(args...)
}

func (l *captureLogger) Error(ctx context.Context, args ...interface{}) {
	l.append(args...)
}

func (l *captureLogger) append(args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	parts := make([]string, 0, len(args))
	for _, arg := range args {
		parts = append(parts, fmt.Sprint(arg))
	}
	l.entries = append(l.entries, strings.Join(parts, " "))
}

func (l *captureLogger) String() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return strings.Join(l.entries, "\n")
}

func newTestMessageChannel(t *testing.T) (*dispatcher.EventDispatcher, Channel, *channelImpl) {
	t.Helper()

	d := dispatcher.NewEventDispatcher("", "")
	wsCli := larkws.NewClient("appID", "appSecret", larkws.WithEventHandler(d))
	cli := lark.NewClient("appID", "appSecret")
	requireMention := false
	safetyCfg := types.DefaultChannelConfig().Safety
	safetyCfg.Batch.DelayMs = time.Millisecond
	safetyCfg.Batch.LongDelayMs = time.Millisecond

	ch := newChannel(
		cli,
		wsCli,
		types.WithSafetyConfig(safetyCfg),
		types.WithPolicyConfig(types.PolicyConfig{
			RequireMention: &requireMention,
		}),
	)
	impl := ch.(*channelImpl)
	impl.botIdentity = &types.BotIdentity{
		OpenID: "ou_bot",
		Name:   "test-bot",
	}
	return d, ch, impl
}

func TestOnMessageHandlerErrorEmitsOnErrorAndReleasesLock(t *testing.T) {
	d, ch, impl := newTestMessageChannel(t)

	handlerErr := errors.New("handler failed")
	errCh := make(chan error, 1)
	ch.OnError(func(err error) {
		errCh <- err
	})
	ch.OnMessage(func(ctx context.Context, msg *types.NormalizedMessage) error {
		return handlerErr
	})

	triggerMessage(d, "evt_error_1", "om_error_1", time.Now().UnixMilli())

	select {
	case err := <-errCh:
		if !errors.Is(err, handlerErr) {
			t.Fatalf("expected handler error, got %v", err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected OnError to receive message handler error")
	}

	if !impl.processLock.Acquire("om_error_1") {
		t.Fatal("process lock should be released after handler error")
	}
}

func TestOnRejectHandlerErrorEmitsOnError(t *testing.T) {
	d := dispatcher.NewEventDispatcher("", "")
	wsCli := larkws.NewClient("appID", "appSecret", larkws.WithEventHandler(d))
	cli := lark.NewClient("appID", "appSecret")
	requireMention := true
	safetyCfg := types.DefaultChannelConfig().Safety
	safetyCfg.Batch.DelayMs = time.Millisecond
	safetyCfg.Batch.LongDelayMs = time.Millisecond

	ch := newChannel(
		cli,
		wsCli,
		types.WithSafetyConfig(safetyCfg),
		types.WithPolicyConfig(types.PolicyConfig{
			RequireMention: &requireMention,
		}),
	)
	impl := ch.(*channelImpl)
	impl.botIdentity = &types.BotIdentity{
		OpenID: "ou_bot",
		Name:   "test-bot",
	}

	handlerErr := errors.New("reject handler failed")
	errCh := make(chan error, 1)
	ch.OnError(func(err error) {
		errCh <- err
	})
	ch.OnMessage(func(ctx context.Context, msg *types.NormalizedMessage) error {
		return nil
	})
	ch.OnReject(func(ctx context.Context, event *types.RejectEvent) error {
		return handlerErr
	})

	triggerMessage(d, "evt_reject_error_1", "om_reject_error_1", time.Now().UnixMilli())

	select {
	case err := <-errCh:
		if !errors.Is(err, handlerErr) {
			t.Fatalf("expected reject handler error, got %v", err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected OnError to receive reject handler error")
	}
}

func TestStopIsIdempotentAndReleasesLocalLocks(t *testing.T) {
	_, ch, impl := newTestMessageChannel(t)

	if !impl.processLock.Acquire("resource") {
		t.Fatal("expected lock acquisition")
	}
	if err := ch.Stop(context.Background()); err != nil {
		t.Fatalf("first Stop returned error: %v", err)
	}
	if err := ch.Stop(context.Background()); err != nil {
		t.Fatalf("second Stop returned error: %v", err)
	}
	if !impl.processLock.Acquire("resource") {
		t.Fatal("Stop should clear local processing locks")
	}
}

func TestSendDoesNotUseStandardLogImport(t *testing.T) {
	src, err := os.ReadFile("send.go")
	if err != nil {
		t.Fatalf("read send.go: %v", err)
	}
	parsed, err := parser.ParseFile(token.NewFileSet(), "send.go", src, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parse send.go: %v", err)
	}
	for _, imp := range parsed.Imports {
		if imp.Path.Value == `"log"` {
			t.Fatal("send.go must not import standard library log")
		}
	}
}
