// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"log"
	"os"

	channel "github.com/larksuite/channel-sdk-go"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
)

func main() {
	appID := os.Getenv("APP_ID")
	appSecret := os.Getenv("APP_SECRET")

	ch, err := channel.New(
		appID,
		appSecret,
		channel.WithLogLevel(larkcore.LogLevelInfo),
		channel.WithDomain("https://open.feishu-boe.cn"),
		channel.WithOAuthBaseURL("https://accounts.feishu-boe.cn"),
	)
	if err != nil {
		log.Fatalf("create channel: %v", err)
	}

	if err := ch.Start(context.Background()); err != nil {
		log.Fatalf("start channel: %v", err)
	}
}
