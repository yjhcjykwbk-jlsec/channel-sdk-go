// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package converters

import (
	"sort"

	"github.com/larksuite/channel-sdk-go/types"
)

type Converter func(msgType string, content map[string]interface{}) (string, []types.Resource)

type Registry struct {
	converters map[string]Converter
	fallback   Converter
}

func NewRegistry() *Registry {
	return &Registry{
		converters: make(map[string]Converter),
		fallback:   ConvertFallback,
	}
}

func DefaultRegistry() *Registry {
	r := NewRegistry()
	r.Register("text", ConvertText)
	r.Register("image", ConvertImage)
	r.Register("file", ConvertFile)
	r.Register("folder", ConvertFolder)
	r.Register("audio", ConvertAudio)
	r.Register("media", ConvertVideo)
	r.Register("video", ConvertVideo)
	r.Register("sticker", ConvertSticker)
	r.Register("hongbao", ConvertHongbao)
	r.Register("location", ConvertLocation)
	r.Register("share_chat", ConvertShareChat)
	r.Register("share_user", ConvertShareUser)
	r.Register("system", ConvertSystem)
	r.Register("vote", ConvertVote)
	r.Register("video_chat", ConvertVideoChat)
	r.Register("calendar", ConvertCalendar)
	r.Register("general_calendar", ConvertCalendar)
	r.Register("share_calendar_event", ConvertCalendar)
	r.Register("todo", ConvertTodo)
	r.Register("post", ConvertPost)
	r.Register("interactive", ConvertInteractive)
	r.Register("merge_forward", ConvertMergeForward)
	return r
}

func (r *Registry) Register(msgType string, converter Converter) {
	if msgType == "" || converter == nil {
		return
	}
	r.converters[msgType] = converter
}

func (r *Registry) Convert(msgType string, content map[string]interface{}) (string, []types.Resource) {
	if converter, ok := r.converters[msgType]; ok {
		return converter(msgType, content)
	}
	return r.fallback(msgType, content)
}

func (r *Registry) RegisteredTypes() []string {
	types := make([]string, 0, len(r.converters))
	for msgType := range r.converters {
		types = append(types, msgType)
	}
	sort.Strings(types)
	return types
}
