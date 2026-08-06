// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package outbound

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/larksuite/channel-sdk-go/types"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
)

type messageCall struct {
	kind          string
	receiveIDType string
	receiveID     string
	replyID       string
	msgType       string
	content       string
}

type fakeDriverResult struct {
	resp *SendResponse
	err  error
}

type fakeMessageDriver struct {
	createCalls   []messageCall
	replyCalls    []messageCall
	createResults []fakeDriverResult
	replyResults  []fakeDriverResult
}

func (f *fakeMessageDriver) CreateMessage(ctx context.Context, receiveIDType, receiveID, msgType, content string) (*SendResponse, error) {
	f.createCalls = append(f.createCalls, messageCall{
		kind:          "create",
		receiveIDType: receiveIDType,
		receiveID:     receiveID,
		msgType:       msgType,
		content:       content,
	})
	if len(f.createResults) == 0 {
		return &SendResponse{MessageID: "om_default", ChatID: "oc_default"}, nil
	}
	res := f.createResults[0]
	f.createResults = f.createResults[1:]
	return res.resp, res.err
}

func (f *fakeMessageDriver) ReplyMessage(ctx context.Context, replyMessageID, msgType, content string) (*SendResponse, error) {
	f.replyCalls = append(f.replyCalls, messageCall{
		kind:    "reply",
		replyID: replyMessageID,
		msgType: msgType,
		content: content,
	})
	if len(f.replyResults) == 0 {
		return &SendResponse{MessageID: "om_reply_default", ChatID: "oc_reply_default"}, nil
	}
	res := f.replyResults[0]
	f.replyResults = f.replyResults[1:]
	return res.resp, res.err
}

func testSenderConfig() types.ChannelConfig {
	cfg := types.DefaultChannelConfig()
	cfg.Outbound.Retry.MaxAttempts = 3
	cfg.Outbound.Retry.BaseDelayMs = time.Millisecond
	cfg.Outbound.TextChunkLimit = 3500
	return cfg
}

func TestSenderSendTextCreate(t *testing.T) {
	driver := &fakeMessageDriver{
		createResults: []fakeDriverResult{{resp: &SendResponse{MessageID: "om_1", ChatID: "oc_1"}}},
	}
	sender := NewSender(driver, nil, testSenderConfig(), nil)

	res, err := sender.Send(context.Background(), &types.SendInput{
		ReceiveID: "ou_123",
		Text:      "hello",
	})
	if err != nil {
		t.Fatalf("send text: %v", err)
	}
	if res.MessageID != "om_1" || res.ChatID != "oc_1" {
		t.Fatalf("unexpected result: %+v", res)
	}
	if len(driver.createCalls) != 1 {
		t.Fatalf("expected 1 create call, got %d", len(driver.createCalls))
	}
	call := driver.createCalls[0]
	if call.receiveIDType != "open_id" || call.receiveID != "ou_123" || call.msgType != "text" {
		t.Fatalf("unexpected create call: %+v", call)
	}
	var content map[string]string
	if err := json.Unmarshal([]byte(call.content), &content); err != nil {
		t.Fatalf("unmarshal text content: %v", err)
	}
	if content["text"] != "hello" {
		t.Fatalf("expected text content, got %#v", content)
	}
}

func TestSenderReplyTargetRevokedFallsBackToCreate(t *testing.T) {
	driver := &fakeMessageDriver{
		replyResults: []fakeDriverResult{{err: &larkcore.CodeError{Code: 230011, Msg: "message revoked"}}},
		createResults: []fakeDriverResult{{
			resp: &SendResponse{MessageID: "om_new", ChatID: "oc_new"},
		}},
	}
	sender := NewSender(driver, nil, testSenderConfig(), nil)
	input := &types.SendInput{
		ReceiveID:      "ou_123",
		ReplyMessageID: "om_old",
		Text:           "hello",
	}

	res, err := sender.Send(context.Background(), input)
	if err != nil {
		t.Fatalf("send reply fallback: %v", err)
	}
	if res.MessageID != "om_new" {
		t.Fatalf("expected fallback create message id, got %+v", res)
	}
	if len(driver.replyCalls) != 1 || len(driver.createCalls) != 1 {
		t.Fatalf("expected one reply and one create, got replies=%d creates=%d", len(driver.replyCalls), len(driver.createCalls))
	}
	if input.ReplyMessageID != "" {
		t.Fatalf("expected reply message id to be cleared after fallback, got %q", input.ReplyMessageID)
	}
}

func TestSenderFormatErrorFallsBackToText(t *testing.T) {
	driver := &fakeMessageDriver{
		createResults: []fakeDriverResult{
			{err: &larkcore.CodeError{Code: 230001, Msg: "invalid content"}},
			{resp: &SendResponse{MessageID: "om_text", ChatID: "oc_1"}},
		},
	}
	sender := NewSender(driver, nil, testSenderConfig(), nil)

	res, err := sender.Send(context.Background(), &types.SendInput{
		ReceiveID: "ou_123",
		Markdown:  "hello **markdown**",
	})
	if err != nil {
		t.Fatalf("send format fallback: %v", err)
	}
	if res.MessageID != "om_text" {
		t.Fatalf("expected fallback text result, got %+v", res)
	}
	if len(driver.createCalls) != 2 {
		t.Fatalf("expected 2 create calls, got %d", len(driver.createCalls))
	}
	if driver.createCalls[0].msgType != "post" || driver.createCalls[1].msgType != "text" {
		t.Fatalf("expected post then text fallback, got %+v", driver.createCalls)
	}
}

func TestSenderRetriesRetryableCreateErrors(t *testing.T) {
	driver := &fakeMessageDriver{
		createResults: []fakeDriverResult{
			{err: errors.New("temporary failure")},
			{err: errors.New("temporary failure")},
			{resp: &SendResponse{MessageID: "om_retry", ChatID: "oc_retry"}},
		},
	}
	sender := NewSender(driver, nil, testSenderConfig(), nil)

	res, err := sender.Send(context.Background(), &types.SendInput{
		ReceiveID: "ou_123",
		Text:      "hello",
	})
	if err != nil {
		t.Fatalf("send retry: %v", err)
	}
	if res.MessageID != "om_retry" {
		t.Fatalf("expected retry result, got %+v", res)
	}
	if len(driver.createCalls) != 3 {
		t.Fatalf("expected 3 create attempts, got %d", len(driver.createCalls))
	}
}

func TestSenderNilInput(t *testing.T) {
	sender := NewSender(&fakeMessageDriver{}, nil, testSenderConfig(), nil)

	_, err := sender.Send(context.Background(), nil)
	if err == nil {
		t.Fatal("expected nil input error")
	}
}
