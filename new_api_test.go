// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package channel

import (
	"context"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
)

type mockAssertionProvider struct{}

func (mockAssertionProvider) RetrieveToken(ctx context.Context, aud string) (*larkcore.Token, error) {
	if aud == "" {
		return nil, errors.New("aud is required")
	}
	return &larkcore.Token{Value: "assertion"}, nil
}

type mockCache struct{}

func (mockCache) Set(ctx context.Context, key string, value string, expireTime time.Duration) error {
	return nil
}

func (mockCache) Get(ctx context.Context, key string) (string, error) {
	return "", nil
}

type mockLogger struct{}

func (mockLogger) Debug(context.Context, ...interface{}) {}
func (mockLogger) Info(context.Context, ...interface{})  {}
func (mockLogger) Warn(context.Context, ...interface{})  {}
func (mockLogger) Error(context.Context, ...interface{}) {}

func TestNewRejectsEmptyAppID(t *testing.T) {
	_, err := New("", "secret")
	if err == nil {
		t.Fatal("expected appID error")
	}
}

func TestNewRejectsEmptyAppSecretWithoutAssertion(t *testing.T) {
	_, err := New("cli_a", "")
	if err == nil {
		t.Fatal("expected appSecret error")
	}
}

func TestNewAcceptsClientAssertionWithoutAppSecret(t *testing.T) {
	ch, err := New("cli_a", "", WithClientAssertionProvider(mockAssertionProvider{}))
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	if ch == nil {
		t.Fatal("New returned nil channel")
	}
}

func TestNewExposesRawClient(t *testing.T) {
	ch, err := New("cli_a", "secret")
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	if ch.RawClient() == nil {
		t.Fatal("raw client should be available")
	}
	if _, ok := any(ch).(interface{ RawWSClient() any }); ok {
		t.Fatal("RawWSClient must not be exposed")
	}
}

func TestNewAllowsHandlerRegistration(t *testing.T) {
	ch, err := New("cli_a", "secret")
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	ch.OnReady(func() {})
	ch.OnError(func(err error) {})
	ch.OnMessage(func(ctx context.Context, msg *NormalizedMessage) error {
		return nil
	})
}

func TestRootTypeAliasesCompile(t *testing.T) {
	var _ Channel = nil
	var _ *SendInput = &SendInput{}
	var _ *SendResult = &SendResult{}
	var _ *NormalizedMessage = &NormalizedMessage{}
	var _ *ReactionEvent = &ReactionEvent{}
	var _ *CommentEvent = &CommentEvent{}
	var _ *CardActionEvent = &CardActionEvent{}
	var _ *RejectEvent = &RejectEvent{}

	input := SendInput{
		ReceiveID: "oc_test",
		MsgType:   "text",
		Text:      "hello",
	}
	if input.ReceiveID == "" {
		t.Fatal("ReceiveID should be set")
	}
	input.Media = &UploadInput{Kind: MediaKindImage, SourcePath: "test.png"}

	rateErr := &FeishuChannelError{Code: ErrCodeRateLimited}
	if ClassifyError(rateErr) == nil || !IsRetryable(rateErr) {
		t.Fatal("error helpers should be usable from root package")
	}
	if !IsFormatError(&FeishuChannelError{Code: ErrCodeFormatError}) {
		t.Fatal("IsFormatError should be usable from root package")
	}

	msg := NormalizedMessage{
		MessageID: "om_test",
		ChatID:    "oc_test",
	}
	if msg.MessageID == "" {
		t.Fatal("MessageID should be set")
	}

	evt := CardActionEvent{
		Action: CardActionPayload{Name: "approve"},
	}
	if evt.Action.Name == "" {
		t.Fatal("CardActionEvent should be usable from root package")
	}
}

func TestRootChannelOptionsCompile(t *testing.T) {
	_, err := New("cli_a", "secret", WithSafetyConfig(SafetyConfig{}))
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
}

