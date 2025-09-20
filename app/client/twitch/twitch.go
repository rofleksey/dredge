package twitch

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"log/slog"

	"github.com/gempir/go-twitch-irc/v4"
	"github.com/nicklaw5/helix/v2"
)

type MessageHandler func(channel, username, messageID, text string, tags map[string]string)

type Client struct {
	cfg               Config
	userClient        *helix.Client
	ircClient         *twitch.Client
	messageHandler    MessageHandler
	refreshMutex      sync.Mutex
	connectedChannels map[string]bool
	channelsMutex     sync.RWMutex
}

type Config struct {
	ClientID     string
	ClientSecret string
	RefreshToken string
}

func NewClient(cfg Config, messageHandler MessageHandler) (*Client, error) {
	if messageHandler == nil {
		panic("MessageHandler must not be nil")
	}

	client := &Client{
		cfg:               cfg,
		messageHandler:    messageHandler,
		connectedChannels: make(map[string]bool),
	}

	userClient, accessToken, err := client.createUserClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create user client: %v", err)
	}
	client.userClient = userClient

	client.ircClient = twitch.NewClient(cfg.ClientID, "oauth:"+accessToken)
	client.setupIRCListeners()

	return client, nil
}

func (c *Client) createUserClient() (*helix.Client, string, error) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:    10,
			IdleConnTimeout: 30 * time.Second,
		},
		Timeout: 10 * time.Second,
	}

	client, err := helix.NewClient(&helix.Options{
		ClientID:     c.cfg.ClientID,
		ClientSecret: c.cfg.ClientSecret,
		RefreshToken: c.cfg.RefreshToken,
		HTTPClient:   httpClient,
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to create helix client: %v", err)
	}

	resp, err := client.RefreshUserAccessToken(c.cfg.RefreshToken)
	if err != nil {
		return nil, "", fmt.Errorf("failed to refresh access token: %v", err)
	}
	if resp.StatusCode != 200 {
		return nil, "", fmt.Errorf("failed to refresh token: status %d: %s", resp.StatusCode, resp.ErrorMessage)
	}

	accessToken := resp.Data.AccessToken
	client.SetUserAccessToken(accessToken)
	return client, accessToken, nil
}

func (c *Client) setupIRCListeners() {
	c.ircClient.OnPrivateMessage(func(message twitch.PrivateMessage) {
		username := strings.ToLower(message.User.Name)
		channel := strings.TrimPrefix(message.Channel, "#")
		text := strings.TrimSpace(message.Message)

		c.messageHandler(channel, username, message.ID, text, message.Tags)
	})

	c.ircClient.OnConnect(func() {
		slog.Info("Connected to Twitch IRC")
	})

	c.ircClient.OnReconnectMessage(func(message twitch.ReconnectMessage) {
		slog.Info("Reconnecting to Twitch IRC")
	})
}

func (c *Client) Connect() error {
	return c.ircClient.Connect()
}

func (c *Client) Disconnect() {
	c.ircClient.Disconnect()
}

func (c *Client) JoinChannel(channel string) {
	c.channelsMutex.Lock()
	defer c.channelsMutex.Unlock()

	if !c.connectedChannels[channel] {
		c.ircClient.Join(channel)
		c.connectedChannels[channel] = true
		slog.Info("Joined channel", slog.String("channel", channel))
	}
}

func (c *Client) LeaveChannel(channel string) {
	c.channelsMutex.Lock()
	defer c.channelsMutex.Unlock()

	if c.connectedChannels[channel] {
		c.ircClient.Depart(channel)
		delete(c.connectedChannels, channel)
		slog.Info("Left channel", slog.String("channel", channel))
	}
}

func (c *Client) TokenRefreshLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.refreshTokens()
		}
	}
}

func (c *Client) refreshTokens() {
	c.refreshMutex.Lock()
	defer c.refreshMutex.Unlock()

	slog.Debug("Refreshing Twitch tokens")

	resp, err := c.userClient.RefreshUserAccessToken(c.cfg.RefreshToken)
	if err != nil {
		slog.Error("Failed to refresh user token", slog.Any("error", err))
		return
	}

	if resp.StatusCode != 200 {
		slog.Error("Failed to refresh token", slog.Int("status", resp.StatusCode), slog.String("error", resp.ErrorMessage))
		return
	}

	accessToken := resp.Data.AccessToken
	c.userClient.SetUserAccessToken(accessToken)
	c.ircClient.SetIRCToken("oauth:" + accessToken)
	slog.Debug("Twitch tokens refreshed successfully")
}
