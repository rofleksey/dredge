package twitch

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rofleksey/dredge/internal/entity"
)

func TestBuildActivityTimelineSegments_ircPresence(t *testing.T) {
	t.Parallel()

	chID := int64(10)
	t0 := time.Date(2025, 3, 1, 12, 0, 0, 0, time.UTC)
	t1 := time.Date(2025, 3, 1, 12, 30, 0, 0, time.UTC)
	windowEnd := t1.Add(time.Hour)

	ev := []entity.UserActivityEvent{
		{
			ID:                  1,
			ChatterTwitchUserID: 1,
			EventType:           entity.UserActivityChatOnline,
			ChannelTwitchUserID: &chID,
			ChannelLogin:        "chan",
			CreatedAt:           t0,
		},
		{
			ID:                  2,
			ChatterTwitchUserID: 1,
			EventType:           entity.UserActivityChatOffline,
			ChannelTwitchUserID: &chID,
			ChannelLogin:        "chan",
			CreatedAt:           t1,
		},
	}

	segs := BuildActivityTimelineSegments(ev, windowEnd)
	if len(segs) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segs))
	}

	if segs[0].ChannelTwitchUserID != chID || segs[0].ChannelLogin != "chan" {
		t.Fatalf("unexpected segment: %+v", segs[0])
	}

	if !segs[0].Start.Equal(t0) || !segs[0].End.Equal(t1) {
		t.Fatalf("unexpected bounds: %+v", segs[0])
	}
}

func TestMergeActivityTimelineSegments_overlap(t *testing.T) {
	t.Parallel()

	chID := int64(5)
	a := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	b := time.Date(2025, 1, 1, 10, 30, 0, 0, time.UTC)

	segs := []entity.ActivityTimelineSegment{
		{ChannelTwitchUserID: chID, ChannelLogin: "x", Start: a, End: a.Add(20 * time.Minute)},
		{ChannelTwitchUserID: chID, ChannelLogin: "x", Start: a.Add(10 * time.Minute), End: b},
	}

	merged := mergeActivityTimelineSegments(segs)
	if len(merged) != 1 {
		t.Fatalf("expected 1 merged segment, got %d", len(merged))
	}

	if !merged[0].Start.Equal(a) || !merged[0].End.Equal(b) {
		t.Fatalf("unexpected merge: %+v", merged[0])
	}
}

func TestBuildActivityTimelineSegments_empty(t *testing.T) {
	t.Parallel()

	windowEnd := time.Now().UTC()
	segs := BuildActivityTimelineSegments(nil, windowEnd)
	require.Nil(t, segs)

	segs = BuildActivityTimelineSegments([]entity.UserActivityEvent{}, windowEnd)
	require.Nil(t, segs)
}

func TestBuildActivityTimelineSegments_openEndsAtWindow(t *testing.T) {
	t.Parallel()

	chID := int64(7)
	t0 := time.Date(2025, 4, 1, 10, 0, 0, 0, time.UTC)
	windowEnd := t0.Add(2 * time.Hour)

	ev := []entity.UserActivityEvent{
		{
			ID:                  1,
			ChatterTwitchUserID: 1,
			EventType:           entity.UserActivityChatOnline,
			ChannelTwitchUserID: &chID,
			ChannelLogin:        "chan",
			CreatedAt:           t0,
		},
	}

	segs := BuildActivityTimelineSegments(ev, windowEnd)
	require.Len(t, segs, 1)
	assert.Equal(t, windowEnd, segs[0].End)
	assert.True(t, segs[0].Start.Equal(t0))
}

func TestMergeActivityTimelineSegments_empty(t *testing.T) {
	t.Parallel()

	require.Nil(t, mergeActivityTimelineSegments(nil))
	require.Nil(t, mergeActivityTimelineSegments([]entity.ActivityTimelineSegment{}))
}
