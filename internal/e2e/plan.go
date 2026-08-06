// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package e2e

import "strings"

type CaseMode string

const (
	CaseModeAuto   CaseMode = "auto"
	CaseModeManual CaseMode = "manual"
)

type CaseStatus string

const (
	CaseStatusReady   CaseStatus = "ready"
	CaseStatusSkipped CaseStatus = "skipped"
)

type Case struct {
	ID      string
	Name    string
	Mode    CaseMode
	Status  CaseStatus
	Reason  string
	Depends []string
}

type Plan struct {
	Cases []Case
}

func (p Plan) ReadyCases() []Case {
	var out []Case
	for _, c := range p.Cases {
		if c.Status == CaseStatusReady {
			out = append(out, c)
		}
	}
	return out
}

func (p Plan) SkippedCases() []Case {
	var out []Case
	for _, c := range p.Cases {
		if c.Status == CaseStatusSkipped {
			out = append(out, c)
		}
	}
	return out
}

func (p Plan) HasReadyCase(id string) bool {
	for _, c := range p.Cases {
		if c.ID == id && c.Status == CaseStatusReady {
			return true
		}
	}
	return false
}

func (p Plan) HasSkippedCase(id string) bool {
	for _, c := range p.Cases {
		if c.ID == id && c.Status == CaseStatusSkipped {
			return true
		}
	}
	return false
}

