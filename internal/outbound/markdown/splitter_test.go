// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package markdown

import (
	"reflect"
	"strings"
	"testing"
)

func TestSplitWithCodeFences(t *testing.T) {
	t.Run("empty text", func(t *testing.T) {
		out := SplitWithCodeFences("", 100)
		if !reflect.DeepEqual(out, []string{""}) {
			t.Fatalf("expected empty chunk, got %#v", out)
		}
	})

	t.Run("short text", func(t *testing.T) {
		text := "Hello world"
		out := SplitWithCodeFences(text, 100)
		if !reflect.DeepEqual(out, []string{text}) {
			t.Fatalf("expected %v, got %v", []string{text}, out)
		}
	})

	t.Run("exact limit", func(t *testing.T) {
		text := strings.Repeat("A", 10)
		out := SplitWithCodeFences(text, 10)
		if !reflect.DeepEqual(out, []string{text}) {
			t.Fatalf("expected single exact-limit chunk, got %#v", out)
		}
	})

	t.Run("split plain text", func(t *testing.T) {
		text := strings.Repeat("A", 40)
		out := SplitWithCodeFences(text, 30)
		// Single long line is not split inside SplitWithCodeFences.
		if len(out) != 1 {
			t.Fatalf("expected 1 chunk, got %d", len(out))
		}
	})

	t.Run("split with code fences", func(t *testing.T) {
		text := "Line 1\n```go\nfunc main() {\n\tfmt.Println(\"Hello\")\n}\n```\nLine 2"
		out := SplitWithCodeFences(text, 40)

		if len(out) != 2 {
			t.Fatalf("expected 2 chunks, got %d", len(out))
		}
		if !strings.HasSuffix(out[0], "```") {
			t.Fatalf("first chunk should close the fence, got: %s", out[0])
		}
		if !strings.HasPrefix(out[1], "```go") {
			t.Fatalf("second chunk should reopen the fence, got: %s", out[1])
		}
	})

	t.Run("split with heading", func(t *testing.T) {
		text := strings.Repeat("A", 30) + "\n# Heading\nMore text"
		out := SplitWithCodeFences(text, 35)

		if len(out) != 2 {
			t.Fatalf("expected 2 chunks, got %d", len(out))
		}
		if strings.Contains(out[0], "# Heading") {
			t.Fatalf("first chunk should break before heading, got: %s", out[0])
		}
		if !strings.HasPrefix(out[1], "# Heading") {
			t.Fatalf("second chunk should start with heading, got: %s", out[1])
		}
	})
}
