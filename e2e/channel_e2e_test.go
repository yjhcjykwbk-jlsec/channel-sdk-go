// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

//go:build e2e

package e2e_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	channel "github.com/larksuite/channel-sdk-go"
	internale2e "github.com/larksuite/channel-sdk-go/internal/e2e"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

func TestChannelE2E(t *testing.T) {
	root, err := findRepoRoot()
	if err != nil {
		t.Fatalf("find repo root: %v", err)
	}

	cfg, err := internale2e.LoadConfig(root, filepath.Join(root, ".env"), os.Environ())
	if err != nil {
		t.Fatalf("load e2e config: %v", err)
	}

	issues := cfg.Validate()
	if errs := internale2e.ErrorIssues(issues); len(errs) > 0 {
		for _, issue := range errs {
			t.Errorf("%s: %s", issue.Key, issue.Message)
		}
		t.FailNow()
	}

	plan := internale2e.BuildPlan(cfg)
	logPlan(t, plan)
	if cfg.DryRun {
		t.Log("CHANNEL_E2E_DRY_RUN is enabled; validated config and case plan without network calls")
		return
	}

	ch, err := channel.New(cfg.AppID, cfg.AppSecret, channelOptions(cfg)...)
	if err != nil {
		t.Fatalf("create channel: %v", err)
	}
	t.Cleanup(func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := ch.Stop(stopCtx); err != nil {
			t.Logf("stop channel: %v", err)
		}
	})

	trace := fmt.Sprintf("channel-sdk-go-e2e-%d", time.Now().UTC().UnixNano())
	if cfg.SkipAuto {
		t.Log("CHANNEL_E2E_SKIP_AUTO is enabled; skipped automatic network cases")
	} else {
		runAutomaticCases(t, cfg, plan, ch, trace)
	}
	if cfg.Manual {
		runManualCases(t, cfg, plan, ch, trace)
	}
}

func channelOptions(cfg internale2e.Config) []channel.Option {
	opts := []channel.Option{
		channel.WithLogLevel(larkcore.LogLevelWarn),
		channel.WithReqTimeout(cfg.RequestTimeout),
	}
	if cfg.Domain != "" {
		opts = append(opts, channel.WithDomain(cfg.Domain))
	}
	if cfg.OpenBaseURL != "" {
		opts = append(opts, channel.WithOpenBaseURL(cfg.OpenBaseURL))
	}
	if cfg.WSBaseURL != "" {
		opts = append(opts, channel.WithWebSocketDomain(cfg.WSBaseURL))
	}
	return opts
}

