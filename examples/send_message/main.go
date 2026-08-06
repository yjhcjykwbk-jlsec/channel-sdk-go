// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"log"
	"os"
	"time"

	channel "github.com/larksuite/channel-sdk-go"
)

func main() {
	ch, err := channel.New(os.Getenv("APP_ID"), os.Getenv("APP_SECRET"))
	if err != nil {
		log.Fatalf("create channel: %v", err)
	}
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := ch.Stop(stopCtx); err != nil {
			log.Printf("stop channel: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = ch.Send(ctx, &channel.SendInput{
		ChatID:  os.Getenv("CHAT_ID"),
		MsgType: "text",
		Text:    "hello from channel-sdk-go",
	})
	if err != nil {
		log.Fatalf("send message: %v", err)
	}
}
