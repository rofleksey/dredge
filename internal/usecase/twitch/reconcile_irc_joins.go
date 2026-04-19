package twitch

import "context"

// ReconcileIRCJoins updates IRC Join/Depart set from current settings without restarting the IRC connection.
func (s *Service) ReconcileIRCJoins(ctx context.Context) {
	s.live.ReconcileIRCJoins(ctx)
}
