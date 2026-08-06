// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package inbound

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/larksuite/channel-sdk-go/internal/pipeline"
	"github.com/larksuite/channel-sdk-go/internal/safety"
	"github.com/larksuite/channel-sdk-go/types"
	larkevent "github.com/larksuite/oapi-sdk-go/v3/event"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

type fakeHandlers struct {
	messageHandlers []MessageHandler
	rejectHandlers  []RejectHandler
}

func (f *fakeHandlers) MessageHandlers() []MessageHandler       { return f.messageHandlers }
func (f *fakeHandlers) CommentHandlers() []CommentHandler       { return nil }
func (f *fakeHandlers) ReactionHandlers() []ReactionHandler     { return nil }
func (f *fakeHandlers) BotAddedHandlers() []BotAddedHandler     { return nil }
func (f *fakeHandlers) CardActionHandlers() []CardActionHandler { return nil }
func (f *fakeHandlers) RejectHandlers() []RejectHandler         { return f.rejectHandlers }

type fakeBotIdentity struct {
	bot *types.BotIdentity
}

func (f fakeBotIdentity) GetBotIdentity(ctx context.Context) *types.BotIdentity {
	return f.bot
}

type fakePolicy struct {
	decision types.PolicyDecision
}

func (f fakePolicy) Evaluate(msg *types.NormalizedMessage) types.PolicyDecision {
	return f.decision
}

type immediatePipeline struct {
	errs *[]error
}

func (p immediatePipeline) Push(ctx context.Context, scope string, msg *types.NormalizedMessage, handler pipeline.FlushHandler) {
	if err := handler(ctx, &types.BatchedDispatch{Message: msg, SourceIDs: []string{msg.MessageID}}); err != nil && p.errs != nil {
		*p.errs = append(*p.errs, err)
	}
}

func (p immediatePipeline) Run(ctx context.Context, scope string, task func() error) error {
	return task()
}

type collectingErrors struct {
	errs []error
}

func (c *collectingErrors) EmitError(ctx context.Context, err error) {
	if err != nil {
		c.errs = append(c.errs, err)
	}
}

func TestHandleMessageIgnoresSelfMessage(t *testing.T) {
	var handled int
	dispatcher := NewDispatcher(Dependencies{
		Handlers: &fakeHandlers{
			messageHandlers: []MessageHandler{func(ctx context.Context, msg *types.NormalizedMessage) error {
				handled++
				return nil
			}},
		},
		BotIdentity: fakeBotIdentity{bot: &types.BotIdentity{OpenID: "ou_bot"}},
		Pipeline:    immediatePipeline{},
		Policy:      fakePolicy{decision: types.PolicyDecision{Allowed: true}},
		StaleWindow: time.Hour,
	})

	if err := dispatcher.HandleMessage(context.Background(), newMessageEvent("evt_1", "om_1", "ou_bot", nil)); err != nil {
		t.Fatalf("handle message: %v", err)
	}
	if handled != 0 {
		t.Fatalf("self message should be ignored, handled %d times", handled)
	}
}

func TestHandleMessageMarksMentionedBot(t *testing.T) {
	botOpenID := "ou_bot"
	mentionName := "bot"
	var got *types.NormalizedMessage
	dispatcher := NewDispatcher(Dependencies{
		Handlers: &fakeHandlers{
			messageHandlers: []MessageHandler{func(ctx context.Context, msg *types.NormalizedMessage) error {
				got = msg
				return nil
			}},
		},
		BotIdentity: fakeBotIdentity{bot: &types.BotIdentity{OpenID: botOpenID}},
		Pipeline:    immediatePipeline{},
		Policy:      fakePolicy{decision: types.PolicyDecision{Allowed: true}},
		StaleWindow: time.Hour,
	})

	event := newMessageEvent("evt_1", "om_1", "ou_sender", []*larkim.MentionEvent{
		{Key: ptr("@_user_1"), Id: &larkim.UserId{OpenId: &botOpenID}, Name: &mentionName},
	})
	if err := dispatcher.HandleMessage(context.Background(), event); err != nil {
		t.Fatalf("handle message: %v", err)
	}
	if got == nil {
		t.Fatalf("expected message handler to run")
	}
	if !got.MentionedBot {
		t.Fatalf("expected MentionedBot to be true")
	}
	if len(got.Mentions) != 1 || !got.Mentions[0].IsBot {
		t.Fatalf("expected mention to be marked as bot: %+v", got.Mentions)
	}
}

