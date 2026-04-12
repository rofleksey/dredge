package twitch

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// OAuth performs the Twitch authorization-code flow (same token endpoints as twitchium).
type OAuth struct {
	clientID     string
	clientSecret string
	redirectURI  string
	returnURL    string
	scopes       []string
	hmacSecret   string
	httpClient   *http.Client

	stateMu sync.Mutex
	// Intentional: single-process nonce store; multiple app instances do not share replay state.
	usedStateNonces map[string]int64 // nonce -> state expiry (unix); single-use within TTL
}

// NewOAuth builds the OAuth helper; twitch.oauth_redirect_uri and twitch.oauth_return_url must be set in config (validated at load).
func NewOAuth(clientID, clientSecret, redirectURI, returnURL, stateHMACSecret string) *OAuth {
	return &OAuth{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
		returnURL:    returnURL,
		scopes:       UserOAuthScopes,
		hmacSecret:   stateHMACSecret,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}
}

type oauthStatePayload struct {
	Exp int64  `json:"exp"`
	N   string `json:"n"`
	// Ret is an optional post-OAuth SPA URL (same origin as configured return URL), carried only in signed state.
	Ret string `json:"ret,omitempty"`
}

// ValidateSPAReturnURL checks that candidate is a safe redirect target (same origin as configured oauth_return_url).
func (o *OAuth) ValidateSPAReturnURL(candidate string) error {
	candidate = strings.TrimSpace(candidate)
	if candidate == "" {
		return nil
	}

	if len(candidate) > 2048 {
		return errors.New("return url too long")
	}

	base, err := url.Parse(o.returnURL)
	if err != nil {
		return fmt.Errorf("oauth: invalid configured return url: %w", err)
	}

	cand, err := url.Parse(candidate)
	if err != nil {
		return fmt.Errorf("invalid return url: %w", err)
	}

	if cand.Scheme != base.Scheme || cand.Host != base.Host {
		return errors.New("return url origin mismatch")
	}

	return nil
}

// NewState builds a signed OAuth state parameter (CSRF protection). spaReturnURL is optional (empty = use configured return URL only).
func (o *OAuth) NewState(spaReturnURL string) (string, error) {
	if err := o.ValidateSPAReturnURL(spaReturnURL); err != nil {
		return "", err
	}

	var raw [8]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}

	p := oauthStatePayload{
		Exp: time.Now().Add(15 * time.Minute).Unix(),
		N:   hex.EncodeToString(raw[:]),
		Ret: strings.TrimSpace(spaReturnURL),
	}

	payload, err := json.Marshal(p)
	if err != nil {
		return "", err
	}

	mac := hmac.New(sha256.New, []byte(o.hmacSecret))
	_, _ = mac.Write(payload)
	sig := mac.Sum(nil)

	return base64.RawURLEncoding.EncodeToString(payload) + "." + base64.RawURLEncoding.EncodeToString(sig), nil
}

// VerifyState checks the HMAC and expiry on a state value from Twitch's redirect.
// It returns an optional SPA return URL embedded in state (empty = use configured oauth_return_url).
func (o *OAuth) VerifyState(state string) (spaReturnURL string, err error) {
	parts := strings.Split(state, ".")
	if len(parts) != 2 {
		return "", errors.New("invalid state")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", fmt.Errorf("invalid state payload: %w", err)
	}

	wantSig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("invalid state sig: %w", err)
	}

	mac := hmac.New(sha256.New, []byte(o.hmacSecret))
	_, _ = mac.Write(payload)
	got := mac.Sum(nil)

	if !hmac.Equal(wantSig, got) {
		return "", errors.New("state signature mismatch")
	}

	var p oauthStatePayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return "", fmt.Errorf("invalid state json: %w", err)
	}

	if time.Now().Unix() > p.Exp {
		return "", errors.New("state expired")
	}

	if p.Ret != "" {
		if err := o.ValidateSPAReturnURL(p.Ret); err != nil {
			return "", fmt.Errorf("state return url: %w", err)
		}
	}

	o.stateMu.Lock()
	defer o.stateMu.Unlock()

	if o.usedStateNonces == nil {
		o.usedStateNonces = make(map[string]int64)
	}

	now := time.Now().Unix()

	for n, exp := range o.usedStateNonces {
		if now > exp {
			delete(o.usedStateNonces, n)
		}
	}

	if _, dup := o.usedStateNonces[p.N]; dup {
		return "", errors.New("state already used")
	}

	o.usedStateNonces[p.N] = p.Exp

	return p.Ret, nil
}

// AuthorizeURL is the Twitch page where the user approves scopes (twitchium-style authorize URL).
func (o *OAuth) AuthorizeURL(state string) string {
	params := url.Values{}
	params.Set("client_id", o.clientID)
	params.Set("redirect_uri", o.redirectURI)
	params.Set("response_type", "code")
	params.Set("scope", strings.Join(o.scopes, " "))
	params.Set("state", state)

	return "https://id.twitch.tv/oauth2/authorize?" + params.Encode()
}

type userTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

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

// ReturnURL is the configured SPA URL for post-OAuth redirect.
func (o *OAuth) ReturnURL() string {
	return o.returnURL
}
