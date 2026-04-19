package twitch

import "context"

// PersistContext returns the long-lived context used for IRC-driven persistence.
func (s *Usecase) PersistContext() context.Context {
	return s.persistContext()
}
