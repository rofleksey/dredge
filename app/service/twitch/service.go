package twitch

import (
	"context"
	"dredge/app/client/twitch_api"
	"dredge/app/client/twitch_irc"
	"dredge/app/util/telemetry"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/avast/retry-go"
	"github.com/nicklaw5/helix/v2"
	"github.com/samber/do"
	"github.com/samber/oops"
)

var serviceName = "twitch"

type Service struct {
	tracing   *telemetry.Tracing
	apiClient *twitch_api.Client
	ircClient *twitch_irc.Client

	oldChannelMap map[string]struct{}
}

func New(di *do.Injector) (*Service, error) {
	return &Service{
		tracing:       do.MustInvoke[*telemetry.Tracing](di),
		apiClient:     do.MustInvoke[*twitch_api.Client](di),
		ircClient:     do.MustInvoke[*twitch_irc.Client](di),
		oldChannelMap: make(map[string]struct{}),
	}, nil
}

func (s *Service) doFetch(ctx context.Context) error {
	slog.Debug("Starting fetch")

	started := time.Now()

	var after string
	newChannelMap := make(map[string]struct{})

	for {
		chunk, err := s.fetchChunkWithRetry(ctx, after)
		if err != nil {
			return oops.Errorf("fetchChunkWithRetry: %w", err)
		}
		if chunk.Pagination.Cursor == "" {
			break
		}

		for _, stream := range chunk.Streams {
			streamName := strings.ToLower(stream.UserLogin)
			newChannelMap[streamName] = struct{}{}
		}

		after = chunk.Pagination.Cursor

		select {
		case <-ctx.Done():
			return oops.Errorf("fetchAllLiveStreams: context canceled")
		case <-time.After(3 * time.Second):
		}
	}

	slog.Debug("Fetch finished",
		slog.Duration("duration", time.Since(started)),
		slog.Int("count", len(newChannelMap)),
	)

	for streamName := range s.oldChannelMap {
		_, stillExists := newChannelMap[streamName]
		if stillExists {
			continue
		}

		s.ircClient.LeaveChannel(streamName)
	}

	for streamName := range newChannelMap {
		_, isNotNew := s.oldChannelMap[streamName]
		if isNotNew {
			continue
		}

		s.ircClient.JoinChannel(streamName)
	}

	s.oldChannelMap = newChannelMap

	return nil
}

func (s *Service) fetchChunkWithRetry(ctx context.Context, after string) (*helix.ManyStreams, error) {
	var result *helix.ManyStreams

	attempts := 3

	err := retry.Do(func() error {
		chunk, err := s.apiClient.GetLiveDBDStreams(after)
		if err != nil {
			return oops.Errorf("GetLiveDBDStreams: %w", err)
		}

		result = &chunk

		return nil
	}, retry.Context(ctx), retry.Attempts(uint(attempts)), retry.Delay(time.Second*5))
	if err != nil {
		return nil, fmt.Errorf("retry.Do: %w", err)
	}

	return result, nil
}

func (s *Service) RunFetchLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if err := s.doFetch(ctx); err != nil {
			slog.ErrorContext(ctx, "Failed to fetch streams",
				slog.Any("error", err),
			)
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Minute):
		}
	}
}