func runAutomaticCases(t *testing.T, cfg internale2e.Config, plan internale2e.Plan, ch channel.Channel, trace string) {
	t.Helper()

	run := func(id string, fn func(context.Context) error) {
		t.Helper()
		if !plan.HasReadyCase(id) {
			t.Logf("skip %s: %s", id, skippedReason(plan, id))
			return
		}
		t.Run(id, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), cfg.RequestTimeout)
			defer cancel()
			if err := fn(ctx); err != nil {
				t.Fatalf("%s failed: %v", id, err)
			}
		})
	}

	run("send.text", func(ctx context.Context) error {
		_, err := sendAndRequireMessage(ctx, ch, &channel.SendInput{
			ReceiveID: cfg.ReceiveOpenID,
			MsgType:   "text",
			Text:      "Channel SDK Go E2E text " + trace,
		})
		return err
	})

	run("lifecycle.bot_identity", func(ctx context.Context) error {
		identity := ch.GetBotIdentity(ctx)
		if identity == nil || identity.OpenID == "" {
			return fmt.Errorf("bot identity is empty")
		}
		return nil
	})

	run("lifecycle.stop", func(ctx context.Context) error {
		return ch.Stop(ctx)
	})

	run("raw.chat_info", func(ctx context.Context) error {
		resp, err := ch.RawClient().Im.V1.Chat.Get(ctx, larkim.NewGetChatReqBuilder().
			ChatId(cfg.GroupChatID).
			Build())
		if err != nil {
			return err
		}
		if !resp.Success() {
			return &larkcore.CodeError{Code: resp.Code, Msg: resp.Msg}
		}
		if resp.Data == nil || resp.Data.Name == nil || *resp.Data.Name == "" {
			return fmt.Errorf("chat info response missing name")
		}
		return nil
	})

	run("raw.message_list", func(ctx context.Context) error {
		end := time.Now().Unix()
		start := end - 24*60*60
		resp, err := ch.RawClient().Im.V1.Message.List(ctx, larkim.NewListMessageReqBuilder().
			ContainerIdType("chat").
			ContainerId(cfg.GroupChatID).
			StartTime(fmt.Sprint(start)).
			EndTime(fmt.Sprint(end)).
			SortType(larkim.ReadHistoryMessageV1SortTypeByCreateTimeDesc).
			PageSize(10).
			Build())
		if err != nil {
			return err
		}
		if !resp.Success() {
			return &larkcore.CodeError{Code: resp.Code, Msg: resp.Msg}
		}
		if resp.Data == nil {
			return fmt.Errorf("message list response missing data")
		}
		return nil
	})

	run("send.markdown", func(ctx context.Context) error {
		_, err := sendAndRequireMessage(ctx, ch, &channel.SendInput{
			ReceiveID: cfg.ReceiveOpenID,
			Title:     "Channel SDK Go E2E",
			Markdown:  "### Channel SDK Go E2E\n\n- case: send.markdown\n- trace: `" + trace + "`",
		})
		return err
	})

	run("send.long_markdown", func(ctx context.Context) error {
		res, err := sendAndRequireMessage(ctx, ch, &channel.SendInput{
			ReceiveID: cfg.ReceiveOpenID,
			Title:     "Channel SDK Go E2E Long Markdown",
			Markdown:  longMarkdown(trace),
		})
		if err != nil {
			return err
		}
		if len(res.ChunkIDs) < 2 {
			return fmt.Errorf("expected long markdown to split into chunks, got %d chunk ids", len(res.ChunkIDs))
		}
		return nil
	})

	run("send.post", func(ctx context.Context) error {
		_, err := sendAndRequireMessage(ctx, ch, &channel.SendInput{
			ReceiveID: cfg.ReceiveOpenID,
			Post:      postJSON(trace),
		})
		return err
	})

	run("send.mention", func(ctx context.Context) error {
		_, err := sendAndRequireMessage(ctx, ch, &channel.SendInput{
			ChatID: cfg.GroupChatID,
			Text:   "Channel SDK Go E2E mention " + trace,
			Mentions: []channel.Mention{
				{
					UserID: cfg.MentionUserID,
					Name:   "E2E",
				},
			},
		})
		return err
	})

	run("send.group_text", func(ctx context.Context) error {
		_, err := sendAndRequireMessage(ctx, ch, &channel.SendInput{
			ChatID:  cfg.GroupChatID,
			MsgType: "text",
			Text:    "Channel SDK Go E2E group text " + trace,
		})
		return err
	})

	for _, imageCase := range []struct {
		id   string
		path string
	}{
		{id: "send.image_jpg", path: cfg.Media.ImageJPG},
		{id: "send.image_png", path: cfg.Media.ImagePNG},
		{id: "send.image_gif", path: cfg.Media.ImageGIF},
		{id: "send.image_webp", path: cfg.Media.ImageWEBP},
	} {
		imageCase := imageCase
		run(imageCase.id, func(ctx context.Context) error {
			path, err := cfg.ResolveFixturePath(imageCase.path)
			if err != nil {
				return err
			}
			input := &channel.SendInput{
				ReceiveID: cfg.ReceiveOpenID,
				ImagePath: path,
			}
			if _, err := sendAndRequireMessage(ctx, ch, input); err != nil {
				return err
			}
			return downloadAndRequireBytes(ctx, ch, input.ImageKey, "image")
		})
	}

	run("send.file_pdf", func(ctx context.Context) error {
		path, err := cfg.ResolveFixturePath(cfg.Media.FilePDF)
		if err != nil {
			return err
		}
		input := &channel.SendInput{
			ReceiveID: cfg.ReceiveOpenID,
			FilePath:  path,
		}
		if _, err := sendAndRequireMessage(ctx, ch, input); err != nil {
			return err
		}
		return downloadAndRequireBytes(ctx, ch, input.FileKey, "file")
	})

	run("send.file_office", func(ctx context.Context) error {
		for _, item := range []struct {
			name string
			path string
		}{
			{name: "docx", path: cfg.Media.FileDOCX},
			{name: "xlsx", path: cfg.Media.FileXLSX},
			{name: "pptx", path: cfg.Media.FilePPTX},
		} {
			path, err := cfg.ResolveFixturePath(item.path)
			if err != nil {
				return fmt.Errorf("%s fixture: %w", item.name, err)
			}
			input := &channel.SendInput{
				ReceiveID: cfg.ReceiveOpenID,
				FilePath:  path,
			}
			if _, err := sendAndRequireMessage(ctx, ch, input); err != nil {
				return fmt.Errorf("send %s: %w", item.name, err)
			}
		}
		return nil
	})

	run("send.audio", func(ctx context.Context) error {
		path, err := cfg.ResolveFixturePath(cfg.Media.AudioOGG)
		if err != nil {
			return err
		}
		_, err = sendAndRequireMessage(ctx, ch, &channel.SendInput{
			ReceiveID: cfg.ReceiveOpenID,
			Media: &channel.UploadInput{
				Kind:       channel.MediaKindAudio,
				SourcePath: path,
				FileName:   "channel-e2e-audio.ogg",
			},
		})
		return err
	})

	run("send.video", func(ctx context.Context) error {
		path, err := cfg.ResolveFixturePath(cfg.Media.VideoMP4)
		if err != nil {
			return err
		}
		_, err = sendAndRequireMessage(ctx, ch, &channel.SendInput{
			ReceiveID: cfg.ReceiveOpenID,
			Media: &channel.UploadInput{
				Kind:       channel.MediaKindVideo,
				SourcePath: path,
				FileName:   "channel-e2e-video.mp4",
			},
		})
		return err
	})

	run("send.share_chat", func(ctx context.Context) error {
		_, err := sendAndRequireMessage(ctx, ch, &channel.SendInput{
			ReceiveID:   cfg.ReceiveOpenID,
			ShareChatID: cfg.ShareChatID,
		})
		return err
	})

	run("send.share_user", func(ctx context.Context) error {
		_, err := sendAndRequireMessage(ctx, ch, &channel.SendInput{
			ReceiveID:   cfg.ReceiveOpenID,
			ShareUserID: cfg.MentionUserID,
		})
		return err
	})

	run("send.sticker", func(ctx context.Context) error {
		_, err := sendAndRequireMessage(ctx, ch, &channel.SendInput{
			ReceiveID:      cfg.ReceiveOpenID,
			StickerFileKey: cfg.StickerFileKey,
		})
		return err
	})

	run("send.card", func(ctx context.Context) error {
		card, err := testCard("send.card", trace)
		if err != nil {
			return err
		}
		_, err = sendAndRequireMessage(ctx, ch, &channel.SendInput{
			ReceiveID: cfg.ReceiveOpenID,
			Card:      card,
		})
		return err
	})

	run("reply.text", func(ctx context.Context) error {
		base, err := sendAndRequireMessage(ctx, ch, &channel.SendInput{
			ReceiveID: cfg.ReceiveOpenID,
			MsgType:   "text",
			Text:      "Channel SDK Go E2E reply base " + trace,
		})
		if err != nil {
			return err
		}
		_, err = sendAndRequireMessage(ctx, ch, &channel.SendInput{
			ReceiveID:      cfg.ReceiveOpenID,
			ReplyMessageID: base.MessageID,
			MsgType:        "text",
			Text:           "Channel SDK Go E2E text reply " + trace,
		})
		return err
	})

	run("reply.markdown", func(ctx context.Context) error {
		base, err := sendAndRequireMessage(ctx, ch, &channel.SendInput{
			ReceiveID: cfg.ReceiveOpenID,
			MsgType:   "text",
			Text:      "Channel SDK Go E2E markdown reply base " + trace,
		})
		if err != nil {
			return err
		}
		_, err = sendAndRequireMessage(ctx, ch, &channel.SendInput{
			ReceiveID:      cfg.ReceiveOpenID,
			ReplyMessageID: base.MessageID,
			Title:          "Channel SDK Go E2E Reply",
			Markdown:       "reply markdown `" + trace + "`",
		})
		return err
	})

	run("reply.image", func(ctx context.Context) error {
		base, err := sendAndRequireMessage(ctx, ch, &channel.SendInput{
			ReceiveID: cfg.ReceiveOpenID,
			MsgType:   "text",
			Text:      "Channel SDK Go E2E image reply base " + trace,
		})
		if err != nil {
			return err
		}
		path, err := cfg.ResolveFixturePath(cfg.Media.ImageJPG)
		if err != nil {
			return err
		}
		_, err = sendAndRequireMessage(ctx, ch, &channel.SendInput{
			ReceiveID:      cfg.ReceiveOpenID,
			ReplyMessageID: base.MessageID,
			ImagePath:      path,
		})
		return err
	})

	run("stream.markdown", func(ctx context.Context) error {
		stream, err := ch.Stream(ctx, &channel.SendInput{
			ReceiveID: cfg.ReceiveOpenID,
			Title:     "Channel SDK Go E2E Stream",
			Markdown:  "stream start `" + trace + "`",
		})
		if err != nil {
			return err
		}
		if err := stream.Append(ctx, "\n\nstream append `"+trace+"`"); err != nil {
			return err
		}
		if err := stream.Flush(ctx); err != nil {
			return err
		}
		return stream.Close(ctx)
	})

	run("stream.card_update", func(ctx context.Context) error {
		initial, err := testCard("stream.card_update.initial", trace)
		if err != nil {
			return err
		}
		stream, err := ch.Stream(ctx, &channel.SendInput{
			ReceiveID: cfg.ReceiveOpenID,
			Card:      initial,
		})
		if err != nil {
			return err
		}
		updated, err := testCard("stream.card_update.updated", trace)
		if err != nil {
			return err
		}
		if err := stream.UpdateCard(ctx, updated); err != nil {
			return err
		}
		return stream.Close(ctx)
	})
}

