package helix

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var twitchLoginRe = regexp.MustCompile(`^[a-zA-Z0-9_]{4,25}$`)

func normalizeChannelInput(raw string) string {
	s := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(raw), "#"))
	return strings.ToLower(s)
}

// ResolvedChannel is the Twitch user id and canonical login from Helix.
type ResolvedChannel struct {
	ID       int64
	Username string
}

// ResolveChannel checks Helix for the login and returns the broadcaster user id and canonical login.
func (c *Client) ResolveChannel(ctx context.Context, raw string) (ResolvedChannel, error) {
	ctx, span := c.Obs.StartSpan(ctx, "service.twitch.resolve_channel")
	defer span.End()

	login := normalizeChannelInput(raw)
	if login == "" || !twitchLoginRe.MatchString(login) {
		return ResolvedChannel{}, ErrInvalidChannelName
	}

	token, err := c.appAccessToken(ctx)
	if err != nil {
		c.Obs.LogError(ctx, span, "app access token failed", err)
		return ResolvedChannel{}, err
	}

	u, err := url.Parse("https://api.twitch.tv/helix/users")
	if err != nil {
		return ResolvedChannel{}, err
	}

	q := u.Query()
	q.Set("login", login)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return ResolvedChannel{}, err
	}

	req.Header.Set("Client-Id", c.ClientID)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		c.Obs.LogError(ctx, span, "helix users request failed", err, zap.String("login", login))
		return ResolvedChannel{}, err
	}

	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ResolvedChannel{}, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = fmt.Errorf("helix users: status %d: %s", resp.StatusCode, string(body))
		c.Obs.LogError(ctx, span, "helix users rejected", err, zap.String("login", login))
		return ResolvedChannel{}, err
	}

	var out struct {
		Data []struct {
			ID    string `json:"id"`
			Login string `json:"login"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		c.Obs.LogError(ctx, span, "decode helix users failed", err)
		return ResolvedChannel{}, err
	}

	if len(out.Data) == 0 {
		return ResolvedChannel{}, ErrUnknownTwitchChannel
	}

	row := out.Data[0]

	uid, err := strconv.ParseInt(row.ID, 10, 64)
	if err != nil {
		c.Obs.LogError(ctx, span, "parse twitch user id failed", err, zap.String("id", row.ID))
		return ResolvedChannel{}, err
	}

	return ResolvedChannel{ID: uid, Username: row.Login}, nil
}

func (c *Client) appAccessToken(ctx context.Context) (string, error) {
	c.appTokenMu.Lock()
	defer c.appTokenMu.Unlock()

	if c.appToken != "" && time.Now().Before(c.appTokenExp.Add(-2*time.Minute)) {
		return c.appToken, nil
	}

	form := url.Values{}
	form.Set("client_id", c.ClientID)
	form.Set("client_secret", c.ClientSecret)
	form.Set("grant_type", "client_credentials")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://id.twitch.tv/oauth2/token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}

	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("client credentials: status %d: %s", resp.StatusCode, string(body))
	}

	var tok struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &tok); err != nil {
		return "", err
	}

	if tok.AccessToken == "" {
		return "", fmt.Errorf("empty app access token")
	}

	exp := time.Duration(tok.ExpiresIn) * time.Second
	if exp <= 0 {
		exp = time.Hour
	}

	c.appToken = tok.AccessToken
	c.appTokenExp = time.Now().Add(exp)

	return c.appToken, nil
}

const helixUserBatch = 100

// HelixUserRecord is a subset of Get Users fields used for enrichment.
type HelixUserRecord struct {
	ID              int64
	Login           string
	CreatedAt       *time.Time
	ProfileImageURL string
}

// HelixUsersByIDs fetches users by id (app token). Up to 100 ids per request.
func (c *Client) HelixUsersByIDs(ctx context.Context, ids []int64) ([]HelixUserRecord, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	ctx, span := c.Obs.StartSpan(ctx, "service.twitch.helix_users_by_ids")
	defer span.End()

	token, err := c.appAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var out []HelixUserRecord

	for i := 0; i < len(ids); i += helixUserBatch {
		end := i + helixUserBatch
		if end > len(ids) {
			end = len(ids)
		}

		chunk := ids[i:end]

		u, err := url.Parse("https://api.twitch.tv/helix/users")
		if err != nil {
			return nil, err
		}

		q := u.Query()
		for _, id := range chunk {
			q.Add("id", strconv.FormatInt(id, 10))
		}

		u.RawQuery = q.Encode()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Client-Id", c.ClientID)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			c.Obs.LogError(ctx, span, "helix users by id batch failed", err)
			return nil, err
		}

		body, err := readHelixBody(resp)
		_ = resp.Body.Close()

		if err != nil {
			return nil, err
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			err = fmt.Errorf("helix users: status %d: %s", resp.StatusCode, string(body))
			c.Obs.LogError(ctx, span, "helix users rejected", err)
			return nil, err
		}

		var parsed struct {
			Data []struct {
				ID              string `json:"id"`
				Login           string `json:"login"`
				CreatedAt       string `json:"created_at"`
				ProfileImageURL string `json:"profile_image_url"`
			} `json:"data"`
		}
		if err := json.Unmarshal(body, &parsed); err != nil {
			return nil, err
		}

		for _, row := range parsed.Data {
			uid, err := strconv.ParseInt(row.ID, 10, 64)
			if err != nil {
				continue
			}

			rec := HelixUserRecord{ID: uid, Login: row.Login, ProfileImageURL: row.ProfileImageURL}
			if row.CreatedAt != "" {
				if t, err := time.Parse(time.RFC3339, row.CreatedAt); err == nil {
					rec.CreatedAt = &t
				}
			}

			out = append(out, rec)
		}
	}

	return out, nil
}

// HelixUsersByLogins resolves logins to ids (app token). Up to 100 logins per request.
func (c *Client) HelixUsersByLogins(ctx context.Context, logins []string) (map[string]int64, error) {
	if len(logins) == 0 {
		return map[string]int64{}, nil
	}

	ctx, span := c.Obs.StartSpan(ctx, "service.twitch.helix_users_by_logins")
	defer span.End()

	token, err := c.appAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	out := make(map[string]int64, len(logins))

	for i := 0; i < len(logins); i += helixUserBatch {
		end := i + helixUserBatch
		if end > len(logins) {
			end = len(logins)
		}

		chunk := logins[i:end]

		u, err := url.Parse("https://api.twitch.tv/helix/users")
		if err != nil {
			return nil, err
		}

		q := u.Query()

		for _, login := range chunk {
			ln := normalizeChannelInput(login)
			if ln == "" {
				continue
			}

			q.Add("login", ln)
		}

		u.RawQuery = q.Encode()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Client-Id", c.ClientID)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			c.Obs.LogError(ctx, span, "helix users by login batch failed", err)
			return nil, err
		}

		body, err := readHelixBody(resp)
		_ = resp.Body.Close()

		if err != nil {
			return nil, err
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			err = fmt.Errorf("helix users: status %d: %s", resp.StatusCode, string(body))
			c.Obs.LogError(ctx, span, "helix users rejected", err)
			return nil, err
		}

		var parsed struct {
			Data []struct {
				ID    string `json:"id"`
				Login string `json:"login"`
			} `json:"data"`
		}
		if err := json.Unmarshal(body, &parsed); err != nil {
			return nil, err
		}

		for _, row := range parsed.Data {
			uid, err := strconv.ParseInt(row.ID, 10, 64)
			if err != nil {
				continue
			}

			out[strings.ToLower(row.Login)] = uid
		}
	}

	return out, nil
}

func readHelixBody(resp *http.Response) ([]byte, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// ChannelFollowerPage is one follower from Helix channels/followers.
type ChannelFollowerPage struct {
	UserID     int64
	UserLogin  string
	FollowedAt *time.Time
	Cursor     string
}

// FetchChannelFollowersPage returns one page of channel followers (user access token).
func (c *Client) FetchChannelFollowersPage(ctx context.Context, accessToken string, broadcasterID, moderatorID int64, after string) ([]ChannelFollowerPage, string, error) {
	ctx, span := c.Obs.StartSpan(ctx, "service.twitch.channel_followers_page")
	defer span.End()

	u, err := url.Parse("https://api.twitch.tv/helix/channels/followers")
	if err != nil {
		return nil, "", err
	}

	q := u.Query()
	q.Set("broadcaster_id", strconv.FormatInt(broadcasterID, 10))
	q.Set("moderator_id", strconv.FormatInt(moderatorID, 10))
	q.Set("first", "100")

	if after != "" {
		q.Set("after", after)
	}

	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, "", err
	}

	req.Header.Set("Client-Id", c.ClientID)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		c.Obs.LogError(ctx, span, "helix followers request failed", err)
		return nil, "", err
	}

	body, err := readHelixBody(resp)
	_ = resp.Body.Close()

	if err != nil {
		return nil, "", err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = fmt.Errorf("helix followers: status %d: %s", resp.StatusCode, string(body))
		c.Obs.LogError(ctx, span, "helix followers rejected", err)
		return nil, "", err
	}

	var parsed struct {
		Data []struct {
			UserID    string `json:"user_id"`
			UserLogin string `json:"user_login"`
			Followed  string `json:"followed_at"`
		} `json:"data"`
		Pagination struct {
			Cursor string `json:"cursor"`
		} `json:"pagination"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, "", err
	}

	out := make([]ChannelFollowerPage, 0, len(parsed.Data))

	for _, row := range parsed.Data {
		uid, err := strconv.ParseInt(row.UserID, 10, 64)
		if err != nil {
			continue
		}

		var fa *time.Time

		if row.Followed != "" {
			if t, err := time.Parse(time.RFC3339, row.Followed); err == nil {
				fa = &t
			}
		}

		out = append(out, ChannelFollowerPage{UserID: uid, UserLogin: row.UserLogin, FollowedAt: fa})
	}

	return out, parsed.Pagination.Cursor, nil
}

