// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package streaming

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/larksuite/channel-sdk-go/types"
)

func TestCardControllerUpdateCardPatchesMessage(t *testing.T) {
	driver := &fakeDriver{}
	cfg := types.DefaultChannelConfig()
	cfg.Outbound.Retry.MaxAttempts = 1
	ctrl := NewCardController(driver, cfg, "om_card")

	if err := ctrl.UpdateCard(context.Background(), `{"type":"card"}`); err != nil {
		t.Fatalf("update card: %v", err)
	}
	if got := driver.patchCount(); got != 1 {
		t.Fatalf("patch count = %d, want 1", got)
	}

	driver.mu.Lock()
	patch := driver.patches[0]
	driver.mu.Unlock()
	if patch.messageID != "om_card" || patch.content != `{"type":"card"}` {
		t.Fatalf("unexpected patch op: %+v", patch)
	}
}

func TestCardControllerUpdateCardReturnsDriverError(t *testing.T) {
	driver := &fakeDriver{patchErr: errors.New("patch failed")}
	cfg := types.DefaultChannelConfig()
	cfg.Outbound.Retry.MaxAttempts = 1
	ctrl := NewCardController(driver, cfg, "om_card")

	err := ctrl.UpdateCard(context.Background(), `{"type":"card"}`)
	if err == nil || !strings.Contains(err.Error(), "patch failed") {
		t.Fatalf("expected patch error, got %v", err)
	}
}

func TestCardControllerRejectsUpdateAfterClose(t *testing.T) {
	driver := &fakeDriver{}
	cfg := types.DefaultChannelConfig()
	cfg.Outbound.Retry.MaxAttempts = 1
	ctrl := NewCardController(driver, cfg, "om_card")

	if err := ctrl.Close(context.Background()); err != nil {
		t.Fatalf("close: %v", err)
	}
	err := ctrl.UpdateCard(context.Background(), `{"type":"card"}`)
	if err == nil || !strings.Contains(err.Error(), "stream is closed") {
		t.Fatalf("expected closed stream error, got %v", err)
	}
}
