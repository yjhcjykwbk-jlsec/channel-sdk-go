// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package media

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/larksuite/channel-sdk-go/internal/safety"
	"github.com/larksuite/channel-sdk-go/types"
	lark "github.com/larksuite/oapi-sdk-go/v3"
)

type MockUploader struct {
	UploadImagePathFunc func(ctx context.Context, imageType string, path string) (string, error)
	UploadFilePathFunc  func(ctx context.Context, fileType string, path string) (string, error)
}

func (m *MockUploader) UploadImagePath(ctx context.Context, imageType string, path string) (string, error) {
	if m.UploadImagePathFunc != nil {
		return m.UploadImagePathFunc(ctx, imageType, path)
	}
	return "", nil
}

func (m *MockUploader) UploadFilePath(ctx context.Context, fileType string, path string) (string, error) {
	if m.UploadFilePathFunc != nil {
		return m.UploadFilePathFunc(ctx, fileType, path)
	}
	return "", nil
}

func TestUploader(t *testing.T) {
	mockHttp := &MockHttpClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			respBody := `{"code":0,"msg":"success","data":{"image_key":"img_123"}}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(respBody)),
				Header:     make(http.Header),
			}, nil
		},
	}

	client := lark.NewClient("appId", "appSecret", lark.WithHttpClient(mockHttp))
	uploader := NewUploader(client)

	// Just test instantiation since actual upload involves reading files
	if uploader == nil {
		t.Errorf("expected uploader to be non-nil")
	}
}

func TestUploadMediaNoSource(t *testing.T) {
	uploader := NewUploader(lark.NewClient("appId", "appSecret"))

	_, err := uploader.UploadMedia(context.Background(), &types.UploadInput{Kind: types.MediaKindImage}, nil)
	if err == nil || !strings.Contains(err.Error(), "no source provided") {
		t.Fatalf("expected no source error, got %v", err)
	}
}

func TestUploadMediaInvalidPath(t *testing.T) {
	uploader := NewUploader(lark.NewClient("appId", "appSecret"))

	_, err := uploader.UploadMedia(context.Background(), &types.UploadInput{
		Kind:       types.MediaKindImage,
		SourcePath: "/path/that/does/not/exist",
	}, nil)
	if err == nil || !strings.Contains(err.Error(), "failed to read file") {
		t.Fatalf("expected read file error, got %v", err)
	}
}

func TestUploadMediaRejectsPrivateURLByDefault(t *testing.T) {
	uploader := NewUploader(lark.NewClient("appId", "appSecret"))

	_, err := uploader.UploadMedia(context.Background(), &types.UploadInput{
		Kind:      types.MediaKindImage,
		SourceURL: "http://127.0.0.1/image.png",
	}, nil)
	if err == nil || !strings.Contains(err.Error(), "ssrf_blocked") {
		t.Fatalf("expected ssrf_blocked error, got %v", err)
	}
}

func TestUploadMediaAllowsAllowlistedHost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("not-an-opus-file")); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("parse server url: %v", err)
	}
	uploader := NewUploader(lark.NewClient("appId", "appSecret"))

	_, err = uploader.UploadMedia(context.Background(), &types.UploadInput{
		Kind:      types.MediaKindAudio,
		SourceURL: server.URL,
	}, &SourceOptions{
		SsrfGuard: &safety.SsrfGuardOptions{
			Allowlist: []string{serverURL.Hostname()},
		},
		MaxDownloadSize: 1024,
	})
	if err == nil {
		t.Fatalf("expected invalid duration error")
	}
	if strings.Contains(err.Error(), "ssrf_blocked") {
		t.Fatalf("expected allowlisted host to pass SSRF guard, got %v", err)
	}
	if !strings.Contains(err.Error(), "duration could not be determined") {
		t.Fatalf("expected invalid duration error, got %v", err)
	}
}

func TestUploadMediaURLDownloadLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("abcdef")); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("parse server url: %v", err)
	}
	uploader := NewUploader(lark.NewClient("appId", "appSecret"))

	_, err = uploader.UploadMedia(context.Background(), &types.UploadInput{
		Kind:      types.MediaKindImage,
		SourceURL: server.URL,
	}, &SourceOptions{
		SsrfGuard: &safety.SsrfGuardOptions{
			Allowlist: []string{serverURL.Hostname()},
		},
		MaxDownloadSize: 3,
	})
	if err == nil || !strings.Contains(err.Error(), "body exceeds max download size") {
		t.Fatalf("expected max download size error, got %v", err)
	}
}

func TestUploadMediaPathTraversalRejected(t *testing.T) {
	uploader := NewUploader(lark.NewClient("appId", "appSecret"))

	_, err := uploader.UploadMedia(context.Background(), &types.UploadInput{
		Kind:       types.MediaKindImage,
		SourcePath: "../secret.png",
	}, nil)
	if err == nil || !strings.Contains(err.Error(), "path traversal is not allowed") {
		t.Fatalf("expected path traversal error, got %v", err)
	}
}

// MockHttpClient is a mock HTTP client for testing.
type MockHttpClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHttpClient) Do(req *http.Request) (*http.Response, error) {
	if m.DoFunc != nil {
		return m.DoFunc(req)
	}
	return nil, nil
}
