// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package markdown

import (
	"encoding/json"
	"testing"

	"github.com/larksuite/channel-sdk-go/types"
)

func TestSimpleMarkdownToPost(t *testing.T) {
	post, err := SimpleMarkdownToPost("Title", "Hello **World**", nil)
	if err != nil {
		t.Fatal(err)
	}

	var decoded postContent
	if err := json.Unmarshal([]byte(post), &decoded); err != nil {
		t.Fatalf("unmarshal post: %v", err)
	}
	if decoded.ZhCn.Title != "Title" {
		t.Fatalf("expected title, got %q", decoded.ZhCn.Title)
	}
	if len(decoded.ZhCn.Content) != 1 || len(decoded.ZhCn.Content[0]) != 1 {
		t.Fatalf("expected one markdown paragraph, got %#v", decoded.ZhCn.Content)
	}
	md := decoded.ZhCn.Content[0][0]
	if md.Tag != "md" || md.Text != "Hello **World**" {
		t.Fatalf("unexpected markdown element: %#v", md)
	}
}

func TestSimpleMarkdownToPostWithMentions(t *testing.T) {
	post, err := SimpleMarkdownToPost("Title", "Body", []types.Mention{
		{UserID: "ou_123", Name: "Alice"},
	})
	if err != nil {
		t.Fatal(err)
	}

	var decoded postContent
	if err := json.Unmarshal([]byte(post), &decoded); err != nil {
		t.Fatalf("unmarshal post: %v", err)
	}
	if len(decoded.ZhCn.Content) != 2 {
		t.Fatalf("expected mention paragraph plus markdown paragraph, got %#v", decoded.ZhCn.Content)
	}
	at := decoded.ZhCn.Content[0][0]
	if at.Tag != "at" || at.UserID != "ou_123" || at.UserName != "Alice" {
		t.Fatalf("unexpected mention element: %#v", at)
	}
	spacer := decoded.ZhCn.Content[0][1]
	if spacer.Tag != "text" || spacer.Text != " " {
		t.Fatalf("unexpected mention spacer: %#v", spacer)
	}
}
