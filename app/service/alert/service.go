package alert

import (
	"dredge/app/client/twitch_irc"
	"dredge/app/config"
	"dredge/app/util/telemetry"
	"fmt"
	"log/slog"
	"regexp"

	"github.com/elliotchance/pie/v2"
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
		if entry.Message == "" {
			entry.Message = ".*"
		}

		messageRegex, err := regexp.Compile(entry.Message)
		if err != nil {
			return nil, fmt.Errorf("invalid message regex '%s': %w", entry.Message, err)
		}

		selectors = append(selectors, Selector{
			AlertEntry: entry,
			MsgRegex:   messageRegex,
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
	if pie.Contains(s.cfg.Alert.ExcludeUsernames, username) {
		return false
	}

	for _, selector := range s.selectors {
		if pie.Contains(selector.AlertEntry.ExcludeChannels, channel) {
			continue
		}

		if pie.Contains(selector.AlertEntry.ExcludeUsernames, username) {
			continue
		}

		if !selector.MsgRegex.MatchString(text) {
			continue
		}

		return true
	}

	return false
}

func (s *Service) handleMessage(channel, username, _, text string, _ map[string]string) {
	if !s.isMessageAlertable(channel, username, text) {
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
