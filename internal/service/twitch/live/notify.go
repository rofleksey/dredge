package live

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"
)

// NotifyChatKeyword sends keyword-style notifications to all enabled providers (rules engine).
// When textTemplate is empty, uses the default Telegram line matching [$CHANNEL] $USERNAME: $TEXT.
func (r *Runtime) NotifyChatKeyword(ctx context.Context, channel, user, message, textTemplate string) {
	_ = ctx

	entries, err := r.repo.ListEnabledNotificationEntries(r.persistContext())
	if err != nil || len(entries) == 0 {
		return
	}

	payload := map[string]any{
		"type":    "keyword_match",
		"channel": channel,
		"user":    user,
		"message": truncateString(message, 400),
	}

	if strings.TrimSpace(textTemplate) != "" {
		payload["text"] = textTemplate
	}

	body, _ := json.Marshal(payload)

	for _, e := range entries {
		e := e

		go func() {
			r.notifySem <- struct{}{}

			defer func() { <-r.notifySem }()

			nctx, cancel := context.WithTimeout(r.persistContext(), 15*time.Second)
			defer cancel()

			switch e.Provider {
			case "telegram":
				r.sendTelegram(nctx, e.Settings, channel, user, message, textTemplate)
			case "webhook":
				r.postWebhook(nctx, e.Settings, body)
			default:
				r.obs.Logger.Debug("unknown notification provider", zap.String("provider", e.Provider))
			}
		}()
	}
}

// NotifyRuleText sends a rules-engine notification with only channel and rendered text (e.g. interval rules).
func (r *Runtime) NotifyRuleText(ctx context.Context, channel, text string) {
	_ = ctx

	if strings.TrimSpace(text) == "" {
		return
	}

	entries, err := r.repo.ListEnabledNotificationEntries(r.persistContext())
	if err != nil || len(entries) == 0 {
		return
	}

	payload := map[string]any{
		"type":    "rule_text",
		"channel": channel,
		"text":    text,
	}

	body, _ := json.Marshal(payload)

	for _, e := range entries {
		e := e

		go func() {
			r.notifySem <- struct{}{}

			defer func() { <-r.notifySem }()

			nctx, cancel := context.WithTimeout(r.persistContext(), 15*time.Second)
			defer cancel()

			switch e.Provider {
			case "telegram":
				r.sendTelegramRuleText(nctx, e.Settings, text)
			case "webhook":
				r.postWebhook(nctx, e.Settings, body)
			default:
				r.obs.Logger.Debug("unknown notification provider", zap.String("provider", e.Provider))
			}
		}()
	}
}

// NotifyStreamStart sends stream go-live notifications (rules engine).
// When textTemplate is empty, uses the legacy default Telegram line for stream start.
func (r *Runtime) NotifyStreamStart(ctx context.Context, channelLogin, title, textTemplate string) {
	_ = ctx

	entries, err := r.repo.ListEnabledNotificationEntries(r.persistContext())
	if err != nil || len(entries) == 0 {
		return
	}

	payload := map[string]any{
		"type":    "stream_start",
		"channel": channelLogin,
		"title":   title,
	}

	if strings.TrimSpace(textTemplate) != "" {
		payload["text"] = textTemplate
	}

	body, _ := json.Marshal(payload)

	for _, e := range entries {
		e := e

		go func() {
			r.notifySem <- struct{}{}

			defer func() { <-r.notifySem }()

			nctx, cancel := context.WithTimeout(r.persistContext(), 15*time.Second)
			defer cancel()

			switch e.Provider {
			case "telegram":
				r.sendTelegramStreamStart(nctx, e.Settings, channelLogin, title, textTemplate)
			case "webhook":
				r.postWebhook(nctx, e.Settings, body)
			default:
				r.obs.Logger.Debug("unknown notification provider", zap.String("provider", e.Provider))
			}
		}()
	}
}

// NotifyStreamEnd sends stream offline notifications (rules engine).
func (r *Runtime) NotifyStreamEnd(ctx context.Context, channelLogin, textTemplate string) {
	_ = ctx

	entries, err := r.repo.ListEnabledNotificationEntries(r.persistContext())
	if err != nil || len(entries) == 0 {
		return
	}

	payload := map[string]any{
		"type":    "stream_end",
		"channel": channelLogin,
	}

	if strings.TrimSpace(textTemplate) != "" {
		payload["text"] = textTemplate
	}

	body, _ := json.Marshal(payload)

	for _, e := range entries {
		e := e

		go func() {
			r.notifySem <- struct{}{}

			defer func() { <-r.notifySem }()

			nctx, cancel := context.WithTimeout(r.persistContext(), 15*time.Second)
			defer cancel()

			switch e.Provider {
			case "telegram":
				r.sendTelegramStreamEnd(nctx, e.Settings, channelLogin, textTemplate)
			case "webhook":
				r.postWebhook(nctx, e.Settings, body)
			default:
				r.obs.Logger.Debug("unknown notification provider", zap.String("provider", e.Provider))
			}
		}()
	}
}

