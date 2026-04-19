package settings

import (
	"context"
)

func (s *Service) ListChannelBlacklist(ctx context.Context) ([]string, error) {
	ctx, span := s.obs.StartSpan(ctx, "usecase.settings.list_channel_blacklist")
	defer span.End()

	out, err := s.repo.ListChannelBlacklist(ctx)
	if err != nil {
		s.obs.LogError(ctx, span, "list channel blacklist failed", err)
	}

	return out, err
}
