// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package outbound

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/larksuite/channel-sdk-go/types"
)

func TestBuildStaticContentPriority(t *testing.T) {
	msgType, content, ok, err := buildStaticContent(&types.SendInput{
		ImageKey: "img_123",
		Text:     "text should not win",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !ok || msgType != "image" {
		t.Fatalf("expected image content, got msgType=%q ok=%t", msgType, ok)
	}

	var decoded map[string]string
	if err := json.Unmarshal([]byte(content), &decoded); err != nil {
		t.Fatalf("unmarshal content: %v", err)
	}
	if decoded["image_key"] != "img_123" {
		t.Fatalf("unexpected content: %#v", decoded)
	}
}

func TestBuildStaticContentShareUser(t *testing.T) {
	msgType, content, ok, err := buildStaticContent(&types.SendInput{ShareUserID: "ou_user"})
	if err != nil {
		t.Fatal(err)
	}
	if !ok || msgType != "share_user" {
		t.Fatalf("expected share_user content, got msgType=%q ok=%t", msgType, ok)
	}

	var decoded map[string]string
	if err := json.Unmarshal([]byte(content), &decoded); err != nil {
		t.Fatalf("unmarshal content: %v", err)
	}
	if decoded["user_id"] != "ou_user" {
		t.Fatalf("unexpected content: %#v", decoded)
	}
}

func TestBuildStaticContentNoMatch(t *testing.T) {
	msgType, content, ok, err := buildStaticContent(&types.SendInput{})
	if err != nil {
		t.Fatal(err)
	}
	if ok || msgType != "" || content != "" {
		t.Fatalf("expected no static content, got msgType=%q content=%q ok=%t", msgType, content, ok)
	}
}

func TestSplitPlain(t *testing.T) {
	if got := splitPlain("hello", 10); !reflect.DeepEqual(got, []string{"hello"}) {
		t.Fatalf("short text mismatch: %#v", got)
	}
	if got := splitPlain("abcdef", 2); !reflect.DeepEqual(got, []string{"ab", "cd", "ef"}) {
		t.Fatalf("split text mismatch: %#v", got)
	}
	if got := splitPlain("hello", 0); !reflect.DeepEqual(got, []string{"hello"}) {
		t.Fatalf("zero limit mismatch: %#v", got)
	}
}
