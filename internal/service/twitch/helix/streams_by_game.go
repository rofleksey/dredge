package helix

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// HelixGameStreamRow is one live stream from GET /helix/streams filtered by game_id.
type HelixGameStreamRow struct {
	UserID      int64
	UserLogin   string
	ViewerCount int64
	Title       string
	GameName    string
	Tags        []string
}

// HelixStreamsByGameIDPage fetches one page of live streams for a Twitch category (game_id).
// first must be in [1, 100]. Pass after="" for the first page; use the returned nextCursor for the next page (empty when done).
func (c *Client) HelixStreamsByGameIDPage(ctx context.Context, gameID string, first int, after string) ([]HelixGameStreamRow, string, error) {
	ctx, span := c.Obs.StartSpan(ctx, "service.twitch.helix_streams_by_game_id")
	defer span.End()

	gameID = strings.TrimSpace(gameID)
	if gameID == "" {
		return nil, "", ErrInvalidGameID
	}

	if first < 1 {
		first = 1
	}

	if first > 100 {
		first = 100
	}

	token, err := c.appAccessToken(ctx)
	if err != nil {
		c.Obs.LogError(ctx, span, "helix streams by game app token failed", err)
		return nil, "", err
	}

	u, err := url.Parse("https://api.twitch.tv/helix/streams")
	if err != nil {
		return nil, "", err
	}

	q := u.Query()
	q.Set("game_id", gameID)
	q.Set("first", strconv.Itoa(first))

	if strings.TrimSpace(after) != "" {
		q.Set("after", after)
	}

	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, "", err
	}

	req.Header.Set("Client-Id", c.ClientID)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		c.Obs.LogError(ctx, span, "helix streams by game request failed", err)
		return nil, "", err
	}

	body, err := readHelixBody(resp)
	_ = resp.Body.Close()

	if err != nil {
		return nil, "", err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = fmt.Errorf("helix streams by game: status %d: %s", resp.StatusCode, string(body))
		c.Obs.LogError(ctx, span, "helix streams by game rejected", err)
		return nil, "", err
	}

	var parsed struct {
		Data []struct {
			UserID      string   `json:"user_id"`
			UserLogin   string   `json:"user_login"`
			Title       string   `json:"title"`
			GameName    string   `json:"game_name"`
			ViewerCount int      `json:"viewer_count"`
			Tags        []string `json:"tags"`
		} `json:"data"`
		Pagination struct {
			Cursor string `json:"cursor"`
		} `json:"pagination"`
	}

	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, "", fmt.Errorf("helix streams by game: decode: %w", err)
	}

	out := make([]HelixGameStreamRow, 0, len(parsed.Data))

	for _, row := range parsed.Data {
		uid, err := strconv.ParseInt(row.UserID, 10, 64)
		if err != nil {
			continue
		}

		tags := row.Tags
		if tags == nil {
			tags = []string{}
		}

		out = append(out, HelixGameStreamRow{
			UserID:      uid,
			UserLogin:   strings.ToLower(strings.TrimSpace(row.UserLogin)),
			ViewerCount: int64(row.ViewerCount),
			Title:       row.Title,
			GameName:    row.GameName,
			Tags:        tags,
		})
	}

	return out, parsed.Pagination.Cursor, nil
}
