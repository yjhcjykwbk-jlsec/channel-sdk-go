// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package pipeline

import (
	"strings"

	"github.com/larksuite/channel-sdk-go/types"
)

func mergeBatch(batch []*types.NormalizedMessage) *types.NormalizedMessage {
	if len(batch) == 1 {
		return batch[0]
	}

	last := batch[len(batch)-1]

	var contents []string
	for _, m := range batch {
		if m.Content != "" {
			contents = append(contents, m.Content)
		}
	}
	content := strings.Join(contents, "\n\n")

	var mentionAll, mentionedBot bool
	var resources []types.Resource
	var mentions []types.Mention

	seenResources := make(map[string]bool)
	seenMentions := make(map[string]bool)

	for _, m := range batch {
		if m.MentionAll {
			mentionAll = true
		}
		if m.MentionedBot {
			mentionedBot = true
		}

		for _, r := range m.Resources {
			if !seenResources[r.FileKey] {
				seenResources[r.FileKey] = true
				resources = append(resources, r)
			}
		}

		for _, mn := range m.Mentions {
			id := mn.UserID
			if id == "" {
				id = mn.Name
			}
			if !seenMentions[id] {
				seenMentions[id] = true
				mentions = append(mentions, mn)
			}
		}
	}

	merged := *last
	merged.Content = content
	merged.MentionAll = mentionAll
	merged.MentionedBot = mentionedBot
	merged.Resources = resources
	merged.Mentions = mentions

	return &merged
}

func extractSourceIDs(batch []*types.NormalizedMessage) []string {
	ids := make([]string, len(batch))
	for i, m := range batch {
		ids[i] = m.MessageID
	}
	return ids
}
