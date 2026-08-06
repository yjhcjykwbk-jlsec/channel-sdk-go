// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package e2e

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigKeepsBlockedUserOptional(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "testdata/e2e/image.jpg")
	writeFixture(t, root, "testdata/e2e/file.pdf")

	dotenv := filepath.Join(root, ".env")
	if err := os.WriteFile(dotenv, []byte(`
APP_ID=cli_e2e
APP_SECRET=secret
CHANNEL_E2E_RECEIVE_OPEN_ID=ou_receive
CHANNEL_E2E_ALLOWED_USER_OPEN_ID=ou_allowed
CHANNEL_E2E_IMAGE_JPG=./testdata/e2e/image.jpg
CHANNEL_E2E_FILE_PDF=./testdata/e2e/file.pdf
`), 0600); err != nil {
		t.Fatalf("write dotenv: %v", err)
	}

	cfg, err := LoadConfig(root, dotenv, nil)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	issues := cfg.Validate()
	if hasErrorIssue(issues, "CHANNEL_E2E_BLOCKED_USER_OPEN_ID") {
		t.Fatalf("blocked user must be optional, got issues: %#v", issues)
	}
	if hasErrorIssue(issues, "APP_ID") || hasErrorIssue(issues, "APP_SECRET") || hasErrorIssue(issues, "CHANNEL_E2E_RECEIVE_OPEN_ID") {
		t.Fatalf("core env should be valid, got issues: %#v", issues)
	}
}

func TestValidateConfigRejectsFixturePathTraversal(t *testing.T) {
	root := t.TempDir()
	dotenv := filepath.Join(root, ".env")
	if err := os.WriteFile(dotenv, []byte(`
APP_ID=cli_e2e
APP_SECRET=secret
CHANNEL_E2E_RECEIVE_OPEN_ID=ou_receive
CHANNEL_E2E_IMAGE_JPG=../outside.jpg
`), 0600); err != nil {
		t.Fatalf("write dotenv: %v", err)
	}

	cfg, err := LoadConfig(root, dotenv, nil)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	issues := cfg.Validate()
	if !hasErrorIssue(issues, "CHANNEL_E2E_IMAGE_JPG") {
		t.Fatalf("expected image path traversal issue, got %#v", issues)
	}
}

func TestLoadConfigReadsSkipAuto(t *testing.T) {
	cfg, err := LoadConfig(t.TempDir(), "", []string{
		"APP_ID=cli_e2e",
		"APP_SECRET=secret",
		"CHANNEL_E2E_RECEIVE_OPEN_ID=ou_receive",
		"CHANNEL_E2E_SKIP_AUTO=1",
	})
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.SkipAuto {
		t.Fatalf("expected CHANNEL_E2E_SKIP_AUTO to be true")
	}
}

func TestBuildPlanSkipsBlockedUserPolicyWhenUnset(t *testing.T) {
	cfg := Config{
		AppID:             "cli_e2e",
		AppSecret:         "secret",
		ReceiveOpenID:     "ou_receive",
		AllowedUserOpenID: "ou_allowed",
		Manual:            true,
		Policy:            true,
	}

	plan := BuildPlan(cfg)
	if !plan.HasSkippedCase("policy.blocked_user") {
		t.Fatalf("expected blocked-user policy case to be skipped, got %#v", plan)
	}
	if plan.HasReadyCase("policy.blocked_user") {
		t.Fatalf("blocked-user policy case must not be ready without CHANNEL_E2E_BLOCKED_USER_OPEN_ID")
	}
}

func writeFixture(t *testing.T, root string, name string) {
	t.Helper()

	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatalf("mkdir fixture dir: %v", err)
	}
	if err := os.WriteFile(path, []byte("fixture"), 0600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
}

func hasErrorIssue(issues []Issue, key string) bool {
	for _, issue := range issues {
		if issue.Severity == SeverityError && issue.Key == key {
			return true
		}
	}
	return false
}
