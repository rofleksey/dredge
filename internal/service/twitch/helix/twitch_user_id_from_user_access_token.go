package helix

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"go.uber.org/zap"
)

// TwitchUserIDFromUserAccessToken returns the authorized user's Twitch user id (GET /helix/users with a user access token).
func (c *Client) TwitchUserIDFromUserAccessToken(ctx context.Context, userAccessToken string) (int64, error) {
	ctx, span := c.Obs.StartSpan(ctx, "service.twitch.helix_user_from_token")
	defer span.End()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.twitch.tv/helix/users", nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("Client-Id", c.ClientID)
	req.Header.Set("Authorization", "Bearer "+userAccessToken)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		c.Obs.LogError(ctx, span, "helix users (user token) request failed", err)
		return 0, err
	}

	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = helixAPIFailure(resp.StatusCode, body)
		c.Obs.LogError(ctx, span, "helix users (user token) rejected", err)
		return 0, err
	}

	var out struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		c.Obs.LogError(ctx, span, "decode helix users (user token) failed", err)
		return 0, err
	}

	if len(out.Data) == 0 || out.Data[0].ID == "" {
		err = fmt.Errorf("helix users: empty data")
		c.Obs.LogError(ctx, span, "helix users (user token) missing user", err)
		return 0, err
	}

	uid, err := strconv.ParseInt(out.Data[0].ID, 10, 64)
	if err != nil {
		c.Obs.LogError(ctx, span, "parse twitch user id failed", err, zap.String("id", out.Data[0].ID))
		return 0, err
	}

	return uid, nil
}