func runManualCases(t *testing.T, cfg internale2e.Config, plan internale2e.Plan, ch channel.Channel, trace string) {
	t.Helper()

	if cfg.Policy {
		applyPolicyConfig(ch, cfg)
	}

	tracker := newEventTracker()
	ready := make(chan struct{})
	var readyOnce sync.Once
	targets := newManualTargets()

	ch.OnReady(func() {
		readyOnce.Do(func() {
			close(ready)
		})
	})
	ch.OnError(func(err error) {
		tracker.mark("ws.error", err.Error())
	})
	ch.OnMessage(func(ctx context.Context, msg *channel.NormalizedMessage) error {
		if strings.Contains(msg.Content, trace) {
			tracker.mark("event.message", "received manual trace message")
		}
		if cfg.Policy && msg.ChatID == cfg.AllowedGroupID {
			tracker.mark("policy.allowed_group", "received allowed group message")
		}
		if cfg.Policy && msg.UserID == cfg.AllowedUserOpenID {
			tracker.mark("policy.allowed_user", "received allowed direct message")
		}
		return nil
	})
	ch.OnReaction(func(ctx context.Context, event *channel.ReactionEvent) error {
		reactionTarget := targets.reactionMessageID()
		if reactionTarget == "" || event.MessageID != reactionTarget {
			return nil
		}
		if event.Action == "added" {
			tracker.mark("event.reaction_created", "reaction created on target message")
		}
		if event.Action == "removed" {
			tracker.mark("event.reaction_deleted", "reaction removed on target message")
		}
		return nil
	})
	ch.OnCardAction(func(ctx context.Context, event *channel.CardActionEvent) error {
		caseValue, hasCase := event.Action.Value["case"].(string)
		traceValue, hasTrace := event.Action.Value["trace"].(string)
		if hasCase && hasTrace && caseValue == "event.card_action" && traceValue == trace {
			tracker.mark("event.card_action", "card action callback received")
		}
		return nil
	})
	ch.OnComment(func(ctx context.Context, event *channel.CommentEvent) error {
		if cfg.DocToken == "" || event.FileToken == cfg.DocToken {
			tracker.mark("event.comment", "document comment event received")
		}
		return nil
	})
	ch.OnBotAdded(func(ctx context.Context, event *channel.BotAddedEvent) error {
		if cfg.GroupChatID == "" || event.ChatID == cfg.GroupChatID {
			tracker.mark("event.bot_added", "bot added event received")
		}
		return nil
	})
	ch.OnReject(func(ctx context.Context, event *channel.RejectEvent) error {
		if cfg.Policy && cfg.BlockedGroupID != "" && event.ChatID == cfg.BlockedGroupID {
			tracker.mark("policy.blocked_group", "blocked group message rejected")
		}
		if cfg.Policy && cfg.BlockedUserOpenID != "" && event.SenderID == cfg.BlockedUserOpenID {
			tracker.mark("policy.blocked_user", "blocked direct-message sender rejected")
		}
		return nil
	})

	manualCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- ch.Start(manualCtx)
	}()

	select {
	case <-ready:
	case err := <-errCh:
		if err != nil {
			t.Fatalf("start websocket: %v", err)
		}
		t.Fatal("websocket stopped before ready")
	case <-time.After(cfg.WaitTimeout):
		t.Fatalf("websocket was not ready within %s", cfg.WaitTimeout)
	}

	if plan.HasReadyCase("event.reaction_created") || plan.HasReadyCase("event.reaction_deleted") {
		sendCtx, sendCancel := context.WithTimeout(manualCtx, cfg.RequestTimeout)
		res, err := sendAndRequireMessage(sendCtx, ch, &channel.SendInput{
			ReceiveID: cfg.ReceiveOpenID,
			MsgType:   "text",
			Text:      "Channel SDK Go E2E reaction target " + trace,
		})
		sendCancel()
		if err != nil {
			t.Fatalf("send reaction target: %v", err)
		}
		targets.setReactionMessageID(res.MessageID)
	}
	if plan.HasReadyCase("event.card_action") {
		card, err := testCard("event.card_action", trace)
		if err != nil {
			t.Fatalf("build manual card: %v", err)
		}
		sendCtx, sendCancel := context.WithTimeout(manualCtx, cfg.RequestTimeout)
		_, err = sendAndRequireMessage(sendCtx, ch, &channel.SendInput{
			ReceiveID: cfg.ReceiveOpenID,
			Card:      card,
		})
		sendCancel()
		if err != nil {
			t.Fatalf("send manual card: %v", err)
		}
	}

	logManualInstructions(t, cfg, plan, trace)
	for _, c := range plan.ReadyCases() {
		if c.Mode != internale2e.CaseModeManual {
			continue
		}
		waitCtx, waitCancel := context.WithTimeout(context.Background(), cfg.WaitTimeout)
		detail, err := tracker.wait(waitCtx, c.ID)
		waitCancel()
		if err != nil {
			if wsDetail, ok := tracker.get("ws.error"); ok {
				t.Fatalf("%s not observed: %v; websocket error: %s", c.ID, err, wsDetail)
			}
			t.Fatalf("%s not observed: %v", c.ID, err)
		}
		t.Logf("observed %s: %s", c.ID, detail)
	}

	cancel()
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer stopCancel()
	if err := ch.Stop(stopCtx); err != nil {
		t.Fatalf("stop websocket: %v", err)
	}
	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, context.Canceled) {
			t.Logf("websocket stopped with: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("websocket did not stop within 5s")
	}
}

