// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package runtime

import (
	"testing"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"
)

func TestNewChannel(t *testing.T) {
	client := lark.NewClient("appID", "appSecret")
	wsClient := larkws.NewClient("appID", "appSecret")

	ch := newChannel(client, wsClient)
	if ch == nil {
		t.Fatal("expected channel instance, got nil")
	}

	impl, ok := ch.(*channelImpl)
	if !ok {
		t.Fatal("expected channelImpl type")
	}

	if impl.client != client {
		t.Errorf("expected client to be injected correctly")
	}

	if impl.wsClient != wsClient {
		t.Errorf("expected wsClient to be injected correctly")
	}
}
