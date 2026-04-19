package twitch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// ExchangeCode trades an authorization code for tokens (POST https://id.twitch.tv/oauth2/token).
func (o *OAuth) ExchangeCode(ctx context.Context, code string) (userTokenResponse, error) {
	var out userTokenResponse

	form := url.Values{}
	form.Set("client_id", o.clientID)
	form.Set("client_secret", o.clientSecret)
	form.Set("code", code)
	form.Set("grant_type", "authorization_code")
	form.Set("redirect_uri", o.redirectURI)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://id.twitch.tv/oauth2/token", strings.NewReader(form.Encode()))
	if err != nil {
		return out, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return out, err
	}

	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return out, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return out, fmt.Errorf("twitch token exchange: status %d: %s", resp.StatusCode, string(body))
	}

	if err := json.Unmarshal(body, &out); err != nil {
		return out, err
	}

	if out.AccessToken == "" {
		return out, errors.New("twitch token exchange: empty access_token")
	}

	if out.RefreshToken == "" {
		return out, errors.New("twitch token exchange: empty refresh_token")
	}

	return out, nil
}
