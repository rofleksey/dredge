// Package gql implements Twitch's private GraphQL API (gql.twitch.tv) for data not exposed on Helix.
package gql

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const gqlURL = "https://gql.twitch.tv/gql"

// gqlWebClientID is the Client-ID used by twitch.tv in the browser. Public GQL operations such as
// user follows are validated against this id without a user token (same approach as common browser tooling).
const gqlWebClientID = "kd1unb4b3q4t58fwlpcbzcbnm76a8fp"

// FollowedChannel is one outgoing follow edge from a Twitch user.
type FollowedChannel struct {
	ChannelID    int64
	ChannelLogin string
	FollowedAt   *time.Time
}

// Client calls Twitch's public GraphQL API (gql.twitch.tv).
type Client struct {
	HTTPClient *http.Client
}

// NewClient returns a GQL client with a 60s HTTP timeout if HTTPClient is nil.
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 60 * time.Second}
	}
	return &Client{HTTPClient: httpClient}
}

const gqlQueryFollows = `query GetUserFollowing($login: String!, $limit: Int!, $cursor: Cursor) {
  user(login: $login) {
    id
    login
    follows(first: $limit, after: $cursor, order: DESC) {
      totalCount
      pageInfo { hasNextPage endCursor }
      edges { followedAt node { id login displayName } }
    }
  }
}`

// FetchUserFollows returns all channels the user follows (outgoing), paginating until done or maxPages.
// It uses the same anonymous Client-ID as twitch.tv; do not send a user OAuth token (Helix Bearer tokens are rejected as OAuth).
func (c *Client) FetchUserFollows(ctx context.Context, login string, limit int, maxPages int) ([]FollowedChannel, int, error) {
	if limit < 1 || limit > 100 {
		limit = 100
	}

	if maxPages < 1 {
		maxPages = 1
	}

	login = strings.ToLower(strings.TrimSpace(login))
	if login == "" {
		return nil, 0, errors.New("empty login")
	}

	var (
		all       []FollowedChannel
		cursor    *string
		totalSeen int
	)

	for page := 0; page < maxPages; page++ {
		payload := map[string]any{
			"operationName": "GetUserFollowing",
			"variables": map[string]any{
				"login":  login,
				"limit":  limit,
				"cursor": cursor,
			},
			"query": gqlQueryFollows,
		}

		body, err := json.Marshal(payload)
		if err != nil {
			return nil, 0, err
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, gqlURL, bytes.NewReader(body))
		if err != nil {
			return nil, 0, err
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Client-ID", gqlWebClientID)

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			return nil, 0, err
		}

		rb, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		if err != nil {
			return nil, 0, err
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, 0, fmt.Errorf("gql: status %d: %s", resp.StatusCode, string(rb))
		}

		var parsed gqlFollowResponse
		if err := json.Unmarshal(rb, &parsed); err != nil {
			return nil, 0, fmt.Errorf("gql decode: %w", err)
		}

		if parsed.Data.User == nil {
			return all, totalSeen, nil
		}

		f := parsed.Data.User.Follows
		if page == 0 {
			totalSeen = f.TotalCount
		}

		for _, e := range f.Edges {
			chID, err := parseTwitchUserID(e.Node.ID)
			if err != nil {
				continue
			}

			loginLower := strings.ToLower(strings.TrimSpace(e.Node.Login))
			if loginLower == "" {
				continue
			}

			var fa *time.Time

			if e.FollowedAt != "" {
				if t, err := time.Parse(time.RFC3339, e.FollowedAt); err == nil {
					tu := t.UTC()
					fa = &tu
				}
			}

			all = append(all, FollowedChannel{
				ChannelID:    chID,
				ChannelLogin: loginLower,
				FollowedAt:   fa,
			})
		}

		if !f.PageInfo.HasNextPage || f.PageInfo.EndCursor == "" {
			break
		}

		cursor = &f.PageInfo.EndCursor
	}

	return all, totalSeen, nil
}

type gqlFollowResponse struct {
	Data struct {
		User *struct {
			Follows struct {
				TotalCount int `json:"totalCount"`
				PageInfo   struct {
					HasNextPage bool   `json:"hasNextPage"`
					EndCursor   string `json:"endCursor"`
				} `json:"pageInfo"`
				Edges []struct {
					FollowedAt string `json:"followedAt"`
					Node       struct {
						ID    string `json:"id"`
						Login string `json:"login"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"follows"`
		} `json:"user"`
	} `json:"data"`
}

func parseTwitchUserID(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errors.New("empty id")
	}

	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		return n, nil
	}

	var b strings.Builder

	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}

	if b.Len() == 0 {
		return 0, fmt.Errorf("unparseable twitch user id %q", s)
	}

	return strconv.ParseInt(b.String(), 10, 64)
}
