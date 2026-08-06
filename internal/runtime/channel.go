// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package runtime

import (
	"context"
	"sync"
	"time"

	"github.com/larksuite/channel-sdk-go/internal/inbound"
	"github.com/larksuite/channel-sdk-go/internal/outbound"
	"github.com/larksuite/channel-sdk-go/internal/outbound/media"
	"github.com/larksuite/channel-sdk-go/internal/pipeline"
	"github.com/larksuite/channel-sdk-go/internal/safety"
	"github.com/larksuite/channel-sdk-go/types"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"
)

// Client is the default implementation of the Channel interface.
type Client struct {
	client          *lark.Client
	wsClient        *larkws.Client
	config          types.ChannelConfig
	sender          *outbound.Sender
	inbound         *inbound.Dispatcher
	dedupCache      *safety.DedupCache
	pipelineManager *pipeline.ChatPipelineManager
	policyGate      *safety.PolicyGate
	processLock     *safety.ProcessingLock
	staleWindow     time.Duration
	logger          larkcore.Logger
	stopOnce        sync.Once

	botIdentity              *types.BotIdentity
	botIdentityFetchedAt     time.Time
	botIdentityLastFailureAt time.Time
	botMu                    sync.Mutex

	// Handler registries
	onMessageHandlers    []func(ctx context.Context, msg *types.NormalizedMessage) error
	onCommentHandlers    []func(ctx context.Context, event *types.CommentEvent) error
	onReactionHandlers   []func(ctx context.Context, event *types.ReactionEvent) error
	onBotAddedHandlers   []func(ctx context.Context, event *types.BotAddedEvent) error
	onCardActionHandlers []func(ctx context.Context, event *types.CardActionEvent) error
	onRejectHandlers     []func(ctx context.Context, event *types.RejectEvent) error

	onReadyHandlers        []func()
	onErrorHandlers        []func(err error)
	onReconnectingHandlers []func()
	onReconnectedHandlers  []func()
	onDisconnectedHandlers []func()

	messageHandlerReg  bool
	commentHandlerReg  bool
	reactionHandlerReg bool
	botAddedHandlerReg bool
}

func New(client *lark.Client, wsClient *larkws.Client, logger larkcore.Logger, opts ...types.ChannelOption) *Client {
	cfg := types.DefaultChannelConfig()
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.BotIdentityCache.TTL <= 0 {
		cfg.BotIdentityCache.TTL = 30 * time.Minute
	}
	if cfg.BotIdentityCache.MinRefreshInterval <= 0 {
		cfg.BotIdentityCache.MinRefreshInterval = 1 * time.Minute
	} else if cfg.BotIdentityCache.MinRefreshInterval < 30*time.Second {
		cfg.BotIdentityCache.MinRefreshInterval = 30 * time.Second
	}
	if cfg.Outbound.MediaSource.MaxDownloadBytes <= 0 {
		cfg.Outbound.MediaSource.MaxDownloadBytes = types.DefaultMediaSourceMaxDownloadBytes
	}

	if logger == nil {
		logger = larkcore.NewEventLogger()
	}

	uploader := media.NewUploader(client)
	ch := &Client{
		client:          client,
		wsClient:        wsClient,
		config:          cfg,
		sender:          outbound.NewSender(outbound.NewLarkMessageDriver(client), outboundUploader{inner: uploader, config: cfg}, cfg, logger),
		dedupCache:      safety.NewDedupCache(cfg.Safety.Dedup.MaxEntries, cfg.Safety.Dedup.SweepIntervalMs),
		pipelineManager: pipeline.NewChatPipelineManager(cfg.Safety.Batch),
		policyGate:      safety.NewPolicyGate(&cfg.Policy, nil),
		processLock:     safety.NewProcessingLock(types.DefaultLockTTL, 1*time.Minute),
		staleWindow:     cfg.Safety.StaleMessageWindowMs,
		logger:          logger,
	}
	ch.inbound = inbound.NewDispatcher(inbound.Dependencies{
		Handlers:    ch,
		Errors:      inbound.ErrorSinkFunc(ch.emitError),
		BotIdentity: ch,
		Dedup:       ch.dedupCache,
		Lock:        ch.processLock,
		Pipeline:    ch.pipelineManager,
		Policy:      ch.policyGate,
		StaleWindow: ch.staleWindow,
	})
	ch.pipelineManager.SetErrorHandler(func(ctx context.Context, err error) {
		ch.emitError(ctx, err)
	})
	return ch
}

type outboundUploader struct {
	inner  media.Uploader
	config types.ChannelConfig
}

func (u outboundUploader) UploadImagePath(ctx context.Context, imageType string, imagePath string) (string, error) {
	return u.inner.UploadImagePath(ctx, imageType, imagePath)
}

func (u outboundUploader) UploadFilePath(ctx context.Context, fileType string, filePath string) (string, error) {
	return u.inner.UploadFilePath(ctx, fileType, filePath)
}

func (u outboundUploader) UploadMedia(ctx context.Context, input *types.UploadInput) (*types.UploadResult, error) {
	return u.inner.UploadMedia(ctx, input, &media.SourceOptions{
		SsrfGuard: &safety.SsrfGuardOptions{
			Allowlist: u.config.Outbound.MediaSource.URLAllowlist,
		},
		MaxDownloadSize: u.config.Outbound.MediaSource.MaxDownloadBytes,
	})
}

func (ch *Client) RawClient() *lark.Client {
	return ch.client
}

func (ch *Client) emitError(ctx context.Context, err error) {
	if err == nil {
		return
	}
	if ch.logger != nil {
		ch.logger.Warn(ctx, "[Channel] handler error")
	}
	for _, h := range ch.onErrorHandlers {
		h(err)
	}
}

func (ch *Client) logInfo(ctx context.Context, message string) {
	if ch.logger != nil {
		ch.logger.Info(ctx, message)
	}
}

func (ch *Client) logWarn(ctx context.Context, message string) {
	if ch.logger != nil {
		ch.logger.Warn(ctx, message)
	}
}
