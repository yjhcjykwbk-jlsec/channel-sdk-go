// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package e2e

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Severity string

const (
	SeverityError Severity = "error"
	SeverityWarn  Severity = "warn"
)

type Issue struct {
	Severity Severity
	Key      string
	Message  string
}

type Config struct {
	RootDir string

	AppID     string
	AppSecret string

	ReceiveOpenID     string
	GroupChatID       string
	MentionUserID     string
	DocToken          string
	AllowedGroupID    string
	BlockedGroupID    string
	AllowedUserOpenID string
	BlockedUserOpenID string
	ShareChatID       string
	StickerFileKey    string

	Media MediaPaths

	Domain      string
	OpenBaseURL string
	WSBaseURL   string

	DryRun         bool
	SkipAuto       bool
	Manual         bool
	Policy         bool
	BotAdded       bool
	WaitTimeout    time.Duration
	RequestTimeout time.Duration
}

type MediaPaths struct {
	ImageJPG   string
	ImagePNG   string
	ImageGIF   string
	ImageWEBP  string
	FilePDF    string
	FileDOCX   string
	FileXLSX   string
	FilePPTX   string
	AudioOGG   string
	VideoMP4   string
	VideoCover string
}

func LoadConfig(rootDir string, dotenvPath string, environ []string) (Config, error) {
	values := defaultValues()

	if dotenvPath != "" {
		file, err := os.Open(dotenvPath)
		if err != nil {
			if !os.IsNotExist(err) {
				return Config{}, fmt.Errorf("open dotenv: %w", err)
			}
		} else {
			defer file.Close()
			parsed, err := ParseDotenv(file)
			if err != nil {
				return Config{}, fmt.Errorf("parse dotenv: %w", err)
			}
			for key, value := range parsed {
				values[key] = value
			}
		}
	}

	for _, item := range environ {
		key, value, ok := strings.Cut(item, "=")
		if !ok || key == "" {
			continue
		}
		values[key] = value
	}

	absRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return Config{}, fmt.Errorf("resolve root dir: %w", err)
	}

	return configFromValues(absRoot, values), nil
}

func ParseDotenv(r io.Reader) (map[string]string, error) {
	out := make(map[string]string)
	scanner := bufio.NewScanner(r)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return nil, fmt.Errorf("line %d: expected KEY=VALUE", lineNo)
		}
		key = strings.TrimSpace(key)
		if key == "" {
			return nil, fmt.Errorf("line %d: key is empty", lineNo)
		}
		value = strings.TrimSpace(value)
		unquoted, err := strconv.Unquote(value)
		if err == nil {
			value = unquoted
		}
		out[key] = value
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan dotenv: %w", err)
	}
	return out, nil
}

func (c Config) Validate() []Issue {
	var issues []Issue
	for _, req := range []struct {
		key   string
		value string
	}{
		{key: "APP_ID", value: c.AppID},
		{key: "APP_SECRET", value: c.AppSecret},
		{key: "CHANNEL_E2E_RECEIVE_OPEN_ID", value: c.ReceiveOpenID},
	} {
		if strings.TrimSpace(req.value) == "" {
			issues = append(issues, Issue{
				Severity: SeverityError,
				Key:      req.key,
				Message:  req.key + " is required",
			})
		}
	}

	for key, path := range c.mediaPathVars() {
		if strings.TrimSpace(path) == "" {
			continue
		}
		if err := ValidateFixturePath(c.RootDir, path); err != nil {
			issues = append(issues, Issue{
				Severity: SeverityError,
				Key:      key,
				Message:  err.Error(),
			})
		}
	}
	return issues
}

func ErrorIssues(issues []Issue) []Issue {
	var out []Issue
	for _, issue := range issues {
		if issue.Severity == SeverityError {
			out = append(out, issue)
		}
	}
	return out
}

func ValidateFixturePath(rootDir string, sourcePath string) error {
	for _, segment := range strings.Split(filepath.ToSlash(sourcePath), "/") {
		if segment == ".." {
			return fmt.Errorf("path traversal is not allowed")
		}
	}

	cleaned := filepath.Clean(sourcePath)
	if !filepath.IsAbs(cleaned) {
		cleaned = filepath.Join(rootDir, cleaned)
	}

	absRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return fmt.Errorf("resolve root dir: %w", err)
	}
	absPath, err := filepath.Abs(cleaned)
	if err != nil {
		return fmt.Errorf("resolve fixture path: %w", err)
	}
	rel, err := filepath.Rel(absRoot, absPath)
	if err != nil {
		return fmt.Errorf("check fixture path: %w", err)
	}
	if rel == ".." || strings.HasPrefix(filepath.ToSlash(rel), "../") {
		return fmt.Errorf("fixture path must stay under repository root")
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("fixture file is not accessible: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("fixture path points to a directory")
	}
	return nil
}

func (c Config) ResolveFixturePath(sourcePath string) (string, error) {
	if err := ValidateFixturePath(c.RootDir, sourcePath); err != nil {
		return "", err
	}
	if filepath.IsAbs(sourcePath) {
		return filepath.Clean(sourcePath), nil
	}
	return filepath.Join(c.RootDir, filepath.Clean(sourcePath)), nil
}

