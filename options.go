// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package channel

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/larksuite/channel-sdk-go/types"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkevent "github.com/larksuite/oapi-sdk-go/v3/event"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"
)

type config struct {
	clientOptions  []lark.ClientOptionFunc
	wsOptions      []larkws.ClientOption
	eventOptions   []larkevent.OptionFunc
	channelOptions []types.ChannelOption
	logger         larkcore.Logger
	hasAssertion   bool
}

type Option func(*config)

func WithLogger(logger larkcore.Logger) Option {
	return func(c *config) {
		c.logger = logger
		c.clientOptions = append(c.clientOptions, lark.WithLogger(logger))
		c.wsOptions = append(c.wsOptions, larkws.WithLogger(logger))
		c.eventOptions = append(c.eventOptions, larkevent.WithLogger(logger))
	}
}

func WithLogLevel(level larkcore.LogLevel) Option {
	return func(c *config) {
		c.clientOptions = append(c.clientOptions, lark.WithLogLevel(level))
		c.wsOptions = append(c.wsOptions, larkws.WithLogLevel(level))
		c.eventOptions = append(c.eventOptions, larkevent.WithLogLevel(level))
	}
}

func WithDomain(domain string) Option {
	return func(c *config) {
		c.clientOptions = append(c.clientOptions, lark.WithOpenBaseUrl(domain))
		c.wsOptions = append(c.wsOptions, larkws.WithDomain(domain))
	}
}

func WithOpenBaseURL(baseURL string) Option {
	return func(c *config) {
		c.clientOptions = append(c.clientOptions, lark.WithOpenBaseUrl(baseURL))
	}
}

func WithWebSocketDomain(domain string) Option {
	return func(c *config) {
		c.wsOptions = append(c.wsOptions, larkws.WithDomain(domain))
	}
}

func WithOAuthBaseURL(baseURL string) Option {
	return func(c *config) {
		c.clientOptions = append(c.clientOptions, lark.WithOAuthBaseUrl(baseURL))
	}
}

func WithHTTPClient(httpClient larkcore.HttpClient) Option {
	return func(c *config) {
		c.clientOptions = append(c.clientOptions, lark.WithHttpClient(httpClient))
	}
}

func WithReqTimeout(timeout time.Duration) Option {
	return func(c *config) {
		c.clientOptions = append(c.clientOptions, lark.WithReqTimeout(timeout))
	}
}

func WithTokenCache(cache larkcore.Cache) Option {
	return func(c *config) {
		c.clientOptions = append(c.clientOptions, lark.WithTokenCache(cache))
	}
}

func WithEnableTokenCache(enable bool) Option {
	return func(c *config) {
		c.clientOptions = append(c.clientOptions, lark.WithEnableTokenCache(enable))
	}
}

func WithHeaders(headers http.Header) Option {
	return func(c *config) {
		c.clientOptions = append(c.clientOptions, lark.WithHeaders(headers))
		c.wsOptions = append(c.wsOptions, larkws.WithHeaders(headers))
	}
}

func WithSource(source string) Option {
	return func(c *config) {
		c.clientOptions = append(c.clientOptions, lark.WithSource(source))
		c.wsOptions = append(c.wsOptions, larkws.WithSource(source))
	}
}

func WithClientAssertionProvider(provider larkcore.ClientAssertionProvider) Option {
	return func(c *config) {
		if provider != nil {
			c.hasAssertion = true
		}
		c.clientOptions = append(c.clientOptions, lark.WithClientAssertionProvider(provider))
		c.wsOptions = append(c.wsOptions, larkws.WithClientAssertionProvider(provider))
	}
}

func WithSafetyConfig(cfg types.SafetyConfig) Option {
	return func(c *config) {
		c.channelOptions = append(c.channelOptions, types.WithSafetyConfig(cfg))
	}
}

func WithPolicyConfig(cfg types.PolicyConfig) Option {
	return func(c *config) {
		c.channelOptions = append(c.channelOptions, types.WithPolicyConfig(cfg))
	}
}

func WithOutboundConfig(cfg types.OutboundConfig) Option {
	return func(c *config) {
		c.channelOptions = append(c.channelOptions, types.WithOutboundConfig(cfg))
	}
}

func WithBotIdentityCacheConfig(cfg types.BotIdentityCacheConfig) Option {
	return func(c *config) {
		c.channelOptions = append(c.channelOptions, types.WithBotIdentityCacheConfig(cfg))
	}
}

func validateCredentialConfig(appID, appSecret string, cfg config) error {
	if strings.TrimSpace(appID) == "" {
		return errors.New("appID is required")
	}
	if strings.TrimSpace(appSecret) == "" && !cfg.hasAssertion {
		return errors.New("appSecret is required when client assertion is not configured")
	}
	return nil
}