func applyPolicyConfig(ch channel.Channel, cfg internale2e.Config) {
	ch.UpdatePolicy(channel.PolicyConfig{
		GroupAllowlist: uniqueNonEmpty([]string{
			cfg.GroupChatID,
			cfg.AllowedGroupID,
		}),
		DMMode: "allowlist",
		DMAllowlist: uniqueNonEmpty([]string{
			cfg.ReceiveOpenID,
			cfg.AllowedUserOpenID,
		}),
	})
}

func sendAndRequireMessage(ctx context.Context, ch channel.Channel, input *channel.SendInput) (*channel.SendResult, error) {
	res, err := ch.Send(ctx, input)
	if err != nil {
		return nil, err
	}
	if res == nil || res.MessageID == "" {
		return nil, fmt.Errorf("send returned empty message id")
	}
	return res, nil
}

func downloadAndRequireBytes(ctx context.Context, ch channel.Channel, fileKey string, mediaType string) error {
	if fileKey == "" {
		return fmt.Errorf("file key is empty")
	}
	data, err := ch.DownloadFile(ctx, fileKey, mediaType)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return fmt.Errorf("download returned empty body")
	}
	return nil
}

func testCard(caseID string, trace string) (string, error) {
	card := map[string]interface{}{
		"config": map[string]interface{}{
			"wide_screen_mode": true,
		},
		"header": map[string]interface{}{
			"title": map[string]string{
				"tag":     "plain_text",
				"content": "Channel SDK Go E2E",
			},
		},
		"elements": []interface{}{
			map[string]interface{}{
				"tag": "div",
				"text": map[string]string{
					"tag":     "lark_md",
					"content": "**Channel SDK Go E2E**\ncase: `" + caseID + "`\ntrace: `" + trace + "`",
				},
			},
			map[string]interface{}{
				"tag": "action",
				"actions": []interface{}{
					map[string]interface{}{
						"tag":  "button",
						"name": "channel_e2e_button",
						"text": map[string]string{
							"tag":     "plain_text",
							"content": "E2E OK",
						},
						"type": "primary",
						"value": map[string]string{
							"case":  caseID,
							"trace": trace,
						},
					},
				},
			},
		},
	}
	data, err := json.Marshal(card)
	if err != nil {
		return "", fmt.Errorf("marshal test card: %w", err)
	}
	return string(data), nil
}

