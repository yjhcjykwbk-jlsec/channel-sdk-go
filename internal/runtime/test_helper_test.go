// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package runtime

import (
	"github.com/larksuite/channel-sdk-go/types"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"
)

type Channel = types.Channel
type channelImpl = Client

func newChannel(client *lark.Client, wsClient *larkws.Client, opts ...types.ChannelOption) Channel {
	return New(client, wsClient, nil, opts...)
}
