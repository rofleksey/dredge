package helix

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
)

// SendChatMessage POSTs to Helix /helix/chat/messages (requires user:write:chat on the access token).
func (c *Client) SendChatMessage(ctx context.Context, userAccessToken string, broadcasterID, senderID int64, message string) error {
	ctx, span := c.Obs.StartSpan(ctx, "service.twitch.helix_send_chat_message")
	defer span.End()

	payload := struct {
		BroadcasterID string `json:"broadcaster_id"`
		SenderID      string `json:"sender_id"`
		Message       string `json:"message"`
	}{
		BroadcasterID: strconv.FormatInt(broadcasterID, 10),
		SenderID:      strconv.FormatInt(senderID, 10),
		Message:       message,
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.twitch.tv/helix/chat/messages", bytes.NewReader(raw))
	if err != nil {
		return err
	}

	req.Header.Set("Client-Id", c.ClientID)
	req.Header.Set("Authorization", "Bearer "+userAccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		c.Obs.LogError(ctx, span, "helix send chat request failed", err)
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err := helixAPIFailure(resp.StatusCode, body)
		c.Obs.LogError(ctx, span, "helix send chat rejected", err)
		return err
	}

	var out struct {
		Data []struct {
			IsSent     bool   `json:"is_sent"`
			DropReason string `json:"drop_reason"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		c.Obs.LogError(ctx, span, "decode helix send chat response failed", err)
		return err
	}

	if len(out.Data) == 0 {
		return &ChatSendError{Message: "Twitch did not accept the message (empty response)."}
	}

	if !out.Data[0].IsSent {
		msg := "Twitch did not deliver the message."
		if out.Data[0].DropReason != "" {
			msg = out.Data[0].DropReason
		}
		return &ChatSendError{Message: msg}
	}

	return nil
}
