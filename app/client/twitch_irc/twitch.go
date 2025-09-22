package twitch_irc

import (
	"context"
	"dredge/app/client/twitch_api"
	"fmt"
	"strings"
	"sync"
	"time"

	"log/slog"

	"github.com/gempir/go-twitch-irc/v4"
)

type MessageHandler func(channel, username, messageID, text string, tags map[string]string)

type Client struct {
	cfg            Config
	apiClient      *twitch_api.Client
	ircClient      *twitch.Client
	messageHandler MessageHandler

	connectedChannels map[string]bool
	channelsMutex     sync.RWMutex
}

type Config struct {
	ClientID     string
	ClientSecret string
	Username     string
	RefreshToken string
}

func NewClient(cfg Config, messageHandler MessageHandler) (*Client, error) {
	if messageHandler == nil {
		panic("MessageHandler must not be nil")
	}

	apiClient, err := twitch_api.NewClient(twitch_api.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Username:     cfg.Username,
		RefreshToken: cfg.RefreshToken,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user client: %v", err)
	}

	client := &Client{
		cfg:               cfg,
		apiClient:         apiClient,
		messageHandler:    messageHandler,
		connectedChannels: make(map[string]bool),
	}

	client.ircClient = twitch.NewClient(cfg.ClientID, "oauth:"+apiClient.GetAccessToken())
	client.setupIRCListeners()

	return client, nil
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
			c.RefreshTokens()
		}
	}
}

func (c *Client) RefreshTokens() {
	c.apiClient.RefreshToken()

	newToken := c.apiClient.GetAccessToken()
	c.ircClient.SetIRCToken("oauth:" + newToken)
}
