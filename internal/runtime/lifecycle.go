// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package runtime

import (
	"context"
	"fmt"

	"github.com/larksuite/channel-sdk-go/types"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
)

// OnReady registers a handler for WS ready events.
func (ch *Client) OnReady(handler func()) {
	ch.onReadyHandlers = append(ch.onReadyHandlers, handler)
}

// OnError registers a handler for WS error events.
func (ch *Client) OnError(handler func(err error)) {
	ch.onErrorHandlers = append(ch.onErrorHandlers, handler)
}

// OnReconnecting registers a handler for WS reconnecting events.
func (ch *Client) OnReconnecting(handler func()) {
	ch.onReconnectingHandlers = append(ch.onReconnectingHandlers, handler)
}

// OnReconnected registers a handler for WS reconnected events.
func (ch *Client) OnReconnected(handler func()) {
	ch.onReconnectedHandlers = append(ch.onReconnectedHandlers, handler)
}

// OnDisconnected registers a handler for WS disconnected events.
func (ch *Client) OnDisconnected(handler func()) {
	ch.onDisconnectedHandlers = append(ch.onDisconnectedHandlers, handler)
}

// UpdatePolicy updates the policy configuration for the channel.
func (ch *Client) UpdatePolicy(cfg types.PolicyConfig) {
	ch.policyGate.UpdateConfig(cfg)
}

// GetPolicy returns the current policy configuration.
func (ch *Client) GetPolicy() types.PolicyConfig {
	return ch.policyGate.GetConfig()
}

// Start starts the underlying WebSocket client and wires up lifecycle events.
func (ch *Client) Start(ctx context.Context) error {
	if ch.wsClient == nil {
		larkcore.NewEventLogger().Info(ctx, "[Channel] Start called but wsClient is nil, skipping WebSocket connection.")
		return nil
	}
	ch.wsClient.SetOnReady(func() {
		botInfo := ch.GetBotIdentity(ctx)
		botIdentityStr := ""
		if botInfo != nil {
			botIdentityStr = fmt.Sprintf("botIdentity: {\n  openId: '%s',\n  name: '%s'\n}", botInfo.OpenID, botInfo.Name)
		}

		larkcore.NewEventLogger().Info(ctx, fmt.Sprintf("receive events or callbacks through persistent connection only available in self-build & Feishu app, Configured in:\n"+
			"    Developer Console(开发者后台) \n"+
			"        ->\n"+
			"    Events and Callbacks(事件与回调)\n"+
			"        -> \n"+
			"    Mode of event/callback subscription(订阅方式)\n"+
			"        -> \n"+
			"    Receive events/callbacks through persistent connection(使用长连接接收事件/回调)\n\n"+
			"WebSocket 连接成功, %s", botIdentityStr))

		for _, h := range ch.onReadyHandlers {
			h()
		}
	})
	ch.wsClient.SetOnError(func(err error) {
		ch.emitError(ctx, err)
	})
	ch.wsClient.SetOnReconnecting(func() {
		for _, h := range ch.onReconnectingHandlers {
			h()
		}
	})
	ch.wsClient.SetOnReconnected(func() {
		for _, h := range ch.onReconnectedHandlers {
			h()
		}
	})
	ch.wsClient.SetOnDisconnected(func() {
		for _, h := range ch.onDisconnectedHandlers {
			h()
		}
	})
	return ch.wsClient.Start(ctx)
}

// Stop gracefully stops the underlying WebSocket client.
func (ch *Client) Stop(ctx context.Context) error {
	ch.stopOnce.Do(func() {
		if ch.wsClient != nil {
			ch.wsClient.Close()
		}
		if ch.pipelineManager != nil {
			ch.pipelineManager.Dispose()
		}
		if ch.processLock != nil {
			ch.processLock.Dispose()
		}
	})
	return nil
}
