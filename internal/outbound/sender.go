// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package outbound

import (
	"context"
	"fmt"

	outboundmarkdown "github.com/larksuite/channel-sdk-go/internal/outbound/markdown"
	"github.com/larksuite/channel-sdk-go/types"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
)

type Uploader interface {
	UploadImagePath(ctx context.Context, imageType string, imagePath string) (string, error)
	UploadFilePath(ctx context.Context, fileType string, filePath string) (string, error)
	UploadMedia(ctx context.Context, input *types.UploadInput) (*types.UploadResult, error)
}

type Sender struct {
	driver   MessageDriver
	uploader Uploader
	config   types.ChannelConfig
	logger   larkcore.Logger
}

func NewSender(driver MessageDriver, uploader Uploader, config types.ChannelConfig, logger larkcore.Logger) *Sender {
	return &Sender{
		driver:   driver,
		uploader: uploader,
		config:   config,
		logger:   logger,
	}
}

func (s *Sender) Send(ctx context.Context, input *types.SendInput) (*types.SendResult, error) {
	if input == nil {
		return nil, fmt.Errorf("input cannot be nil")
	}

	receiveIDType, receiveID, err := resolveReceiveTarget(input)
	if err != nil {
		return nil, err
	}

	if err := s.prepareUploads(ctx, input); err != nil {
		return nil, err
	}

	msgType, content, ok, err := buildStaticContent(input)
	if err != nil {
		return nil, err
	}
	if ok {
		if msgType == "" || content == "" {
			return nil, fmt.Errorf("no valid message content provided")
		}
		res, err := s.sendOneWithFallback(ctx, receiveIDType, receiveID, msgType, content, input)
		if err != nil {
			return nil, err
		}
		return &types.SendResult{MessageID: res.MessageID, ChatID: res.ChatID}, nil
	}

	if input.Markdown != "" {
		return s.sendMarkdown(ctx, receiveIDType, receiveID, input)
	}
	if input.Text != "" {
		return s.sendText(ctx, receiveIDType, receiveID, input)
	}

	return nil, fmt.Errorf("no valid message content provided")
}

func resolveReceiveTarget(input *types.SendInput) (string, string, error) {
	receiveIDType := "open_id"
	receiveID := input.UserID
	if input.ReceiveID != "" {
		receiveID = input.ReceiveID
		t, err := DetectReceiveIdType(receiveID)
		if err == nil {
			receiveIDType = string(t)
		}
	} else if input.ChatID != "" {
		receiveIDType = "chat_id"
		receiveID = input.ChatID
	}

	if receiveID == "" {
		return "", "", fmt.Errorf("ReceiveID, ChatID, or UserID must be provided")
	}
	return receiveIDType, receiveID, nil
}

func (s *Sender) prepareUploads(ctx context.Context, input *types.SendInput) error {
	if input.ImagePath != "" && input.ImageKey == "" {
		if s.uploader == nil {
			return fmt.Errorf("uploader is required for image path")
		}
		imageKey, err := s.uploader.UploadImagePath(ctx, "message", input.ImagePath)
		if err != nil {
			return fmt.Errorf("failed to upload image: %v", err)
		}
		input.ImageKey = imageKey
	}

	if input.FilePath != "" && input.FileKey == "" {
		if s.uploader == nil {
			return fmt.Errorf("uploader is required for file path")
		}
		fileKey, err := s.uploader.UploadFilePath(ctx, "stream", input.FilePath)
		if err != nil {
			return fmt.Errorf("failed to upload file: %v", err)
		}
		input.FileKey = fileKey
	}

	if input.Media != nil && input.AudioKey == "" && input.VideoKey == "" && input.FileKey == "" && input.ImageKey == "" {
		if s.uploader == nil {
			return fmt.Errorf("uploader is required for media")
		}
		res, err := s.uploader.UploadMedia(ctx, input.Media)
		if err != nil {
			return fmt.Errorf("failed to upload media: %v", err)
		}
		if res.Kind == types.MediaKindImage {
			input.ImageKey = res.FileKey
		} else if res.Kind == types.MediaKindAudio {
			input.AudioKey = res.FileKey
		} else if res.Kind == types.MediaKindVideo {
			input.VideoKey = res.FileKey
		} else {
			input.FileKey = res.FileKey
		}
	}
	return nil
}

func (s *Sender) sendMarkdown(ctx context.Context, receiveIDType, receiveID string, input *types.SendInput) (*types.SendResult, error) {
	chunks := outboundmarkdown.SplitWithCodeFences(input.Markdown, s.config.Outbound.TextChunkLimit)
	ids := make([]string, 0, len(chunks))
	firstChatID := ""
	for i, chunk := range chunks {
		var mentions []types.Mention
		if i == 0 {
			mentions = input.Mentions
		}
		postJSON, err := outboundmarkdown.SimpleMarkdownToPost(input.Title, chunk, mentions)
		if err != nil {
			return nil, fmt.Errorf("failed to format markdown: %v", err)
		}
		res, err := s.sendOneWithFallback(ctx, receiveIDType, receiveID, "post", postJSON, input)
		if err != nil {
			return nil, err
		}
		ids = append(ids, res.MessageID)
		if i == 0 {
			firstChatID = res.ChatID
		}
	}
	result := &types.SendResult{MessageID: ids[0], ChatID: firstChatID}
	if len(ids) > 1 {
		result.ChunkIDs = ids
	}
	return result, nil
}

