package twitch

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStartOfWeekMondayUTC(t *testing.T) {
	t.Parallel()

	// Wednesday 2025-04-09 UTC -> week starts Monday 2025-04-07
	wed := time.Date(2025, 4, 9, 15, 30, 0, 0, time.UTC)
	start := startOfWeekMondayUTC(wed)
	assert.Equal(t, 2025, start.Year())
	assert.Equal(t, time.April, start.Month())
	assert.Equal(t, 7, start.Day())
	assert.Equal(t, 0, start.Hour())
}