func postJSON(trace string) string {
	data, _ := json.Marshal(map[string]interface{}{
		"zh_cn": map[string]interface{}{
			"title": "Channel SDK Go E2E Post",
			"content": [][]map[string]string{
				{
					{"tag": "text", "text": "post message " + trace},
				},
			},
		},
	})
	return string(data)
}

func longMarkdown(trace string) string {
	return strings.Join([]string{
		"# Channel SDK Go E2E Long Markdown",
		"",
		"trace: `" + trace + "`",
		"",
		"```go",
		"func main() {",
		`	fmt.Println("channel e2e")`,
		"}",
		"```",
		"",
		strings.Repeat("long markdown paragraph "+trace+"\n", 260),
		"",
		"```json",
		`{"trace":"` + trace + `"}`,
		"```",
	}, "\n")
}

type eventTracker struct {
	mu      sync.Mutex
	seen    map[string]string
	waiters map[string]chan string
}

type manualTargets struct {
	mu                     sync.RWMutex
	reactionMessageIDValue string
}

func newManualTargets() *manualTargets {
	return &manualTargets{}
}

func (t *manualTargets) setReactionMessageID(messageID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.reactionMessageIDValue = messageID
}

func (t *manualTargets) reactionMessageID() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.reactionMessageIDValue
}

