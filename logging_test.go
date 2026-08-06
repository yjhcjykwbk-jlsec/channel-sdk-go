// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package channel

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"testing"
)

type captureLogger struct {
	mu      sync.Mutex
	entries []string
}

func (l *captureLogger) Debug(ctx context.Context, args ...interface{}) {
	l.append(args...)
}

func (l *captureLogger) Info(ctx context.Context, args ...interface{}) {
	l.append(args...)
}

func (l *captureLogger) Warn(ctx context.Context, args ...interface{}) {
	l.append(args...)
}

func (l *captureLogger) Error(ctx context.Context, args ...interface{}) {
	l.append(args...)
}

func (l *captureLogger) append(args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	parts := make([]string, 0, len(args))
	for _, arg := range args {
		parts = append(parts, fmt.Sprint(arg))
	}
	l.entries = append(l.entries, strings.Join(parts, " "))
}

func (l *captureLogger) String() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return strings.Join(l.entries, "\n")
}

func TestSendLogsDoNotContainSensitiveIdentifiers(t *testing.T) {
	var standardLog bytes.Buffer
	previousWriter := log.Writer()
	log.SetOutput(&standardLog)
	t.Cleanup(func() {
		log.SetOutput(previousWriter)
	})

	mockHTTP := &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			respBody := `{"code":230001,"msg":"bad request","data":{}}`
			header := make(http.Header)
			header.Set("Content-Type", "application/json")
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(respBody)),
				Header:     header,
			}, nil
		},
	}
	logger := &captureLogger{}

	ch, err := New("cli_a", "secret", WithHTTPClient(mockHTTP), WithLogger(logger))
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	_, _ = ch.Send(context.Background(), &SendInput{
		ReceiveID:      "oc_sensitive_receive",
		ReplyMessageID: "om_sensitive_reply",
		MsgType:        "interactive",
		Card:           `{"text":"secret card body"}`,
	})

	output := standardLog.String() + "\n" + logger.String()
	for _, secret := range []string{
		"oc_sensitive_receive",
		"om_sensitive_reply",
		"secret card body",
	} {
		if strings.Contains(output, secret) {
			t.Fatalf("send logs contain sensitive value %q in %q", secret, output)
		}
	}
}
