package twitch

import (
	"testing"
	"time"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/stretchr/testify/assert"
)

func TestPresenceSecondsClipped(t *testing.T) {
	t.Parallel()

	winStart := time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC)
	winEnd := time.Date(2020, 1, 1, 13, 0, 0, 0, time.UTC)

	segs := []entity.ActivityTimelineSegment{
		{Start: winStart, End: winEnd},
	}

	sec := presenceSecondsClipped(segs, winStart, winEnd)
	assert.Equal(t, int64(3600), sec)
}
