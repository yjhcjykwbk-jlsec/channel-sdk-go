// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package streaming

import (
	"context"
	"fmt"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"

	"github.com/larksuite/channel-sdk-go/internal/outbound"
)

// Driver hides Lark message update APIs from stream controllers.
type Driver interface {
	UpdateMessage(ctx context.Context, messageID, msgType, content string) error
	ReplyMessage(ctx context.Context, messageID, msgType, content string) (*outbound.SendResponse, error)
	PatchMessage(ctx context.Context, messageID, content string) error
}

type LarkDriver struct {
	client *lark.Client
}

func NewLarkDriver(client *lark.Client) *LarkDriver {
	return &LarkDriver{client: client}
}

func (d *LarkDriver) UpdateMessage(ctx context.Context, messageID, msgType, content string) error {
	req := larkim.NewUpdateMessageReqBuilder().
		MessageId(messageID).
		Body(larkim.NewUpdateMessageReqBodyBuilder().
			MsgType(msgType).
			Content(content).
			Build()).
		Build()

	resp, err := d.client.Im.V1.Message.Update(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to update message: %w", err)
	}
	if !resp.Success() {
		return &larkcore.CodeError{Code: resp.Code, Msg: resp.Msg}
	}
	return nil
}

func (d *LarkDriver) ReplyMessage(ctx context.Context, messageID, msgType, content string) (*outbound.SendResponse, error) {
	req := larkim.NewReplyMessageReqBuilder().
		MessageId(messageID).
		Body(larkim.NewReplyMessageReqBodyBuilder().
			MsgType(msgType).
			Content(content).
			Build()).
		Build()

	resp, err := d.client.Im.V1.Message.Reply(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create new message chunk: %w", err)
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
	return &outbound.SendResponse{MessageID: *resp.Data.MessageId, ChatID: chatID}, nil
}

func (d *LarkDriver) PatchMessage(ctx context.Context, messageID, content string) error {
	req := larkim.NewPatchMessageReqBuilder().
		MessageId(messageID).
		Body(larkim.NewPatchMessageReqBodyBuilder().
			Content(content).
			Build()).
		Build()

	resp, err := d.client.Im.V1.Message.Patch(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to patch message: %w", err)
	}
	if !resp.Success() {
		return &larkcore.CodeError{Code: resp.Code, Msg: resp.Msg}
	}
	return nil
}
