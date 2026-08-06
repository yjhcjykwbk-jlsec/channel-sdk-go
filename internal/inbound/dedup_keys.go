// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package inbound

import (
	"encoding/json"
	"fmt"

	"github.com/larksuite/channel-sdk-go/types"
)

func commentDedupKey(event *types.CommentEvent) string {
	return fmt.Sprintf("comment:%s:%s", event.FileToken, event.CommentID)
}

func reactionDedupKey(event *types.ReactionEvent) string {
	return fmt.Sprintf(
		"rx:%s:%s:%s:%s:%d",
		event.MessageID,
		event.UserID,
		event.ReactionType,
		event.Action,
		event.CreateTimeMs,
	)
}

func cardActionDedupKey(event *types.CardActionEvent) string {
	if event.EventID != "" {
		return event.EventID
	}
	return fmt.Sprintf(
		"card:%s:%s:%s",
		event.MessageID,
		cardActionActorID(event.Operator),
		cardActionID(event.Action),
	)
}

func cardActionActorID(operator types.CardActionOperator) string {
	if operator.OpenID != "" {
		return operator.OpenID
	}
	return operator.UserID
}

func cardActionID(action types.CardActionPayload) string {
	payload := struct {
		Tag        string                 `json:"tag,omitempty"`
		Name       string                 `json:"name,omitempty"`
		Option     string                 `json:"option,omitempty"`
		Value      map[string]interface{} `json:"value,omitempty"`
		FormValue  map[string]interface{} `json:"form_value,omitempty"`
		InputValue string                 `json:"input_value,omitempty"`
		Options    []string               `json:"options,omitempty"`
		Checked    bool                   `json:"checked,omitempty"`
	}{
		Tag:        action.Tag,
		Name:       action.Name,
		Option:     action.Option,
		Value:      action.Value,
		FormValue:  action.FormValue,
		InputValue: action.InputValue,
		Options:    action.Options,
		Checked:    action.Checked,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return action.Tag
	}
	return string(b)
}
