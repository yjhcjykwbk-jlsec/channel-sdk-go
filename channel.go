// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package channel

import (
	runtime "github.com/larksuite/channel-sdk-go/internal/runtime"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkdispatcher "github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"
)

func New(appID, appSecret string, opts ...Option) (Channel, error) {
	cfg := config{}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	if err := validateCredentialConfig(appID, appSecret, cfg); err != nil {
		return nil, err
	}

	client := lark.NewClient(appID, appSecret, cfg.clientOptions...)

	eventDispatcher := larkdispatcher.NewEventDispatcher("", "")
	if len(cfg.eventOptions) > 0 {
		eventDispatcher.InitConfig(cfg.eventOptions...)
	}

	wsOptions := append([]larkws.ClientOption{
		larkws.WithEventHandler(eventDispatcher),
	}, cfg.wsOptions...)
	wsClient := larkws.NewClient(appID, appSecret, wsOptions...)

	return runtime.New(client, wsClient, cfg.logger, cfg.channelOptions...), nil
}
