package twitch

import "github.com/rofleksey/dredge/internal/service/twitch/live"

// LiveRuntime exposes the IRC/live runtime for wiring the rules engine.
func (s *Usecase) LiveRuntime() *live.Runtime {
	return s.live
}