func newEventTracker() *eventTracker {
	return &eventTracker{
		seen:    make(map[string]string),
		waiters: make(map[string]chan string),
	}
}

func (t *eventTracker) mark(name string, detail string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, exists := t.seen[name]; !exists {
		t.seen[name] = detail
	}
	if ch, ok := t.waiters[name]; ok {
		select {
		case ch <- detail:
		default:
		}
	}
}

func (t *eventTracker) wait(ctx context.Context, name string) (string, error) {
	t.mu.Lock()
	if detail, ok := t.seen[name]; ok {
		t.mu.Unlock()
		return detail, nil
	}
	ch, ok := t.waiters[name]
	if !ok {
		ch = make(chan string, 1)
		t.waiters[name] = ch
	}
	t.mu.Unlock()

	select {
	case detail := <-ch:
		return detail, nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func (t *eventTracker) get(name string) (string, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	detail, ok := t.seen[name]
	return detail, ok
}

func logPlan(t *testing.T, plan internale2e.Plan) {
	t.Helper()
	for _, c := range plan.Cases {
		if c.Status == internale2e.CaseStatusReady {
			t.Logf("case ready: %s (%s)", c.ID, c.Mode)
			continue
		}
		t.Logf("case skipped: %s (%s): %s", c.ID, c.Mode, c.Reason)
	}
}

func logManualInstructions(t *testing.T, cfg internale2e.Config, plan internale2e.Plan, trace string) {
	t.Helper()
	t.Log("manual E2E instructions:")
	if plan.HasReadyCase("event.message") {
		t.Logf("- send this exact text to the bot: Channel SDK Go E2E manual message %s", trace)
	}
	if plan.HasReadyCase("event.reaction_created") || plan.HasReadyCase("event.reaction_deleted") {
		t.Log("- add a reaction to the message whose text starts with 'Channel SDK Go E2E reaction target', then remove that reaction")
	}
	if plan.HasReadyCase("event.card_action") {
		t.Log("- click the E2E OK button on the Channel SDK Go E2E card")
	}
	if plan.HasReadyCase("event.comment") {
		t.Logf("- add a test comment in the .env document token %s and mention the bot if the case requires it", mask(cfg.DocToken))
	}
	if plan.HasReadyCase("event.bot_added") {
		t.Log("- add the bot to the configured test group; this case is opt-in via CHANNEL_E2E_ENABLE_BOT_ADDED=1")
	}
	if cfg.Policy {
		if plan.HasReadyCase("policy.allowed_group") {
			t.Log("- send any test message in CHANNEL_E2E_ALLOWED_GROUP_CHAT_ID")
		}
		if plan.HasReadyCase("policy.blocked_group") {
			t.Log("- send any test message in CHANNEL_E2E_BLOCKED_GROUP_CHAT_ID and expect a reject event")
		}
		if plan.HasReadyCase("policy.allowed_user") {
			t.Log("- send any direct test message from CHANNEL_E2E_ALLOWED_USER_OPEN_ID")
		}
		if plan.HasReadyCase("policy.blocked_user") {
			t.Log("- send any direct test message from CHANNEL_E2E_BLOCKED_USER_OPEN_ID and expect a reject event")
		}
	}
}

func skippedReason(plan internale2e.Plan, id string) string {
	for _, c := range plan.Cases {
		if c.ID == id && c.Status == internale2e.CaseStatusSkipped {
			return c.Reason
		}
	}
	return "case is not ready"
}

func findRepoRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		} else if !os.IsNotExist(err) {
			return "", err
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found from %s", wd)
		}
		dir = parent
	}
}

func uniqueNonEmpty(values []string) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func mask(value string) string {
	if value == "" {
		return "<empty>"
	}
	if len(value) <= 8 {
		return strings.Repeat("*", len(value))
	}
	return value[:4] + strings.Repeat("*", len(value)-8) + value[len(value)-4:]
}
