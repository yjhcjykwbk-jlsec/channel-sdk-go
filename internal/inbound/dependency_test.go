// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package inbound

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInboundDoesNotImportRuntime(t *testing.T) {
	err := filepath.WalkDir(".", func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.Contains(string(src), "internal/runtime") {
			t.Fatalf("%s imports internal/runtime", path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk inbound package: %v", err)
	}
}
