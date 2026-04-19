package twitch

import "context"

// SetPersistContext sets the parent context for IRC-driven DB work; cancel it on app shutdown.
func (s *Usecase) SetPersistContext(ctx context.Context) {
	s.persistMu.Lock()
	defer s.persistMu.Unlock()

	s.persistCtx = ctx
}

func (s *Usecase) persistContext() context.Context {
	s.persistMu.RLock()
	defer s.persistMu.RUnlock()

	if s.persistCtx != nil {
		return s.persistCtx
	}

	return context.Background()
}
