package accounts

import (
	"context"
	"dredge/app/client/twitch_api"
	"dredge/pkg/config"
	"fmt"

	"github.com/samber/do"
)

type Service struct {
	appCtx context.Context
	cfg    *config.Config

	clients map[string]*twitch_api.Client
}

func New(di *do.Injector) (*Service, error) {
	cfg := do.MustInvoke[*config.Config](di)

	clients := make(map[string]*twitch_api.Client)

	for _, account := range cfg.Accounts {
		client, err := twitch_api.NewClient(twitch_api.Config{
			ClientID:     cfg.Twitch.ClientID,
			ClientSecret: cfg.Twitch.ClientSecret,
			Username:     account.Username,
			RefreshToken: account.RefreshToken,
		})
		if err != nil {
			return nil, fmt.Errorf("twitch_api.NewClient: %w", err)
		}

		clients[account.Username] = client
	}

	return &Service{
		appCtx:  do.MustInvoke[context.Context](di),
		cfg:     cfg,
		clients: clients,
	}, nil
}

func (s *Service) SendMessage(channel, username, text string) error {
	client, ok := s.clients[username]
	if !ok {
		return fmt.Errorf("client not found for %s", username)
	}

	if err := client.SendMessage(channel, text); err != nil {
		return fmt.Errorf("client.SendMessage: %w", err)
	}

	return nil
}
