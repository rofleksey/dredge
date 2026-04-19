// Package gql implements Twitch's private GraphQL API (gql.twitch.tv) for data not exposed on Helix.
package gql

import (
	"net/http"
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
