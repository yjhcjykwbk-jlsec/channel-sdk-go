// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	channel "github.com/larksuite/channel-sdk-go"
)

func main() {
	ch, err := channel.New(os.Getenv("APP_ID"), os.Getenv("APP_SECRET"))
	if err != nil {
		log.Fatalf("create channel: %v", err)
	}

	ch.OnCardAction(func(ctx context.Context, event *channel.CardActionEvent) error {
		_, err := ch.Send(ctx, &channel.SendInput{
			ChatID:  event.ChatID,
			MsgType: "text",
			Text:    fmt.Sprintf("handled card action: %s", event.Action.Name),
		})
		return err
	})

	if err := ch.Start(context.Background()); err != nil {
		log.Fatalf("start channel: %v", err)
	}
}
