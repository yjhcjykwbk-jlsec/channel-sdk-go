// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package markdown

import (
	"encoding/json"

	"github.com/larksuite/channel-sdk-go/types"
)

type postContent struct {
	ZhCn postLanguage `json:"zh_cn"`
}

type postLanguage struct {
	Title   string          `json:"title,omitempty"`
	Content [][]postElement `json:"content"`
}

type postElement struct {
	Tag      string `json:"tag"`
	Text     string `json:"text,omitempty"`
	Href     string `json:"href,omitempty"`
	ImageKey string `json:"image_key,omitempty"`
	UserID   string `json:"user_id,omitempty"`
	UserName string `json:"user_name,omitempty"`
}

// SimpleMarkdownToPost wraps raw markdown into a Lark Post JSON string using
// the "md" tag so the Feishu client renders it natively.
func SimpleMarkdownToPost(title, markdown string, mentions []types.Mention) (string, error) {
	content := make([][]postElement, 0, 2)

	if len(mentions) > 0 {
		atElements := composePostMentionElements(mentions)
		if len(atElements) > 0 {
			first := make([]postElement, 0, len(atElements)*2)
			for _, el := range atElements {
				first = append(first, el, postElement{Tag: "text", Text: " "})
			}
			content = append(content, first)
		}
	}

	content = append(content, []postElement{{Tag: "md", Text: markdown}})

	post := postContent{
		ZhCn: postLanguage{
			Title:   title,
			Content: content,
		},
	}

	bytes, err := json.Marshal(post)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func composePostMentionElements(mentions []types.Mention) []postElement {
	if len(mentions) == 0 {
		return nil
	}
	var out []postElement
	for _, m := range mentions {
		if m.UserID == "" {
			continue
		}
		out = append(out, postElement{
			Tag:      "at",
			UserID:   m.UserID,
			UserName: m.Name,
		})
	}
	return out
}
