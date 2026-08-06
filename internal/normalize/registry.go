// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package normalize

import "github.com/larksuite/channel-sdk-go/internal/normalize/converters"

var defaultContentRegistry = converters.DefaultRegistry()

func RegisteredContentTypes() []string {
	return defaultContentRegistry.RegisteredTypes()
}
