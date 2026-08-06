// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package channel

import (
	"github.com/larksuite/channel-sdk-go/types"
	lark "github.com/larksuite/oapi-sdk-go/v3"
)

type Channel interface {
	types.Channel
	RawClient() *lark.Client
}

type SendInput = types.SendInput
type SendResult = types.SendResult
type StreamController = types.StreamController
type NormalizedMessage = types.NormalizedMessage
type Mention = types.Mention
type Resource = types.Resource
type UploadInput = types.UploadInput
type UploadResult = types.UploadResult
type MediaKind = types.MediaKind

const (
	MediaKindImage = types.MediaKindImage
	MediaKindFile  = types.MediaKindFile
	MediaKindAudio = types.MediaKindAudio
	MediaKindVideo = types.MediaKindVideo
)

type ReactionEvent = types.ReactionEvent
type OperatorInfo = types.OperatorInfo
type CommentEvent = types.CommentEvent
type BotAddedEvent = types.BotAddedEvent
type CardActionEvent = types.CardActionEvent
type CardActionOperator = types.CardActionOperator
type CardActionPayload = types.CardActionPayload
type CardActionContext = types.CardActionContext
type RejectEvent = types.RejectEvent

type SafetyConfig = types.SafetyConfig
type PolicyConfig = types.PolicyConfig
type OutboundConfig = types.OutboundConfig
type BotIdentityCacheConfig = types.BotIdentityCacheConfig
type RejectReason = types.RejectReason
type PolicyDecision = types.PolicyDecision
type BotIdentity = types.BotIdentity
type BatchConfig = types.BatchConfig
type BatchedDispatch = types.BatchedDispatch

const (
	RejectReasonGroupNotAllowed  = types.RejectReasonGroupNotAllowed
	RejectReasonNoMention        = types.RejectReasonNoMention
	RejectReasonMentionAll       = types.RejectReasonMentionAll
	RejectReasonDMDisabled       = types.RejectReasonDMDisabled
	RejectReasonSenderNotAllowed = types.RejectReasonSenderNotAllowed
)

var DefaultBatchConfig = types.DefaultBatchConfig
var ClassifyError = types.ClassifyError
var IsRetryable = types.IsRetryable
var IsFormatError = types.IsFormatError

type FeishuChannelErrorCode = types.FeishuChannelErrorCode
type FeishuChannelError = types.FeishuChannelError

const (
	ErrCodeTargetRevoked    = types.ErrCodeTargetRevoked
	ErrCodePermissionDenied = types.ErrCodePermissionDenied
	ErrCodeFormatError      = types.ErrCodeFormatError
	ErrCodeRateLimited      = types.ErrCodeRateLimited
	ErrCodeSSRFBlocked      = types.ErrCodeSSRFBlocked
	ErrCodeSendTimeout      = types.ErrCodeSendTimeout
	ErrCodeUnknown          = types.ErrCodeUnknown
)