func TestHandleMessageEmitsRejectEvent(t *testing.T) {
	rejectCh := make(chan *types.RejectEvent, 1)
	dispatcher := NewDispatcher(Dependencies{
		Handlers: &fakeHandlers{
			messageHandlers: []MessageHandler{func(ctx context.Context, msg *types.NormalizedMessage) error {
				return nil
			}},
			rejectHandlers: []RejectHandler{func(ctx context.Context, event *types.RejectEvent) error {
				rejectCh <- event
				return nil
			}},
		},
		BotIdentity: fakeBotIdentity{bot: &types.BotIdentity{OpenID: "ou_bot"}},
		Pipeline:    immediatePipeline{},
		Policy:      fakePolicy{decision: types.PolicyDecision{Allowed: false, Reason: types.RejectReasonNoMention}},
		StaleWindow: time.Hour,
	})

	if err := dispatcher.HandleMessage(context.Background(), newMessageEvent("evt_1", "om_1", "ou_sender", nil)); err != nil {
		t.Fatalf("handle message: %v", err)
	}
	select {
	case event := <-rejectCh:
		if event.MessageID != "om_1" || event.Reason != string(types.RejectReasonNoMention) {
			t.Fatalf("unexpected reject event: %+v", event)
		}
	default:
		t.Fatalf("expected reject event")
	}
}

func TestHandleMessageReleasesLockOnHandlerError(t *testing.T) {
	lock := safety.NewProcessingLock(time.Minute, time.Minute)
	defer lock.Dispose()
	pipelineErrs := []error{}
	dispatcher := NewDispatcher(Dependencies{
		Handlers: &fakeHandlers{
			messageHandlers: []MessageHandler{func(ctx context.Context, msg *types.NormalizedMessage) error {
				return errors.New("handler failed")
			}},
		},
		BotIdentity: fakeBotIdentity{bot: &types.BotIdentity{OpenID: "ou_bot"}},
		Lock:        lock,
		Pipeline:    immediatePipeline{errs: &pipelineErrs},
		Policy:      fakePolicy{decision: types.PolicyDecision{Allowed: true}},
		StaleWindow: time.Hour,
	})

	if err := dispatcher.HandleMessage(context.Background(), newMessageEvent("evt_1", "om_1", "ou_sender", nil)); err != nil {
		t.Fatalf("handle message: %v", err)
	}
	if !lock.Acquire("om_1") {
		t.Fatalf("expected message lock to be released after handler error")
	}
	lock.Release("om_1")
	if len(pipelineErrs) != 1 {
		t.Fatalf("expected pipeline to receive handler error once, got %d", len(pipelineErrs))
	}
}

func newMessageEvent(eventID, messageID, senderOpenID string, mentions []*larkim.MentionEvent) *larkim.P2MessageReceiveV1 {
	content := `{"text":"hello"}`
	msgType := "text"
	chatID := "oc_chat"
	chatType := "group"
	createTime := strconv.FormatInt(time.Now().UnixMilli(), 10)
	return &larkim.P2MessageReceiveV1{
		EventV2Base: &larkevent.EventV2Base{
			Schema: "2.0",
			Header: &larkevent.EventHeader{
				EventID:    eventID,
				EventType:  "im.message.receive_v1",
				CreateTime: createTime,
			},
		},
		Event: &larkim.P2MessageReceiveV1Data{
			Sender: &larkim.EventSender{
				SenderId: &larkim.UserId{OpenId: &senderOpenID},
			},
			Message: &larkim.EventMessage{
				MessageId:   &messageID,
				ChatId:      &chatID,
				ChatType:    &chatType,
				MessageType: &msgType,
				Content:     &content,
				Mentions:    mentions,
			},
		},
	}
}

func ptr(s string) *string {
	return &s
}
