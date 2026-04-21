package settings

import (
	"strings"

	"github.com/rofleksey/dredge/internal/entity"
)

func normalizeTwitchAccountLinkType(accountType string) string {
	if accountType == "bot" {
		return "bot"
	}

	return "main"
}

func normalizeChannelDiscoverySettings(s entity.ChannelDiscoverySettings) entity.ChannelDiscoverySettings {
	out := s

	var tags []string

	for _, t := range s.RequiredStreamTags {
		x := strings.TrimSpace(t)
		if x == "" {
			continue
		}

		tags = append(tags, x)
	}

	out.RequiredStreamTags = tags
	out.GameID = strings.TrimSpace(s.GameID)

	if out.PollIntervalSeconds < 60 {
		out.PollIntervalSeconds = 60
	}

	if out.MaxStreamPagesPerRun < 1 {
		out.MaxStreamPagesPerRun = 1
	}

	if out.MinLiveViewers < 0 {
		out.MinLiveViewers = 0
	}

	return out
}

