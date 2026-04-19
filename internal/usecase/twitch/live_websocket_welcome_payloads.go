package twitch

import "context"

// LiveWebSocketWelcomePayloads returns messages sent to a browser immediately after the live WebSocket upgrade.
func (s *Service) LiveWebSocketWelcomePayloads(ctx context.Context) ([]any, error) {
	msg, err := s.live.LiveWebSocketWelcomePayloads(ctx)
	if err != nil {
		return nil, err
	}

	return []any{msg}, nil
}
