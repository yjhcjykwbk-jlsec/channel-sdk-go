// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package normalize

import "testing"

func TestRegisteredContentTypes(t *testing.T) {
	want := []string{
		"audio",
		"calendar",
		"file",
		"folder",
		"general_calendar",
		"hongbao",
		"image",
		"interactive",
		"location",
		"media",
		"merge_forward",
		"post",
		"share_calendar_event",
		"share_chat",
		"share_user",
		"sticker",
		"system",
		"text",
		"todo",
		"video",
		"video_chat",
		"vote",
	}

	got := RegisteredContentTypes()
	if len(got) != len(want) {
		t.Fatalf("registered type count = %d, want %d: %v", len(got), len(want), got)
	}
	gotSet := make(map[string]bool, len(got))
	for _, msgType := range got {
		gotSet[msgType] = true
	}
	for _, msgType := range want {
		if !gotSet[msgType] {
			t.Fatalf("registered type %q missing from %v", msgType, got)
		}
	}
}

func TestParseContentInvalidJSONFallback(t *testing.T) {
	gotContent, gotResources := ParseContent("text", `{invalid`)
	if gotContent != "[unsupported message]" {
		t.Fatalf("content = %q, want unsupported fallback", gotContent)
	}
	if gotResources != nil {
		t.Fatalf("resources = %+v, want nil", gotResources)
	}
}

func TestParseContentUnknownTypeFallback(t *testing.T) {
	gotContent, gotResources := ParseContent("unknown_type", `{"text":"hello"}`)
	if gotContent != "[unsupported message]" {
		t.Fatalf("content = %q, want unsupported fallback", gotContent)
	}
	if gotResources != nil {
		t.Fatalf("resources = %+v, want nil", gotResources)
	}
}

func TestParseContentMergeForwardRawString(t *testing.T) {
	gotContent, gotResources := ParseContent("merge_forward", "Merged and Forwarded Message")
	if gotContent != "Merged and Forwarded Message" {
		t.Fatalf("content = %q, want raw merge-forward content", gotContent)
	}
	if gotResources != nil {
		t.Fatalf("resources = %+v, want nil", gotResources)
	}
}
