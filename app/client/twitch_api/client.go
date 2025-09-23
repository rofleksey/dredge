package twitch_api

import (
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/nicklaw5/helix/v2"
)

type Client struct {
	cfg        Config
	userClient *helix.Client

	tokenMutex  sync.Mutex
	lastUpdated time.Time
	accessToken string
}

type Config struct {
	ClientID     string
	ClientSecret string
	Username     string
	RefreshToken string
}

func NewClient(cfg Config) (*Client, error) {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	helixClient, err := helix.NewClient(&helix.Options{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RefreshToken: cfg.RefreshToken,
		HTTPClient:   httpClient,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create helix client: %v", err)
	}

	resp, err := helixClient.RefreshUserAccessToken(cfg.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %v", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get token: status %d: %s", resp.StatusCode, resp.ErrorMessage)
	}

	accessToken := resp.Data.AccessToken
	helixClient.SetUserAccessToken(accessToken)

	return &Client{
		cfg:         cfg,
		userClient:  helixClient,
		accessToken: accessToken,
	}, nil
}

func (c *Client) GetUserIDByUsername(username string) (string, error) {
	resp, err := c.userClient.GetUsers(&helix.UsersParams{
		Logins: []string{username},
	})
	if err != nil {
		return "", fmt.Errorf("failed to get user info: %v", err)
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to get user info: status %d: %s", resp.StatusCode, resp.ErrorMessage)
	}

	if len(resp.Data.Users) == 0 {
		return "", fmt.Errorf("failed to get user info: no users found")
	}

	return resp.Data.Users[0].ID, nil
}

func (c *Client) SendMessage(channel, text string) error {
	broadcasterID, err := c.GetUserIDByUsername(channel)
	if err != nil {
		return fmt.Errorf("failed to get broadcaster id: %v", err)
	}

	senderID, err := c.GetUserIDByUsername(c.cfg.Username)
	if err != nil {
		return fmt.Errorf("failed to get sender id: %v", err)
	}

	resp, err := c.userClient.SendChatMessage(&helix.SendChatMessageParams{
		BroadcasterID: broadcasterID,
		SenderID:      senderID,
		Message:       text,
	})
	if err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to send message: status %d: %s", resp.StatusCode, resp.ErrorMessage)
	}

	return nil
}

func (c *Client) RefreshToken() {
	slog.Debug("Refreshing twitch access token",
		slog.String("username", c.cfg.Username),
	)

	resp, err := c.userClient.RefreshUserAccessToken(c.cfg.RefreshToken)
	if err != nil {
		slog.Error("Failed to refresh user access token", slog.Any("error", err))
		return
	}
	if resp.StatusCode != 200 {
		slog.Error("Failed to refresh access token", slog.Int("status", resp.StatusCode), slog.String("error", resp.ErrorMessage))
		return
	}

	c.tokenMutex.Lock()
	c.accessToken = resp.Data.AccessToken
	c.lastUpdated = time.Now()
	c.tokenMutex.Unlock()

	c.userClient.SetUserAccessToken(c.accessToken)

	slog.Debug("Twitch access token refreshed successfully",
		slog.String("username", c.cfg.Username),
	)
}

func (c *Client) ensureTokenIsReady() {
	c.tokenMutex.Lock()
	lastUpdated := c.lastUpdated
	c.tokenMutex.Unlock()

	if lastUpdated.IsZero() || time.Now().Sub(lastUpdated).Minutes() > 30 {
		c.RefreshToken()
	}
}

func (c *Client) GetAccessToken() string {
	c.tokenMutex.Lock()
	defer c.tokenMutex.Unlock()

	return c.accessToken
}
