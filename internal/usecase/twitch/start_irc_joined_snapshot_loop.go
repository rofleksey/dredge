package twitch

import (
	"context"
	"time"

	"go.uber.org/zap"
)

const ircJoinedSnapshotInterval = 30 * time.Minute

// StartIrcJoinedSnapshotLoop records IRC joined counts on an interval until ctx is cancelled.
func (s *Usecase) StartIrcJoinedSnapshotLoop(ctx context.Context) {
	ticker := time.NewTicker(ircJoinedSnapshotInterval)
	defer ticker.Stop()

	s.runOneIrcJoinedSnapshot("startup")

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.runOneIrcJoinedSnapshot("tick")
		}
	}
}

func (s *Usecase) runOneIrcJoinedSnapshot(reason string) {
	snapCtx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	if err := s.RecordIrcJoinedSnapshot(snapCtx); err != nil {
		s.obs.Logger.Warn("irc joined snapshot failed", zap.String("reason", reason), zap.Error(err))
	}
}
