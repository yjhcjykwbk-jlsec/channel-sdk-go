// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package outbound

import (
	"testing"

	"github.com/larksuite/channel-sdk-go/types"
)

func TestComposeMentionsTextPrefix(t *testing.T) {
	mentions := []types.Mention{
		{UserID: "ou_123", Name: "Alice"},
		{UserID: "ou_456", Name: "Bob"},
		{Name: "MissingID"},
	}

	prefix := ComposeMentionsTextPrefix(mentions)
	expected := `<at user_id="ou_123">Alice</at> <at user_id="ou_456">Bob</at> `
	if prefix != expected {
		t.Fatalf("expected %q, got %q", expected, prefix)
	}

	if empty := ComposeMentionsTextPrefix(nil); empty != "" {
		t.Fatalf("expected empty string, got %q", empty)
	}
}