func TestOptionsMapToClientConfig(t *testing.T) {
	mockHTTP := &MockHttpClient{}
	cache := mockCache{}
	logger := mockLogger{}
	headers := http.Header{"X-Test": []string{"channel"}}
	provider := mockAssertionProvider{}
	timeout := 7 * time.Second

	cfg := config{}
	for _, opt := range []Option{
		WithLogger(logger),
		WithLogLevel(larkcore.LogLevelWarn),
		WithDomain("https://open.example.com"),
		WithOAuthBaseURL("https://accounts.example.com"),
		WithHTTPClient(mockHTTP),
		WithReqTimeout(timeout),
		WithTokenCache(cache),
		WithEnableTokenCache(false),
		WithHeaders(headers),
		WithSource("channel-sdk-go-test"),
		WithClientAssertionProvider(provider),
	} {
		opt(&cfg)
	}

	clientConfig := &larkcore.Config{}
	for _, opt := range cfg.clientOptions {
		opt(clientConfig)
	}

	if clientConfig.Logger != logger {
		t.Fatal("WithLogger did not map to lark client config")
	}
	if clientConfig.LogLevel != larkcore.LogLevelWarn {
		t.Fatal("WithLogLevel did not map to lark client config")
	}
	if clientConfig.BaseUrl != "https://open.example.com" {
		t.Fatalf("WithDomain mapped BaseUrl to %q", clientConfig.BaseUrl)
	}
	if clientConfig.OAuthBaseUrl != "https://accounts.example.com" {
		t.Fatalf("WithOAuthBaseURL mapped OAuthBaseUrl to %q", clientConfig.OAuthBaseUrl)
	}
	if clientConfig.HttpClient != mockHTTP {
		t.Fatal("WithHTTPClient did not map to lark client config")
	}
	if clientConfig.ReqTimeout != timeout {
		t.Fatalf("WithReqTimeout mapped ReqTimeout to %v", clientConfig.ReqTimeout)
	}
	if clientConfig.TokenCache != cache {
		t.Fatal("WithTokenCache did not map to lark client config")
	}
	if clientConfig.EnableTokenCache {
		t.Fatal("WithEnableTokenCache(false) did not map to lark client config")
	}
	if clientConfig.Header.Get("X-Test") != "channel" {
		t.Fatal("WithHeaders did not map to lark client config")
	}
	if clientConfig.Source != "channel-sdk-go-test" {
		t.Fatal("WithSource did not map to lark client config")
	}
	if clientConfig.ClientAssertionProvider != provider {
		t.Fatal("WithClientAssertionProvider did not map to lark client config")
	}
	if !cfg.hasAssertion {
		t.Fatal("WithClientAssertionProvider should mark assertion mode")
	}
	if len(cfg.wsOptions) != 6 {
		t.Fatalf("expected logger, log level, domain, headers, source, and assertion ws options, got %d", len(cfg.wsOptions))
	}
	if len(cfg.eventOptions) != 2 {
		t.Fatalf("expected logger and log level event options, got %d", len(cfg.eventOptions))
	}
}

func TestSeparateEndpointOptionsCompile(t *testing.T) {
	cfg := config{}
	for _, opt := range []Option{
		WithOpenBaseURL("https://open-only.example.com"),
		WithWebSocketDomain("https://ws-only.example.com"),
	} {
		opt(&cfg)
	}

	clientConfig := &larkcore.Config{}
	for _, opt := range cfg.clientOptions {
		opt(clientConfig)
	}

	if clientConfig.BaseUrl != "https://open-only.example.com" {
		t.Fatalf("WithOpenBaseURL mapped BaseUrl to %q", clientConfig.BaseUrl)
	}
	if len(cfg.wsOptions) != 1 {
		t.Fatalf("expected one ws option from WithWebSocketDomain, got %d", len(cfg.wsOptions))
	}
}

func TestNewWithHTTPClientSupportsShortLivedSend(t *testing.T) {
	mockHTTP := &MockHttpClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			respBody := `{"code":0,"msg":"success","data":{"message_id":"om_short","chat_id":"oc_short"}}`
			header := make(http.Header)
			header.Set("Content-Type", "application/json")
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(respBody)),
				Header:     header,
			}, nil
		},
	}

	ch, err := New("cli_a", "secret", WithHTTPClient(mockHTTP))
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	t.Cleanup(func() {
		if err := ch.Stop(context.Background()); err != nil {
			t.Fatalf("Stop returned error: %v", err)
		}
	})

	res, err := ch.Send(context.Background(), &SendInput{
		ReceiveID: "oc_short",
		MsgType:   "text",
		Text:      "hello",
	})
	if err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if res.MessageID != "om_short" {
		t.Fatalf("message id = %q", res.MessageID)
	}
}

func TestNoLegacyConstructorsInRootPackage(t *testing.T) {
	for _, name := range []string{
		"NewChannel",
		"NewWithClient",
		"NewWithClients",
	} {
		if rootPackageHasExportedSymbol(t, name) {
			t.Fatalf("root package must not expose %s", name)
		}
	}
}

func TestNoInternalRuntimeConstructorsInRootPackage(t *testing.T) {
	for _, name := range []string{
		"NewUpdateQueue",
		"NewCardStreamController",
		"NewMarkdownStreamController",
	} {
		if rootPackageHasExportedSymbol(t, name) {
			t.Fatalf("root package must not expose %s", name)
		}
	}
}

func rootPackageHasExportedSymbol(t *testing.T, name string) bool {
	t.Helper()

	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob root go files: %v", err)
	}

	fset := token.NewFileSet()
	for _, file := range files {
		if filepath.Ext(file) != ".go" || strings.HasSuffix(filepath.Base(file), "_test.go") {
			continue
		}
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		parsed, err := parser.ParseFile(fset, file, src, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", file, err)
		}
		for _, decl := range parsed.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if d.Recv == nil && d.Name.Name == name {
					return true
				}
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					switch s := spec.(type) {
					case *ast.TypeSpec:
						if s.Name.Name == name {
							return true
						}
					case *ast.ValueSpec:
						for _, ident := range s.Names {
							if ident.Name == name {
								return true
							}
						}
					}
				}
			}
		}
	}
	return false
}
