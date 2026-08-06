// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/larksuite/channel-sdk-go/types"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
)

// GetBotIdentity fetches and caches the bot's identity from the server.
func (ch *Client) GetBotIdentity(ctx context.Context) *types.BotIdentity {
	if ch.cachedBotIdentityFresh() {
		return ch.botIdentity
	}

	ch.botMu.Lock()
	defer ch.botMu.Unlock()
	if ch.cachedBotIdentityFresh() {
		return ch.botIdentity
	}

	now := time.Now()
	if ch.shouldThrottleBotIdentityRefresh(now) {
		return ch.botIdentity
	}

	identity, err := ch.fetchBotIdentity(ctx)
	if err != nil {
		ch.botIdentityLastFailureAt = now
		if ch.botIdentity != nil {
			larkcore.NewEventLogger().Warn(ctx, fmt.Sprintf("[Channel] Failed to refresh bot info, using stale cache. Err: %v", err))
			return ch.botIdentity
		}
		larkcore.NewEventLogger().Error(ctx, fmt.Sprintf("[Channel] Failed to fetch bot info. Err: %v", err))
		return nil
	}

	ch.botIdentity = identity
	ch.botIdentityFetchedAt = time.Now()
	ch.botIdentityLastFailureAt = time.Time{}
	return ch.botIdentity
}

func (ch *Client) cachedBotIdentityFresh() bool {
	if ch.botIdentity == nil {
		return false
	}
	ttl := ch.config.BotIdentityCache.TTL
	if ttl <= 0 {
		return true
	}
	return time.Since(ch.botIdentityFetchedAt) < ttl
}

func (ch *Client) shouldThrottleBotIdentityRefresh(now time.Time) bool {
	if ch.botIdentityLastFailureAt.IsZero() {
		return false
	}
	minInterval := ch.config.BotIdentityCache.MinRefreshInterval
	if minInterval <= 0 {
		return false
	}
	return now.Sub(ch.botIdentityLastFailureAt) < minInterval
}

func (ch *Client) fetchBotIdentity(ctx context.Context) (*types.BotIdentity, error) {
	// Fetch bot info using the official SDK raw method since bot/v3/info is not generated.
	resp, err := ch.client.Get(ctx, "/open-apis/bot/v3/info", nil, larkcore.AccessTokenTypeTenant)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, fmt.Errorf("empty response from bot/v3/info")
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code from bot/v3/info: %d", resp.StatusCode)
	}
	larkcore.NewEventLogger().Debug(ctx, fmt.Sprintf("[Channel] bot/v3/info response received. status=%d raw_body_bytes=%d", resp.StatusCode, len(resp.RawBody)))

	var result struct {
		Code int `json:"code"`
		Bot  struct {
			OpenId         string `json:"open_id"`
			AppName        string `json:"app_name"`
			ActivateStatus int    `json:"activate_status"`
		} `json:"bot"`
	}
	if err := json.Unmarshal(resp.RawBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse bot info: %w", err)
	}
	larkcore.NewEventLogger().Debug(ctx, fmt.Sprintf("[Channel] Parsed bot/v3/info payload. code=%d open_id=%q app_name=%q activate_status=%d", result.Code, result.Bot.OpenId, result.Bot.AppName, result.Bot.ActivateStatus))
	if result.Code != 0 {
		return nil, fmt.Errorf("bot info returned non-zero code: %d", result.Code)
	}

	botName := result.Bot.AppName
	if botName == "" {
		larkcore.NewEventLogger().Debug(ctx, fmt.Sprintf("[Channel] bot/v3/info app_name is empty, using default name. open_id=%q activate_status=%d", result.Bot.OpenId, result.Bot.ActivateStatus))
		botName = "bot"
	}

	identity := &types.BotIdentity{
		OpenID:         result.Bot.OpenId,
		Name:           botName,
		ActivateStatus: result.Bot.ActivateStatus,
	}
	larkcore.NewEventLogger().Debug(ctx, fmt.Sprintf("[Channel] Resolved bot identity. open_id=%q name=%q activate_status=%d", identity.OpenID, identity.Name, identity.ActivateStatus))
	return identity, nil
}
