// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package outbound

import (
	"context"
	"fmt"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

// SendResponse is the small response shape outbound needs from message APIs.
type SendResponse struct {
	MessageID string
	ChatID    string
}

// MessageDriver hides the Lark SDK request/response details from Sender.
type MessageDriver interface {
	CreateMessage(ctx context.Context, receiveIDType, receiveID, msgType, content string) (*SendResponse, error)
	ReplyMessage(ctx context.Context, replyMessageID, msgType, content string) (*SendResponse, error)
}

type LarkMessageDriver struct {
	client *lark.Client
}

func NewLarkMessageDriver(client *lark.Client) *LarkMessageDriver {
	return &LarkMessageDriver{client: client}
}

func (d *LarkMessageDriver) CreateMessage(ctx context.Context, receiveIDType, receiveID, msgType, content string) (*SendResponse, error) {
	req := larkim.NewCreateMessageReqBuilder().
		ReceiveIdType(receiveIDType).
		Body(larkim.NewCreateMessageReqBodyBuilder().
			ReceiveId(receiveID).
			MsgType(msgType).
			Content(content).
			Build()).
		Build()

	resp, err := d.client.Im.V1.Message.Create(ctx, req)
	if err != nil {
		return nil, err
	}
	if !resp.Success() {
		return nil, &larkcore.CodeError{Code: resp.Code, Msg: resp.Msg}
	}
	if resp.Data == nil || resp.Data.MessageId == nil {
		return nil, fmt.Errorf("create message failed: empty response data")
	}

	chatID := ""
	if resp.Data.ChatId != nil {
		chatID = *resp.Data.ChatId
	}
	return &SendResponse{MessageID: *resp.Data.MessageId, ChatID: chatID}, nil
}

func (d *LarkMessageDriver) ReplyMessage(ctx context.Context, replyMessageID, msgType, content string) (*SendResponse, error) {
	req := larkim.NewReplyMessageReqBuilder().
		MessageId(replyMessageID).
		Body(larkim.NewReplyMessageReqBodyBuilder().
			MsgType(msgType).
			Content(content).
			Build()).
		Build()

	resp, err := d.client.Im.V1.Message.Reply(ctx, req)
	if err != nil {
		return nil, err
	}
	if !resp.Success() {
		return nil, &larkcore.CodeError{Code: resp.Code, Msg: resp.Msg}
	}
	if resp.Data == nil || resp.Data.MessageId == nil {
		return nil, fmt.Errorf("reply message failed: empty response data")
	}

	chatID := ""
	if resp.Data.ChatId != nil {
		chatID = *resp.Data.ChatId
	}
	return &SendResponse{MessageID: *resp.Data.MessageId, ChatID: chatID}, nil
}
