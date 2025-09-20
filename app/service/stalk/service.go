package stalk

import (
	"context"
	"dredge/app/client/twitch"
	"dredge/pkg/config"
	"dredge/pkg/database"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/elliotchance/pie/v2"
	"github.com/samber/do"
)

var _ do.Shutdownable = (*Service)(nil)

var nonAlphaNumericRegexp = regexp.MustCompile(`[^a-zA-Z0-9\s]+`)

type Service struct {
	appCtx  context.Context
	cfg     *config.Config
	queries *database.Queries

	client *twitch.Client
}

func New(di *do.Injector) (*Service, error) {
	cfg := do.MustInvoke[*config.Config](di)

	service := &Service{
		appCtx:  do.MustInvoke[context.Context](di),
		queries: do.MustInvoke[*database.Queries](di),
		cfg:     cfg,
	}

	client, err := twitch.NewClient(twitch.Config{
		ClientID:     cfg.Twitch.ClientID,
		ClientSecret: cfg.Twitch.ClientSecret,
		RefreshToken: cfg.Twitch.RefreshToken,
	}, service.HandleMessage)
	if err != nil {
		return nil, fmt.Errorf("twitch.NewClient: %w", err)
	}
	service.client = client

	return service, nil
}

func (s *Service) Run() error {
	for _, streamer := range s.cfg.Stalk.Streamers {
		s.client.JoinChannel(streamer)
	}

	go s.client.TokenRefreshLoop(s.appCtx)

	if err := s.client.Connect(); err != nil {
		return fmt.Errorf("client.Connect: %w", err)
	}

	return nil
}

func (s *Service) isInterestingMessage(text string) bool {
	cleaned := nonAlphaNumericRegexp.ReplaceAllString(text, "")
	words := strings.Fields(cleaned)

	for _, word := range words {
		if pie.Contains(s.cfg.Stalk.Keywords, word) {
			return true
		}
	}

	for _, substr := range s.cfg.Stalk.Substrings {
		if strings.Contains(text, substr) {
			return true
		}
	}

	return false
}

func (s *Service) HandleMessage(channel, username, messageID, text string, tags map[string]string) {
	slog.Debug("Message",
		slog.String("channel", channel),
		slog.String("username", username),
		slog.String("message_id", messageID),
		slog.String("text", text),
	)

	if err := s.queries.CreateMessage(s.appCtx, database.CreateMessageParams{
		ID:       messageID,
		Created:  time.Now(),
		Channel:  channel,
		Username: username,
		Text:     text,
	}); err != nil {
		slog.Error("CreateMessage",
			slog.String("channel", channel),
			slog.String("username", username),
			slog.String("message_id", messageID),
			slog.String("text", text),
			slog.Any("error", err),
		)
	}

	if !s.isInterestingMessage(text) {
		return
	}

	slog.Error(text,
		slog.String("channel", channel),
		slog.String("username", username),
	)
}

func (s *Service) Shutdown() error {
	s.client.Disconnect()

	return nil
}
