package twitch

// ReturnURL is the configured SPA URL for post-OAuth redirect.
func (o *OAuth) ReturnURL() string {
	return o.returnURL
}
