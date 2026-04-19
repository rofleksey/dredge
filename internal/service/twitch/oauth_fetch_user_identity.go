package twitch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// FetchUserIdentity returns the authorized user's Twitch user id and login (lowercase) from Helix /users.
func (o *OAuth) FetchUserIdentity(ctx context.Context, userAccessToken string) (id int64, login string, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.twitch.tv/helix/users", nil)
	if err != nil {
		return 0, "", err
	}

	req.Header.Set("Client-Id", o.clientID)
	req.Header.Set("Authorization", "Bearer "+userAccessToken)

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return 0, "", err
	}

	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, "", err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, "", fmt.Errorf("helix users (oauth): status %d: %s", resp.StatusCode, string(body))
	}

	var out struct {
		Data []struct {
			ID    string `json:"id"`
			Login string `json:"login"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &out); err != nil {
		return 0, "", err
	}

	if len(out.Data) == 0 || out.Data[0].Login == "" {
		return 0, "", errors.New("helix users: empty data")
	}

	uid, err := strconv.ParseInt(out.Data[0].ID, 10, 64)
	if err != nil || uid <= 0 {
		return 0, "", fmt.Errorf("helix users: invalid id: %w", err)
	}

	return uid, strings.ToLower(out.Data[0].Login), nil
}