func (r *Runtime) sendTelegramStreamStart(ctx context.Context, settings map[string]any, channelLogin, title, textTemplate string) {
	tok, _ := settings["bot_token"].(string)

	chatID, _ := settings["chat_id"].(string)
	if tok == "" || chatID == "" {
		r.obs.Logger.Debug("telegram notification missing bot_token or chat_id")
		return
	}

	text := fmt.Sprintf("[live] #%s started streaming", channelLogin)
	if strings.TrimSpace(title) != "" {
		text += ": " + truncateString(title, 500)
	}

	if strings.TrimSpace(textTemplate) != "" {
		text = textTemplate
	}

	form := url.Values{}
	form.Set("chat_id", chatID)
	form.Set("text", text)

	u := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", tok)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, strings.NewReader(form.Encode()))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := r.helix.HTTPClient.Do(req)
	if err != nil {
		r.obs.Logger.Debug("telegram send failed", zap.Error(err))
		return
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		r.obs.Logger.Debug("telegram send rejected", zap.Int("status", resp.StatusCode), zap.ByteString("body", b))
	}
}

func (r *Runtime) sendTelegramStreamEnd(ctx context.Context, settings map[string]any, channelLogin, textTemplate string) {
	tok, _ := settings["bot_token"].(string)

	chatID, _ := settings["chat_id"].(string)
	if tok == "" || chatID == "" {
		r.obs.Logger.Debug("telegram notification missing bot_token or chat_id")
		return
	}

	text := fmt.Sprintf("[offline] #%s stopped streaming", channelLogin)
	if strings.TrimSpace(textTemplate) != "" {
		text = textTemplate
	}

	form := url.Values{}
	form.Set("chat_id", chatID)
	form.Set("text", text)

	u := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", tok)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, strings.NewReader(form.Encode()))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := r.helix.HTTPClient.Do(req)
	if err != nil {
		r.obs.Logger.Debug("telegram send failed", zap.Error(err))
		return
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		r.obs.Logger.Debug("telegram send rejected", zap.Int("status", resp.StatusCode), zap.ByteString("body", b))
	}
}

func (r *Runtime) sendTelegramRuleText(ctx context.Context, settings map[string]any, text string) {
	tok, _ := settings["bot_token"].(string)

	chatID, _ := settings["chat_id"].(string)
	if tok == "" || chatID == "" {
		r.obs.Logger.Debug("telegram notification missing bot_token or chat_id")

		return
	}

	form := url.Values{}
	form.Set("chat_id", chatID)
	form.Set("text", text)

	u := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", tok)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, strings.NewReader(form.Encode()))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := r.helix.HTTPClient.Do(req)
	if err != nil {
		r.obs.Logger.Debug("telegram send failed", zap.Error(err))

		return
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		r.obs.Logger.Debug("telegram send rejected", zap.Int("status", resp.StatusCode), zap.ByteString("body", b))
	}
}

func (r *Runtime) sendTelegram(ctx context.Context, settings map[string]any, channel, user, msg, textTemplate string) {
	tok, _ := settings["bot_token"].(string)

	chatID, _ := settings["chat_id"].(string)
	if tok == "" || chatID == "" {
		r.obs.Logger.Debug("telegram notification missing bot_token or chat_id")
		return
	}

	text := fmt.Sprintf("[%s] %s: %s", channel, user, truncateString(msg, 3500))
	if strings.TrimSpace(textTemplate) != "" {
		text = textTemplate
	}

	form := url.Values{}
	form.Set("chat_id", chatID)
	form.Set("text", text)

	u := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", tok)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, strings.NewReader(form.Encode()))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := r.helix.HTTPClient.Do(req)
	if err != nil {
		r.obs.Logger.Debug("telegram send failed", zap.Error(err))
		return
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		r.obs.Logger.Debug("telegram send rejected", zap.Int("status", resp.StatusCode), zap.ByteString("body", b))
	}
}

func (r *Runtime) postWebhook(ctx context.Context, settings map[string]any, jsonBody []byte) {
	rawURL, _ := settings["url"].(string)
	if rawURL == "" {
		r.obs.Logger.Debug("webhook missing url")
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rawURL, bytes.NewReader(jsonBody))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")

	if h, ok := settings["headers"].(map[string]any); ok {
		for k, v := range h {
			if sv, ok := v.(string); ok {
				req.Header.Set(k, sv)
			}
		}
	}

	resp, err := r.helix.HTTPClient.Do(req)
	if err != nil {
		r.obs.Logger.Debug("webhook post failed", zap.Error(err))
		return
	}

	defer func() { _ = resp.Body.Close() }()

	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		r.obs.Logger.Debug("webhook rejected", zap.Int("status", resp.StatusCode))
	}
}
