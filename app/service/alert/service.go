package alert

import (
	"dredge/app/client/twitch_irc"
	"dredge/app/config"
	"dredge/app/util/telemetry"
	"fmt"
	"log/slog"
	"regexp"

	"github.com/samber/do"
)

var serviceName = "alert"

type Service struct {
	cfg       *config.Config
	tracing   *telemetry.Tracing
	ircClient *twitch_irc.Client

	selectors []Selector
}

func New(di *do.Injector) (*Service, error) {
	cfg := do.MustInvoke[*config.Config](di)

	selectors := make([]Selector, 0, len(cfg.Alert.List))
	for _, entry := range cfg.Alert.List {
		if entry.Channel == "" {
			entry.Channel = ".*"
		}
		if entry.Username == "" {
			entry.Username = ".*"
		}
		if entry.Message == "" {
			entry.Message = ".*"
		}

		channelRegex, err := regexp.Compile(entry.Channel)
		if err != nil {
			return nil, fmt.Errorf("invalid channel regex '%s': %w", entry.Channel, err)
		}

		usernameRegex, err := regexp.Compile(entry.Username)
		if err != nil {
			return nil, fmt.Errorf("invalid username regex '%s': %w", entry.Username, err)
		}

		messageRegex, err := regexp.Compile(entry.Message)
		if err != nil {
			return nil, fmt.Errorf("invalid message regex '%s': %w", entry.Message, err)
		}

		selectors = append(selectors, Selector{
			Channel:  channelRegex,
			Username: usernameRegex,
			Message:  messageRegex,
		})
	}

	return &Service{
		cfg:       cfg,
		tracing:   do.MustInvoke[*telemetry.Tracing](di),
		ircClient: do.MustInvoke[*twitch_irc.Client](di),

		selectors: selectors,
	}, nil
}

func (s *Service) isMessageAlertable(channel, username, text string) bool {
	for _, excludeUsername := range s.cfg.Alert.ExcludeUsernames {
		if username == excludeUsername {
			return false
		}
	}

	for _, selector := range s.selectors {
		if selector.Channel.MatchString(channel) ||
			selector.Username.MatchString(username) ||
			selector.Message.MatchString(text) {
			return true
		}
	}

	return false
}

func (s *Service) handleMessage(channel, username, messageID, text string, tags map[string]string) {
	if !s.isMessageAlertable(channel, username, messageID) {
		return
	}

	slog.Info(text,
		slog.String("channel", channel),
		slog.String("username", username),
		slog.Bool("telegram", true),
	)
}

func (s *Service) Init() {
	s.ircClient.SetListener(s.handleMessage)
}
