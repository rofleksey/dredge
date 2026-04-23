package handler

import (
	"context"
	"errors"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/http/gen"
)

func (h *Handler) GetSystemStats(ctx context.Context) (gen.GetSystemStatsRes, error) {
	if h.stats == nil {
		return nil, errors.New("stats collector not configured")
	}

	snap, err := h.stats.Get(ctx)
	if err != nil {
		return nil, err
	}

	return systemStatsSnapshotToGen(snap), nil
}

func systemStatsSnapshotToGen(s entity.SystemStatsSnapshot) *gen.SystemStatsResponse {
	t := s.Tables

	tables := gen.SystemStatsTables{
		TwitchUsers:               t.TwitchUsers,
		TwitchAccountsActive:      t.TwitchAccountsActive,
		TwitchAccountsAll:         t.TwitchAccountsAll,
		Rules:                     t.Rules,
		NotificationEntries:       t.NotificationEntries,
		Streams:                   t.Streams,
		StreamsOpen:               t.StreamsOpen,
		ChatMessages:              t.ChatMessages,
		ChannelChatters:           t.ChannelChatters,
		UserActivityEvents:        t.UserActivityEvents,
		TwitchUserHelixMeta:       t.TwitchUserHelixMeta,
		TwitchUserChannelFollows:  t.TwitchUserChannelFollows,
		UserFollowedChannels:      t.UserFollowedChannels,
		ChannelBlacklist:          t.ChannelBlacklist,
		RuleTriggerEvents:         t.RuleTriggerEvents,
		IrcJoinedSamples:          t.IrcJoinedSamples,
		TwitchDiscoveryCandidates: t.TwitchDiscoveryCandidates,
		TwitchDiscoveryDenied:     t.TwitchDiscoveryDenied,
		AiConversations:           t.AiConversations,
		AiMessages:                t.AiMessages,
	}

	p := s.Process

	process := gen.SystemStatsProcess{
		Goroutines:      int32(p.Goroutines),
		HeapAllocBytes:  p.HeapAllocBytes,
		HeapSysBytes:    p.HeapSysBytes,
		SysBytes:        p.SysBytes,
		TotalAllocBytes: p.TotalAllocBytes,
		NumGc:           p.NumGC,
		GcCPUFraction:   p.GCCPUFraction,
	}

	host := gen.SystemStatsHost{
		MemoryTotalBytes:  s.Host.MemoryTotalBytes,
		MemoryUsedBytes:   s.Host.MemoryUsedBytes,
		MemoryUsedPercent: s.Host.MemoryUsedPercent,
		DiskPath:          s.Host.DiskPath,
		DiskTotalBytes:    s.Host.DiskTotalBytes,
		DiskUsedBytes:     s.Host.DiskUsedBytes,
		DiskUsedPercent:   s.Host.DiskUsedPercent,
	}

	if s.Host.CPUPercent != nil {
		host.SetCPUPercent(gen.NewNilFloat64(*s.Host.CPUPercent))
	} else {
		var n gen.NilFloat64

		n.SetToNull()
		host.SetCPUPercent(n)
	}

	cc := s.Caches

	caches := gen.SystemStatsCaches{
		HelixUserOAuthCacheEntries: int32(cc.HelixUserOAuthCacheEntries),
		HelixAppAccessTokenCached:  cc.HelixAppAccessTokenCached,
		LoginLimiterTrackedIps:     int32(cc.LoginLimiterTrackedIPs),
		PgxAcquiredConns:           cc.PgxAcquiredConns,
		PgxIdleConns:               cc.PgxIdleConns,
		PgxTotalConns:              cc.PgxTotalConns,
		PgxMaxConns:                cc.PgxMaxConns,
		PgxAcquireCount:            cc.PgxAcquireCount,
		PgxCanceledAcquireCount:    cc.PgxCanceledAcquireCount,
	}

	return &gen.SystemStatsResponse{
		CapturedAt: s.CapturedAt,
		Tables:     tables,
		Process:    process,
		Host:       host,
		Caches:     caches,
	}
}
