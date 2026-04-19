package handler

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/observability"
	twitchoauth "github.com/rofleksey/dredge/internal/service/twitch"
	"github.com/rofleksey/dredge/internal/usecase/settings"
)

// TwitchOAuthCallbackPath is registered on the root mux (not the SPA router) for Twitch's redirect.
const TwitchOAuthCallbackPath = "/oauth/twitch/callback"

// NewTwitchOAuthCallback handles GET /oauth/twitch/callback after Twitch redirects the browser.
func NewTwitchOAuthCallback(oauth *twitchoauth.OAuth, sett *settings.Usecase, obs *observability.Stack) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if oauth == nil {
			http.Error(w, "twitch oauth is not configured", http.StatusServiceUnavailable)
			return
		}

		ctx := r.Context()
		q := r.URL.Query()

		if errParam := q.Get("error"); errParam != "" {
			desc := q.Get("error_description")
			if desc == "" {
				desc = errParam
			}

			http.Redirect(w, r, withTwitchOAuthQuery(oauth.ReturnURL(), "error", desc), http.StatusFound)
			return
		}

		code := q.Get("code")
		state := q.Get("state")

		if code == "" || state == "" {
			http.Redirect(w, r, withTwitchOAuthQuery(oauth.ReturnURL(), "error", "missing code or state"), http.StatusFound)
			return
		}

		spaReturn, err := oauth.VerifyState(state)
		if err != nil {
			obs.Logger.Debug("twitch oauth state rejected", zap.Error(err))
			http.Redirect(w, r, withTwitchOAuthQuery(oauth.ReturnURL(), "error", "invalid or expired session; try again"), http.StatusFound)
			return
		}

		base := twitchOAuthRedirectBase(oauth, spaReturn)

		tok, err := oauth.ExchangeCode(ctx, code)
		if err != nil {
			obs.Logger.Warn("twitch oauth token exchange failed", zap.Error(err))
			http.Redirect(w, r, withTwitchOAuthQuery(base, "error", "could not complete Twitch sign-in"), http.StatusFound)
			return
		}

		twitchUserID, login, err := oauth.FetchUserIdentity(ctx, tok.AccessToken)
		if err != nil {
			obs.Logger.Warn("twitch oauth helix user failed", zap.Error(err))
			http.Redirect(w, r, withTwitchOAuthQuery(base, "error", "could not read Twitch profile"), http.StatusFound)
			return
		}

		_, err = sett.CreateTwitchAccount(ctx, twitchUserID, login, tok.RefreshToken, "main")
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				http.Redirect(w, r, withTwitchOAuthQuery(base, "error", "that Twitch account is already linked"), http.StatusFound)
				return
			}

			obs.Logger.Error("twitch oauth persist account failed", zap.Error(err), zap.String("login", login))
			http.Redirect(w, r, withTwitchOAuthQuery(base, "error", "could not save Twitch account"), http.StatusFound)
			return
		}

		http.Redirect(w, r, withTwitchOAuthQuery(base, "ok", "1"), http.StatusFound)
	})
}

func twitchOAuthRedirectBase(oauth *twitchoauth.OAuth, spaReturn string) string {
	if strings.TrimSpace(spaReturn) != "" {
		return spaReturn
	}

	return oauth.ReturnURL()
}

// withTwitchOAuthQuery appends twitch_oauth_* query params. For hash-based SPAs, params are placed in the URL fragment
// so the router sees them (e.g. http://localhost/#/settings?twitch_oauth_ok=1).
func withTwitchOAuthQuery(base, key, val string) string {
	u, err := url.Parse(base)
	if err != nil {
		return base
	}

	param := "twitch_oauth_" + key

	if u.Fragment != "" {
		frag := u.Fragment
		fragPath, fragQuery, _ := strings.Cut(frag, "?")

		q, err := url.ParseQuery(fragQuery)
		if err != nil {
			q = url.Values{}
		}

		q.Set(param, val)
		u.Fragment = fragPath + "?" + q.Encode()

		return u.String()
	}

	q := u.Query()
	q.Set(param, val)
	u.RawQuery = q.Encode()

	return u.String()
}
