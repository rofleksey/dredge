package twitch

import "time"

// WatchUiHints exposes poll intervals for the SPA (seconds, minimum 1).
func (s *Service) WatchUiHints() (viewerPollSec int, channelChattersSyncSec int, monitoredLivePollSec int) {
	v := int(s.viewerPollInterval / time.Second)
	c := int(s.channelChattersSyncInterval / time.Second)
	m := int(s.streamSessionPollInterval / time.Second)

	if v < 1 {
		v = 10
	}

	if c < 1 {
		c = 10
	}

	if m < 1 {
		m = 60
	}

	return v, c, m
}