// ResolveUserIDByLogin uses Helix Get Users (app token) for a single login.
func (c *Client) ResolveUserIDByLogin(ctx context.Context, login string) (int64, error) {
	m, err := c.HelixUsersByLogins(ctx, []string{login})
	if err != nil {
		return 0, err
	}

	ln := normalizeChannelInput(login)

	id, ok := m[ln]
	if !ok {
		return 0, ErrUnknownTwitchChannel
	}

	return id, nil
}

// ChannelLiveInfo is Helix user + optional stream snapshot.
type ChannelLiveInfo struct {
	BroadcasterID    int64
	BroadcasterLogin string
	DisplayName      string
	ProfileImageURL  string
	IsLive           bool
	Title            string
	GameName         string
	ViewerCount      int64
	StreamStartedAt  *time.Time
}

// GetChannelLive fetches user + stream for a channel login (app token).
func (c *Client) GetChannelLive(ctx context.Context, login string) (ChannelLiveInfo, error) {
	ctx, span := c.Obs.StartSpan(ctx, "service.twitch.get_channel_live")
	defer span.End()

	ln := normalizeChannelInput(login)
	if ln == "" || !twitchLoginRe.MatchString(ln) {
		return ChannelLiveInfo{}, ErrInvalidChannelName
	}

	token, err := c.appAccessToken(ctx)
	if err != nil {
		return ChannelLiveInfo{}, err
	}

	u, err := url.Parse("https://api.twitch.tv/helix/users")
	if err != nil {
		return ChannelLiveInfo{}, err
	}

	q := u.Query()
	q.Set("login", ln)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return ChannelLiveInfo{}, err
	}

	req.Header.Set("Client-Id", c.ClientID)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return ChannelLiveInfo{}, err
	}

	body, err := readHelixBody(resp)
	_ = resp.Body.Close()

	if err != nil {
		return ChannelLiveInfo{}, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ChannelLiveInfo{}, ErrUnknownTwitchChannel
	}

	var usersOut struct {
		Data []struct {
			ID              string `json:"id"`
			Login           string `json:"login"`
			DisplayName     string `json:"display_name"`
			ProfileImageURL string `json:"profile_image_url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &usersOut); err != nil || len(usersOut.Data) == 0 {
		return ChannelLiveInfo{}, ErrUnknownTwitchChannel
	}

	urow := usersOut.Data[0]

	bid, err := strconv.ParseInt(urow.ID, 10, 64)
	if err != nil {
		return ChannelLiveInfo{}, err
	}

	info := ChannelLiveInfo{
		BroadcasterID:    bid,
		BroadcasterLogin: urow.Login,
		DisplayName:      urow.DisplayName,
		ProfileImageURL:  urow.ProfileImageURL,
	}

	su, err := url.Parse("https://api.twitch.tv/helix/streams")
	if err != nil {
		return info, nil
	}

	sq := su.Query()
	sq.Set("user_login", ln)
	su.RawQuery = sq.Encode()

	sreq, err := http.NewRequestWithContext(ctx, http.MethodGet, su.String(), nil)
	if err != nil {
		return info, nil
	}

	sreq.Header.Set("Client-Id", c.ClientID)
	sreq.Header.Set("Authorization", "Bearer "+token)

	sresp, err := c.HTTPClient.Do(sreq)
	if err != nil {
		return info, nil
	}

	sbody, err := readHelixBody(sresp)

	_ = sresp.Body.Close()
	if err != nil || sresp.StatusCode < 200 || sresp.StatusCode >= 300 {
		return info, nil
	}

	var streamOut struct {
		Data []struct {
			ViewerCount int    `json:"viewer_count"`
			Title       string `json:"title"`
			GameName    string `json:"game_name"`
			StartedAt   string `json:"started_at"`
		} `json:"data"`
	}
	if err := json.Unmarshal(sbody, &streamOut); err != nil || len(streamOut.Data) == 0 {
		return info, nil
	}

	st := streamOut.Data[0]
	info.IsLive = true
	info.ViewerCount = int64(st.ViewerCount)
	info.Title = st.Title

	info.GameName = st.GameName
	if st.StartedAt != "" {
		if t, err := time.Parse(time.RFC3339, st.StartedAt); err == nil {
			info.StreamStartedAt = &t
		}
	}

	return info, nil
}

// HelixStreamsLiveByBroadcasterIDs returns true for broadcaster user ids that have an active stream.
func (c *Client) HelixStreamsLiveByBroadcasterIDs(ctx context.Context, ids []int64) (map[int64]bool, error) {
	out := make(map[int64]bool, len(ids))
	if len(ids) == 0 {
		return out, nil
	}

	ctx, span := c.Obs.StartSpan(ctx, "service.twitch.helix_streams_by_user_ids")
	defer span.End()

	token, err := c.appAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	const batch = 100

	for i := 0; i < len(ids); i += batch {
		end := i + batch
		if end > len(ids) {
			end = len(ids)
		}

		chunk := ids[i:end]

		u, err := url.Parse("https://api.twitch.tv/helix/streams")
		if err != nil {
			return nil, err
		}

		q := u.Query()
		for _, id := range chunk {
			q.Add("user_id", strconv.FormatInt(id, 10))
		}

		u.RawQuery = q.Encode()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Client-Id", c.ClientID)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			return nil, err
		}

		body, err := readHelixBody(resp)
		_ = resp.Body.Close()

		if err != nil {
			return nil, err
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, fmt.Errorf("helix streams: status %d: %s", resp.StatusCode, string(body))
		}

		var parsed struct {
			Data []struct {
				UserID string `json:"user_id"`
			} `json:"data"`
		}
		if err := json.Unmarshal(body, &parsed); err != nil {
			return nil, err
		}

		for _, row := range parsed.Data {
			uid, err := strconv.ParseInt(row.UserID, 10, 64)
			if err != nil {
				continue
			}

			out[uid] = true
		}
	}

	return out, nil
}

// HelixStreamSnapshot is one live Helix /streams row for a broadcaster.
type HelixStreamSnapshot struct {
	UserID        int64
	HelixStreamID string
	StartedAt     time.Time
	Title         string
	GameName      string
}

// HelixStreamsMetadataByBroadcasterIDs returns live stream metadata keyed by broadcaster user id.
func (c *Client) HelixStreamsMetadataByBroadcasterIDs(ctx context.Context, ids []int64) (map[int64]HelixStreamSnapshot, error) {
	out := make(map[int64]HelixStreamSnapshot)
	if len(ids) == 0 {
		return out, nil
	}

	ctx, span := c.Obs.StartSpan(ctx, "service.twitch.helix_streams_metadata_by_user_ids")
	defer span.End()

	token, err := c.appAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	const batch = 100

	for i := 0; i < len(ids); i += batch {
		end := i + batch
		if end > len(ids) {
			end = len(ids)
		}

		chunk := ids[i:end]

		u, err := url.Parse("https://api.twitch.tv/helix/streams")
		if err != nil {
			return nil, err
		}

		q := u.Query()
		for _, id := range chunk {
			q.Add("user_id", strconv.FormatInt(id, 10))
		}

		u.RawQuery = q.Encode()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Client-Id", c.ClientID)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			return nil, err
		}

		body, err := readHelixBody(resp)
		_ = resp.Body.Close()

		if err != nil {
			return nil, err
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, fmt.Errorf("helix streams: status %d: %s", resp.StatusCode, string(body))
		}

		var parsed struct {
			Data []struct {
				ID        string `json:"id"`
				UserID    string `json:"user_id"`
				StartedAt string `json:"started_at"`
				Title     string `json:"title"`
				GameName  string `json:"game_name"`
			} `json:"data"`
		}
		if err := json.Unmarshal(body, &parsed); err != nil {
			return nil, err
		}

		for _, row := range parsed.Data {
			uid, err := strconv.ParseInt(row.UserID, 10, 64)
			if err != nil {
				continue
			}

			var st time.Time

			if row.StartedAt != "" {
				if t, err := time.Parse(time.RFC3339, row.StartedAt); err == nil {
					st = t.UTC()
				}
			}

			out[uid] = HelixStreamSnapshot{
				UserID:        uid,
				HelixStreamID: row.ID,
				StartedAt:     st,
				Title:         row.Title,
				GameName:      row.GameName,
			}
		}
	}

	return out, nil
}

// maxUserOAuthCacheTTL is how long we reuse a user (refresh-grant) access token before refreshing again.
const maxUserOAuthCacheTTL = 30 * time.Minute

// RefreshAccessToken exchanges a Twitch refresh token for new OAuth credentials.
func (c *Client) RefreshAccessToken(ctx context.Context, refreshToken string) (string, string, error) {
	ctx, span := c.Obs.StartSpan(ctx, "service.twitch.refresh_access_token")
	defer span.End()

	at, newRT, _, err := c.refreshUserAccessToken(ctx, span, refreshToken)
	return at, newRT, err
}

func (c *Client) refreshUserAccessToken(ctx context.Context, span trace.Span, refreshToken string) (accessToken string, newRefreshToken string, expiresInSec int, err error) {
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)
	form.Set("client_id", c.ClientID)
	form.Set("client_secret", c.ClientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://id.twitch.tv/oauth2/token", nil)
	if err != nil {
		c.Obs.LogError(ctx, span, "build oauth request failed", err)
		return "", "", 0, err
	}

	req.URL.RawQuery = form.Encode()

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		c.Obs.LogError(ctx, span, "oauth request failed", err)
		return "", "", 0, err
	}

	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.Obs.LogError(ctx, span, "read oauth response failed", err)
		return "", "", 0, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = fmt.Errorf("oauth token refresh: status %d: %s", resp.StatusCode, string(body))
		c.Obs.LogError(ctx, span, "oauth request rejected", err)
		return "", "", 0, err
	}

	var out refreshResp

	if err := json.Unmarshal(body, &out); err != nil {
		c.Obs.LogError(ctx, span, "decode oauth response failed", err)
		return "", "", 0, err
	}

	if out.AccessToken == "" {
		err = errors.New("empty access token")
		c.Obs.LogError(ctx, span, "oauth response missing access token", err)
		return "", "", 0, err
	}

	return out.AccessToken, out.RefreshToken, out.ExpiresIn, nil
}

// CachedUserAccessTokenForAccount returns a user access token for Helix user endpoints, reusing a recent
// refresh for up to maxUserOAuthCacheTTL (and never past the token's expires_in from Twitch).
func (c *Client) CachedUserAccessTokenForAccount(ctx context.Context, accountID int64, refreshToken string) (accessToken string, newRefreshToken string, err error) {
	now := time.Now()

	c.userOAuthMu.Lock()
	if c.userOAuth != nil {
		if e, ok := c.userOAuth[accountID]; ok && e.refreshSnapshot == refreshToken && now.Before(e.expiresAt) {
			at := e.accessToken
			c.userOAuthMu.Unlock()
			return at, "", nil
		}
	}
	c.userOAuthMu.Unlock()

	ctx, span := c.Obs.StartSpan(ctx, "service.twitch.refresh_access_token")
	defer span.End()

	at, newRT, expSec, err := c.refreshUserAccessToken(ctx, span, refreshToken)
	if err != nil {
		return "", "", err
	}

	ttl := maxUserOAuthCacheTTL

	if expSec > 0 {
		capTTL := time.Duration(expSec)*time.Second - 90*time.Second
		if capTTL < ttl {
			ttl = capTTL
		}
	}

	if ttl < 0 {
		ttl = 0
	}

	snapshot := refreshToken
	if newRT != "" {
		snapshot = newRT
	}

	if ttl > 0 {
		expiresAt := time.Now().Add(ttl)

		c.userOAuthMu.Lock()

		if c.userOAuth == nil {
			c.userOAuth = make(map[int64]userOAuthCacheEntry)
		}

		c.userOAuth[accountID] = userOAuthCacheEntry{
			accessToken:     at,
			refreshSnapshot: snapshot,
			expiresAt:       expiresAt,
		}
		c.userOAuthMu.Unlock()
	}

	return at, newRT, nil
}
