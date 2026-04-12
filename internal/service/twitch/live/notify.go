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

func (r *Runtime) dispatchRuleHitNotifications(ctx context.Context, channel, user, message string) {
	_ = ctx

	entries, err := r.repo.ListEnabledNotificationEntries(r.persistContext())
	if err != nil || len(entries) == 0 {
		return
	}

	payload := map[string]any{
		"channel": channel,
		"user":    user,
		"message": truncateString(message, 400),
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
				r.sendTelegram(nctx, e.Settings, channel, user, message)
			case "webhook":
				r.postWebhook(nctx, e.Settings, body)
			default:
				r.obs.Logger.Debug("unknown notification provider", zap.String("provider", e.Provider))
			}
		}()
	}
}

func (r *Runtime) sendTelegram(ctx context.Context, settings map[string]any, channel, user, msg string) {
	tok, _ := settings["bot_token"].(string)

	chatID, _ := settings["chat_id"].(string)
	if tok == "" || chatID == "" {
		r.obs.Logger.Debug("telegram notification missing bot_token or chat_id")
		return
	}

	text := fmt.Sprintf("[%s] %s: %s", channel, user, truncateString(msg, 3500))
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
