// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"log"
	"os"

	channel "github.com/larksuite/channel-sdk-go"
)

func main() {
	ch, err := channel.New(os.Getenv("APP_ID"), os.Getenv("APP_SECRET"))
	if err != nil {
		log.Fatalf("create channel: %v", err)
	}

	ch.OnMessage(func(ctx context.Context, msg *channel.NormalizedMessage) error {
		_, err := ch.Send(ctx, &channel.SendInput{
			ReceiveID: msg.ChatID,
			MsgType:   "text",
			Text:      msg.Content,
		})
		return err
	})

	if err := ch.Start(context.Background()); err != nil {
		log.Fatalf("start channel: %v", err)
	}
}