func (s *Sender) sendText(ctx context.Context, receiveIDType, receiveID string, input *types.SendInput) (*types.SendResult, error) {
	prefix := ComposeMentionsTextPrefix(input.Mentions)
	fullText := prefix + input.Text
	chunks := splitPlain(fullText, s.config.Outbound.TextChunkLimit)
	ids := make([]string, 0, len(chunks))
	firstChatID := ""
	for i, chunk := range chunks {
		content, err := marshalContent(map[string]string{"text": chunk})
		if err != nil {
			return nil, err
		}
		res, err := s.sendOneWithFallback(ctx, receiveIDType, receiveID, "text", content, input)
		if err != nil {
			return nil, err
		}
		ids = append(ids, res.MessageID)
		if i == 0 {
			firstChatID = res.ChatID
		}
	}
	result := &types.SendResult{MessageID: ids[0], ChatID: firstChatID}
	if len(ids) > 1 {
		result.ChunkIDs = ids
	}
	return result, nil
}

func (s *Sender) sendOneWithFallback(ctx context.Context, idType, id, msgType, content string, input *types.SendInput) (*SendResponse, error) {
	isReply := input.ReplyMessageID != ""
	s.logInfo(ctx, fmt.Sprintf("[Channel] send message start msgType=%s reply=%t", msgType, isReply))
	res, err := s.rawSendWithRetry(ctx, idType, id, msgType, content, input.ReplyMessageID)
	if err == nil {
		s.logInfo(ctx, fmt.Sprintf("[Channel] send message success msgType=%s reply=%t", msgType, isReply))
		return res, nil
	}

	fce := types.ClassifyError(err)
	s.logWarn(ctx, fmt.Sprintf("[Channel] send message failed msgType=%s reply=%t code=%s", msgType, isReply, fce.Code))

	if fce.Code == types.ErrCodeTargetRevoked && input.ReplyMessageID != "" {
		s.logInfo(ctx, fmt.Sprintf("[Channel] send fallback retry_as_new msgType=%s code=%s", msgType, fce.Code))
		input.ReplyMessageID = ""
		return s.sendOneWithFallback(ctx, idType, id, msgType, content, input)
	}

	if fce.Code == types.ErrCodeFormatError && msgType != "text" {
		s.logInfo(ctx, fmt.Sprintf("[Channel] send fallback downgrade_to_text msgType=%s code=%s", msgType, fce.Code))
		fallbackText := ""
		if input.Markdown != "" {
			fallbackText = input.Markdown
		} else if input.Text != "" {
			fallbackText = input.Text
		} else {
			s.logWarn(ctx, fmt.Sprintf("[Channel] send fallback unavailable msgType=%s code=%s", msgType, fce.Code))
			return nil, err
		}

		prefix := ComposeMentionsTextPrefix(input.Mentions)
		fullText := prefix + fallbackText
		fallbackContent, marshalErr := marshalContent(map[string]string{"text": fullText})
		if marshalErr != nil {
			return nil, marshalErr
		}
		return s.rawSendWithRetry(ctx, idType, id, "text", fallbackContent, input.ReplyMessageID)
	}

	return nil, err
}

func (s *Sender) rawSendWithRetry(ctx context.Context, idType, id, msgType, content, replyMessageID string) (*SendResponse, error) {
	op := func(attempt int) (interface{}, error) {
		if replyMessageID != "" {
			s.logInfo(ctx, fmt.Sprintf("[Channel] send reply attempt=%d msgType=%s reply=%t", attempt, msgType, true))
			resp, err := s.driver.ReplyMessage(ctx, replyMessageID, msgType, content)
			if err != nil {
				s.logWarn(ctx, fmt.Sprintf("[Channel] send reply error attempt=%d msgType=%s", attempt, msgType))
				return nil, err
			}
			return resp, nil
		}

		s.logInfo(ctx, fmt.Sprintf("[Channel] send create attempt=%d msgType=%s reply=%t", attempt, msgType, false))
		resp, err := s.driver.CreateMessage(ctx, idType, id, msgType, content)
		if err != nil {
			s.logWarn(ctx, fmt.Sprintf("[Channel] send create error attempt=%d msgType=%s", attempt, msgType))
			return nil, err
		}
		return resp, nil
	}

	res, err := Retry(ctx, op, &RetryOptions{
		MaxAttempts: s.config.Outbound.Retry.MaxAttempts,
		BaseDelay:   s.config.Outbound.Retry.BaseDelayMs,
	})
	if err != nil {
		return nil, types.ClassifyError(err)
	}
	return res.(*SendResponse), nil
}

func (s *Sender) logInfo(ctx context.Context, message string) {
	if s.logger != nil {
		s.logger.Info(ctx, message)
	}
}

func (s *Sender) logWarn(ctx context.Context, message string) {
	if s.logger != nil {
		s.logger.Warn(ctx, message)
	}
}
