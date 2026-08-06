// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package converters

import (
	"reflect"
	"testing"

	"github.com/larksuite/channel-sdk-go/types"
)

func TestConvertersMissingFields(t *testing.T) {
	tests := []struct {
		name        string
		converter   Converter
		msgType     string
		content     map[string]interface{}
		wantContent string
	}{
		{name: "text", converter: ConvertText, msgType: "text", content: map[string]interface{}{}, wantContent: ""},
		{name: "image", converter: ConvertImage, msgType: "image", content: map[string]interface{}{}, wantContent: "[image]"},
		{name: "file", converter: ConvertFile, msgType: "file", content: map[string]interface{}{}, wantContent: "[file]"},
		{name: "folder", converter: ConvertFolder, msgType: "folder", content: map[string]interface{}{}, wantContent: "[folder]"},
		{name: "audio", converter: ConvertAudio, msgType: "audio", content: map[string]interface{}{}, wantContent: "[audio]"},
		{name: "video", converter: ConvertVideo, msgType: "video", content: map[string]interface{}{}, wantContent: "[video]"},
		{name: "sticker", converter: ConvertSticker, msgType: "sticker", content: map[string]interface{}{}, wantContent: "[sticker]"},
		{name: "system", converter: ConvertSystem, msgType: "system", content: map[string]interface{}{}, wantContent: "[system message]"},
		{name: "vote", converter: ConvertVote, msgType: "vote", content: map[string]interface{}{}, wantContent: "<vote>\n[vote]\n</vote>"},
		{name: "video_chat", converter: ConvertVideoChat, msgType: "video_chat", content: map[string]interface{}{}, wantContent: "<meeting>\n[video chat]\n</meeting>"},
		{name: "calendar", converter: ConvertCalendar, msgType: "calendar", content: map[string]interface{}{}, wantContent: "<calendar_invite>\n[calendar event]\n</calendar_invite>"},
		{name: "todo", converter: ConvertTodo, msgType: "todo", content: map[string]interface{}{}, wantContent: "<todo>\n[todo]\n</todo>"},
		{name: "post", converter: ConvertPost, msgType: "post", content: map[string]interface{}{}, wantContent: "[rich text message]"},
		{name: "interactive", converter: ConvertInteractive, msgType: "interactive", content: map[string]interface{}{}, wantContent: "[interactive card]"},
		{name: "merge_forward", converter: ConvertMergeForward, msgType: "merge_forward", content: map[string]interface{}{}, wantContent: "Merged and Forwarded Message"},
		{name: "fallback", converter: ConvertFallback, msgType: "unknown", content: map[string]interface{}{}, wantContent: "[unsupported message]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotContent, gotResources := tt.converter(tt.msgType, tt.content)
			if gotContent != tt.wantContent {
				t.Fatalf("content = %q, want %q", gotContent, tt.wantContent)
			}
			if gotResources != nil {
				t.Fatalf("resources = %+v, want nil", gotResources)
			}
		})
	}
}

func TestRegistryUnknownTypeUsesFallback(t *testing.T) {
	registry := DefaultRegistry()
	gotContent, gotResources := registry.Convert("unknown", map[string]interface{}{"text": "hello"})
	if gotContent != "[unsupported message]" {
		t.Fatalf("content = %q, want unsupported fallback", gotContent)
	}
	if gotResources != nil {
		t.Fatalf("resources = %+v, want nil", gotResources)
	}
}

func TestPostConverterExtractsRichContent(t *testing.T) {
	content := map[string]interface{}{
		"zh_cn": map[string]interface{}{
			"title": "Post Title",
			"content": []interface{}{
				[]interface{}{
					map[string]interface{}{"tag": "text", "text": "Hello "},
					map[string]interface{}{"tag": "a", "text": "Link", "href": "https://example.com"},
				},
				[]interface{}{
					map[string]interface{}{"tag": "img", "image_key": "img_abc"},
				},
			},
		},
	}

	gotContent, gotResources := ConvertPost("post", content)
	wantContent := "**Post Title**\n\nHello [Link](https://example.com)\n![image](img_abc)"
	if gotContent != wantContent {
		t.Fatalf("content = %q, want %q", gotContent, wantContent)
	}
	wantResources := []types.Resource{{Type: "image", FileKey: "img_abc"}}
	if !reflect.DeepEqual(gotResources, wantResources) {
		t.Fatalf("resources = %+v, want %+v", gotResources, wantResources)
	}
}

func TestPostConverterDoesNotRewriteMarkdownInsideCodeBlock(t *testing.T) {
	content := map[string]interface{}{
		"zh_cn": map[string]interface{}{
			"content_v2": []interface{}{
				[]interface{}{
					map[string]interface{}{"tag": "md", "text": "Before\n```go\n<at user_id=\"ou_1\">name</at>\n![image](k)\n```\nAfter"},
				},
			},
		},
	}

	gotContent, gotResources := ConvertPost("post", content)
	wantContent := "Before\n```go\n<at user_id=\"ou_1\">name</at>\n![image](k)\n```\nAfter"
	if gotContent != wantContent {
		t.Fatalf("content = %q, want %q", gotContent, wantContent)
	}
	if gotResources != nil {
		t.Fatalf("resources = %+v, want nil", gotResources)
	}
}