func (c Config) mediaPathVars() map[string]string {
	return map[string]string{
		"CHANNEL_E2E_IMAGE_JPG":         c.Media.ImageJPG,
		"CHANNEL_E2E_IMAGE_PNG":         c.Media.ImagePNG,
		"CHANNEL_E2E_IMAGE_GIF":         c.Media.ImageGIF,
		"CHANNEL_E2E_IMAGE_WEBP":        c.Media.ImageWEBP,
		"CHANNEL_E2E_FILE_PDF":          c.Media.FilePDF,
		"CHANNEL_E2E_FILE_DOCX":         c.Media.FileDOCX,
		"CHANNEL_E2E_FILE_XLSX":         c.Media.FileXLSX,
		"CHANNEL_E2E_FILE_PPTX":         c.Media.FilePPTX,
		"CHANNEL_E2E_AUDIO_OGG":         c.Media.AudioOGG,
		"CHANNEL_E2E_VIDEO_MP4":         c.Media.VideoMP4,
		"CHANNEL_E2E_VIDEO_COVER_IMAGE": c.Media.VideoCover,
	}
}

func configFromValues(rootDir string, values map[string]string) Config {
	return Config{
		RootDir: rootDir,

		AppID:     values["APP_ID"],
		AppSecret: values["APP_SECRET"],

		ReceiveOpenID:     values["CHANNEL_E2E_RECEIVE_OPEN_ID"],
		GroupChatID:       values["CHANNEL_E2E_GROUP_CHAT_ID"],
		MentionUserID:     values["CHANNEL_E2E_MENTION_USER_OPEN_ID"],
		DocToken:          values["CHANNEL_E2E_DOC_TOKEN"],
		AllowedGroupID:    values["CHANNEL_E2E_ALLOWED_GROUP_CHAT_ID"],
		BlockedGroupID:    values["CHANNEL_E2E_BLOCKED_GROUP_CHAT_ID"],
		AllowedUserOpenID: values["CHANNEL_E2E_ALLOWED_USER_OPEN_ID"],
		BlockedUserOpenID: values["CHANNEL_E2E_BLOCKED_USER_OPEN_ID"],
		ShareChatID:       values["CHANNEL_E2E_SHARE_CHAT_ID"],
		StickerFileKey:    values["CHANNEL_E2E_STICKER_FILE_KEY"],

		Media: MediaPaths{
			ImageJPG:   values["CHANNEL_E2E_IMAGE_JPG"],
			ImagePNG:   values["CHANNEL_E2E_IMAGE_PNG"],
			ImageGIF:   values["CHANNEL_E2E_IMAGE_GIF"],
			ImageWEBP:  values["CHANNEL_E2E_IMAGE_WEBP"],
			FilePDF:    values["CHANNEL_E2E_FILE_PDF"],
			FileDOCX:   values["CHANNEL_E2E_FILE_DOCX"],
			FileXLSX:   values["CHANNEL_E2E_FILE_XLSX"],
			FilePPTX:   values["CHANNEL_E2E_FILE_PPTX"],
			AudioOGG:   values["CHANNEL_E2E_AUDIO_OGG"],
			VideoMP4:   values["CHANNEL_E2E_VIDEO_MP4"],
			VideoCover: values["CHANNEL_E2E_VIDEO_COVER_IMAGE"],
		},

		Domain:      values["CHANNEL_E2E_DOMAIN"],
		OpenBaseURL: values["CHANNEL_E2E_OPEN_BASE_URL"],
		WSBaseURL:   values["CHANNEL_E2E_WS_BASE_URL"],

		DryRun:         boolValue(values["CHANNEL_E2E_DRY_RUN"]),
		SkipAuto:       boolValue(values["CHANNEL_E2E_SKIP_AUTO"]),
		Manual:         boolValue(values["CHANNEL_E2E_MANUAL"]),
		Policy:         boolValue(values["CHANNEL_E2E_ENABLE_POLICY"]),
		BotAdded:       boolValue(values["CHANNEL_E2E_ENABLE_BOT_ADDED"]),
		WaitTimeout:    durationSeconds(values["CHANNEL_E2E_WAIT_SECONDS"], 180*time.Second),
		RequestTimeout: durationSeconds(values["CHANNEL_E2E_REQUEST_TIMEOUT_SECONDS"], 15*time.Second),
	}
}

func defaultValues() map[string]string {
	return map[string]string{
		"CHANNEL_E2E_IMAGE_JPG":               "./testdata/e2e/image.jpg",
		"CHANNEL_E2E_IMAGE_PNG":               "./testdata/e2e/image.png",
		"CHANNEL_E2E_IMAGE_GIF":               "./testdata/e2e/image.gif",
		"CHANNEL_E2E_IMAGE_WEBP":              "./testdata/e2e/image.webp",
		"CHANNEL_E2E_FILE_PDF":                "./testdata/e2e/file.pdf",
		"CHANNEL_E2E_FILE_DOCX":               "./testdata/e2e/file.docx",
		"CHANNEL_E2E_FILE_XLSX":               "./testdata/e2e/file.xlsx",
		"CHANNEL_E2E_FILE_PPTX":               "./testdata/e2e/file.pptx",
		"CHANNEL_E2E_AUDIO_OGG":               "./testdata/e2e/audio.ogg",
		"CHANNEL_E2E_VIDEO_MP4":               "./testdata/e2e/video.mp4",
		"CHANNEL_E2E_VIDEO_COVER_IMAGE":       "./testdata/e2e/video-cover.jpg",
		"CHANNEL_E2E_WAIT_SECONDS":            "180",
		"CHANNEL_E2E_REQUEST_TIMEOUT_SECONDS": "15",
	}
}

func boolValue(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}

func durationSeconds(value string, fallback time.Duration) time.Duration {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	seconds, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || seconds <= 0 {
		return fallback
	}
	return time.Duration(seconds) * time.Second
}
