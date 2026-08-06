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
		stream, err := ch.Stream(ctx, &channel.SendInput{
			ChatID: msg.ChatID,
			Title:  "Channel SDK streaming reply",
		})
		if err != nil {
			return err
		}
		if err := stream.Append(ctx, "received: "); err != nil {
			return err
		}
		if err := stream.Append(ctx, msg.Content); err != nil {
			return err
		}
		if err := stream.Flush(ctx); err != nil {
			return err
		}
		return stream.Close(ctx)
	})

	if err := ch.Start(context.Background()); err != nil {
		log.Fatalf("start channel: %v", err)
	}
}
