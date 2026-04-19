package twitch

type oauthStatePayload struct {
	Exp int64  `json:"exp"`
	N   string `json:"n"`
	// Ret is an optional post-OAuth SPA URL (same origin as configured return URL), carried only in signed state.
	Ret string `json:"ret,omitempty"`
}

type userTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}