func BuildPlan(cfg Config) Plan {
	var plan Plan
	add := func(id, name string, mode CaseMode, deps map[string]string) {
		c := Case{
			ID:      id,
			Name:    name,
			Mode:    mode,
			Status:  CaseStatusReady,
			Depends: sortedKeys(deps),
		}
		var missing []string
		for key, value := range deps {
			if strings.TrimSpace(value) == "" {
				missing = append(missing, key)
			}
		}
		if len(missing) > 0 {
			c.Status = CaseStatusSkipped
			c.Reason = "missing " + strings.Join(sortedStrings(missing), ", ")
		}
		plan.Cases = append(plan.Cases, c)
	}

	add("send.text", "send text message", CaseModeAuto, map[string]string{
		"CHANNEL_E2E_RECEIVE_OPEN_ID": cfg.ReceiveOpenID,
	})
	add("lifecycle.bot_identity", "fetch bot identity", CaseModeAuto, map[string]string{})
	add("lifecycle.stop", "stop channel gracefully", CaseModeAuto, map[string]string{})
	add("raw.chat_info", "get chat info through RawClient", CaseModeAuto, map[string]string{
		"CHANNEL_E2E_GROUP_CHAT_ID": cfg.GroupChatID,
	})
	add("raw.message_list", "list chat message history through RawClient", CaseModeAuto, map[string]string{
		"CHANNEL_E2E_GROUP_CHAT_ID": cfg.GroupChatID,
	})
	add("send.markdown", "send markdown as post", CaseModeAuto, map[string]string{
		"CHANNEL_E2E_RECEIVE_OPEN_ID": cfg.ReceiveOpenID,
	})
	add("send.long_markdown", "send long markdown in chunks", CaseModeAuto, map[string]string{
		"CHANNEL_E2E_RECEIVE_OPEN_ID": cfg.ReceiveOpenID,
	})
	add("send.post", "send raw post message", CaseModeAuto, map[string]string{
		"CHANNEL_E2E_RECEIVE_OPEN_ID": cfg.ReceiveOpenID,
	})
	add("send.mention", "send group text with mention", CaseModeAuto, map[string]string{
		"CHANNEL_E2E_GROUP_CHAT_ID":        cfg.GroupChatID,
		"CHANNEL_E2E_MENTION_USER_OPEN_ID": cfg.MentionUserID,
	})
	add("send.group_text", "send text to group by chat_id", CaseModeAuto, map[string]string{
		"CHANNEL_E2E_GROUP_CHAT_ID": cfg.GroupChatID,
	})
	add("send.image_jpg", "upload, send, and download JPG image", CaseModeAuto, map[string]string{
		"CHANNEL_E2E_RECEIVE_OPEN_ID": cfg.ReceiveOpenID,
		"CHANNEL_E2E_IMAGE_JPG":       cfg.Media.ImageJPG,
	})
	add("send.image_png", "upload, send, and download PNG image", CaseModeAuto, map[string]string{
		"CHANNEL_E2E_RECEIVE_OPEN_ID": cfg.ReceiveOpenID,
		"CHANNEL_E2E_IMAGE_PNG":       cfg.Media.ImagePNG,
	})
	add("send.image_gif", "upload, send, and download GIF image", CaseModeAuto, map[string]string{
		"CHANNEL_E2E_RECEIVE_OPEN_ID": cfg.ReceiveOpenID,
		"CHANNEL_E2E_IMAGE_GIF":       cfg.Media.ImageGIF,
	})
	add("send.image_webp", "upload, send, and download WEBP image", CaseModeAuto, map[string]string{
		"CHANNEL_E2E_RECEIVE_OPEN_ID": cfg.ReceiveOpenID,
		"CHANNEL_E2E_IMAGE_WEBP":      cfg.Media.ImageWEBP,
	})
	add("send.file_pdf", "upload, send, and download PDF file", CaseModeAuto, map[string]string{
		"CHANNEL_E2E_RECEIVE_OPEN_ID": cfg.ReceiveOpenID,
		"CHANNEL_E2E_FILE_PDF":        cfg.Media.FilePDF,
	})
	add("send.file_office", "upload and send DOCX/XLSX/PPTX files", CaseModeAuto, map[string]string{
		"CHANNEL_E2E_RECEIVE_OPEN_ID": cfg.ReceiveOpenID,
		"CHANNEL_E2E_FILE_DOCX":       cfg.Media.FileDOCX,
		"CHANNEL_E2E_FILE_XLSX":       cfg.Media.FileXLSX,
		"CHANNEL_E2E_FILE_PPTX":       cfg.Media.FilePPTX,
	})
	add("send.audio", "upload and send OGG/Opus audio", CaseModeAuto, map[string]string{
		"CHANNEL_E2E_RECEIVE_OPEN_ID": cfg.ReceiveOpenID,
		"CHANNEL_E2E_AUDIO_OGG":       cfg.Media.AudioOGG,
	})
	add("send.video", "upload and send MP4 video", CaseModeAuto, map[string]string{
		"CHANNEL_E2E_RECEIVE_OPEN_ID": cfg.ReceiveOpenID,
		"CHANNEL_E2E_VIDEO_MP4":       cfg.Media.VideoMP4,
	})
	add("send.share_chat", "send group share card", CaseModeAuto, map[string]string{
		"CHANNEL_E2E_RECEIVE_OPEN_ID": cfg.ReceiveOpenID,
		"CHANNEL_E2E_SHARE_CHAT_ID":   cfg.ShareChatID,
	})
	add("send.share_user", "send user share card", CaseModeAuto, map[string]string{
		"CHANNEL_E2E_RECEIVE_OPEN_ID":      cfg.ReceiveOpenID,
		"CHANNEL_E2E_MENTION_USER_OPEN_ID": cfg.MentionUserID,
	})
	add("send.sticker", "send sticker by file key", CaseModeAuto, map[string]string{
		"CHANNEL_E2E_RECEIVE_OPEN_ID":  cfg.ReceiveOpenID,
		"CHANNEL_E2E_STICKER_FILE_KEY": cfg.StickerFileKey,
	})
	add("send.card", "send interactive card", CaseModeAuto, map[string]string{
		"CHANNEL_E2E_RECEIVE_OPEN_ID": cfg.ReceiveOpenID,
	})
	add("reply.text", "reply with text", CaseModeAuto, map[string]string{
		"CHANNEL_E2E_RECEIVE_OPEN_ID": cfg.ReceiveOpenID,
	})
	add("reply.markdown", "reply with markdown", CaseModeAuto, map[string]string{
		"CHANNEL_E2E_RECEIVE_OPEN_ID": cfg.ReceiveOpenID,
	})
	add("reply.image", "reply with image", CaseModeAuto, map[string]string{
		"CHANNEL_E2E_RECEIVE_OPEN_ID": cfg.ReceiveOpenID,
		"CHANNEL_E2E_IMAGE_JPG":       cfg.Media.ImageJPG,
	})
	add("stream.markdown", "send and update markdown stream", CaseModeAuto, map[string]string{
		"CHANNEL_E2E_RECEIVE_OPEN_ID": cfg.ReceiveOpenID,
	})
	add("stream.card_update", "send and patch card stream", CaseModeAuto, map[string]string{
		"CHANNEL_E2E_RECEIVE_OPEN_ID": cfg.ReceiveOpenID,
	})

	if cfg.Manual {
		add("event.message", "receive message event", CaseModeManual, map[string]string{
			"CHANNEL_E2E_RECEIVE_OPEN_ID": cfg.ReceiveOpenID,
		})
		add("event.reaction_created", "receive reaction created event", CaseModeManual, map[string]string{
			"CHANNEL_E2E_RECEIVE_OPEN_ID": cfg.ReceiveOpenID,
		})
		add("event.reaction_deleted", "receive reaction deleted event", CaseModeManual, map[string]string{
			"CHANNEL_E2E_RECEIVE_OPEN_ID": cfg.ReceiveOpenID,
		})
		add("event.card_action", "receive card action callback", CaseModeManual, map[string]string{
			"CHANNEL_E2E_RECEIVE_OPEN_ID": cfg.ReceiveOpenID,
		})
		add("event.comment", "receive document comment event", CaseModeManual, map[string]string{
			"CHANNEL_E2E_DOC_TOKEN": cfg.DocToken,
		})
	}

	if cfg.BotAdded {
		add("event.bot_added", "receive bot-added group event", CaseModeManual, map[string]string{
			"CHANNEL_E2E_GROUP_CHAT_ID": cfg.GroupChatID,
		})
	}

	if cfg.Policy {
		add("policy.allowed_group", "allow configured group message", CaseModeManual, map[string]string{
			"CHANNEL_E2E_ALLOWED_GROUP_CHAT_ID": cfg.AllowedGroupID,
		})
		add("policy.blocked_group", "reject unallowed group message", CaseModeManual, map[string]string{
			"CHANNEL_E2E_BLOCKED_GROUP_CHAT_ID": cfg.BlockedGroupID,
		})
		add("policy.allowed_user", "allow configured direct-message sender", CaseModeManual, map[string]string{
			"CHANNEL_E2E_ALLOWED_USER_OPEN_ID": cfg.AllowedUserOpenID,
		})
		add("policy.blocked_user", "reject unallowed direct-message sender", CaseModeManual, map[string]string{
			"CHANNEL_E2E_BLOCKED_USER_OPEN_ID": cfg.BlockedUserOpenID,
		})
	}

	return plan
}

func sortedKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	return sortedStrings(keys)
}

func sortedStrings(values []string) []string {
	out := append([]string(nil), values...)
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j] < out[j-1]; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}
