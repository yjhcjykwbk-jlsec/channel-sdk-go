// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package docs

import (
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

var markdownLinkPattern = regexp.MustCompile(`\[[^\]]*\]\(([^)]+)\)`)

func TestLocalMarkdownLinks(t *testing.T) {
	rootDir := repositoryRoot(t)
	files := []string{
		filepath.Join(rootDir, "README.md"),
		filepath.Join(rootDir, "README.zh.md"),
		filepath.Join(rootDir, "e2e", "README.md"),
	}

	docsDir := filepath.Join(rootDir, "docs")
	err := filepath.WalkDir(docsDir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk documentation: %v", err)
	}

	for _, markdownPath := range files {
		markdownPath := markdownPath
		t.Run(relativePath(t, rootDir, markdownPath), func(t *testing.T) {
			checkLocalLinks(t, rootDir, markdownPath)
		})
	}
}

func TestUserDocsHaveChineseCounterparts(t *testing.T) {
	rootDir := repositoryRoot(t)
	docsDir := filepath.Join(rootDir, "docs")
	entries, err := os.ReadDir(docsDir)
	if err != nil {
		t.Fatalf("read docs directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		translatedPath := filepath.Join(docsDir, "zh-CN", entry.Name())
		if _, err := os.Stat(translatedPath); err != nil {
			t.Errorf("Chinese counterpart for %s: %v", entry.Name(), err)
		}
	}
}

func checkLocalLinks(t *testing.T, rootDir string, markdownPath string) {
	content, err := os.ReadFile(markdownPath)
	if err != nil {
		t.Fatalf("read Markdown: %v", err)
	}

	for _, match := range markdownLinkPattern.FindAllSubmatchIndex(content, -1) {
		rawTarget := strings.TrimSpace(string(content[match[2]:match[3]]))
		if isExternalOrAnchor(rawTarget) {
			continue
		}

		pathPart := strings.SplitN(rawTarget, "#", 2)[0]
		pathPart = strings.Trim(pathPart, "<>")
		decodedPath, err := url.PathUnescape(pathPart)
		if err != nil {
			t.Errorf("line %d: decode link %q: %v", lineNumber(content, match[0]), rawTarget, err)
			continue
		}

		targetPath := filepath.Clean(filepath.Join(filepath.Dir(markdownPath), filepath.FromSlash(decodedPath)))
		targetRelative, err := filepath.Rel(rootDir, targetPath)
		if err != nil {
			t.Errorf("line %d: resolve local link %q: %v", lineNumber(content, match[0]), rawTarget, err)
			continue
		}
		if targetRelative == ".." || strings.HasPrefix(targetRelative, ".."+string(filepath.Separator)) {
			t.Errorf("line %d: local link %q escapes repository root", lineNumber(content, match[0]), rawTarget)
			continue
		}
		if _, err := os.Stat(targetPath); err != nil {
			t.Errorf("line %d: local link %q: %v", lineNumber(content, match[0]), rawTarget, err)
		}
	}
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test file path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), ".."))
}

func relativePath(t *testing.T, rootDir string, path string) string {
	t.Helper()
	relative, err := filepath.Rel(rootDir, path)
	if err != nil {
		t.Fatalf("make relative path: %v", err)
	}
	return relative
}

func isExternalOrAnchor(target string) bool {
	lower := strings.ToLower(target)
	return strings.HasPrefix(target, "#") ||
		strings.HasPrefix(lower, "http://") ||
		strings.HasPrefix(lower, "https://") ||
		strings.HasPrefix(lower, "mailto:")
}

func lineNumber(content []byte, offset int) int {
	return 1 + strings.Count(string(content[:offset]), "\n")
}
