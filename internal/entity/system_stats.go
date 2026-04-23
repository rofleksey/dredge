package entity

import "time"

// SystemStatsTableCounts holds row counts from a single aggregated SQL query.
type SystemStatsTableCounts struct {
	TwitchUsers               int64
	TwitchAccountsActive      int64
	TwitchAccountsAll         int64
	Rules                     int64
	NotificationEntries       int64
	Streams                   int64
	StreamsOpen               int64
	ChatMessages              int64
	ChannelChatters           int64
	UserActivityEvents        int64
	TwitchUserHelixMeta       int64
	TwitchUserChannelFollows  int64
	UserFollowedChannels      int64
	ChannelBlacklist          int64
	RuleTriggerEvents         int64
	IrcJoinedSamples          int64
	TwitchDiscoveryCandidates int64
	TwitchDiscoveryDenied     int64
	AiConversations           int64
	AiMessages                int64
}

// SystemStatsProcess holds Go runtime memory and scheduler metrics.
type SystemStatsProcess struct {
	Goroutines      int
	HeapAllocBytes  int64
	HeapSysBytes    int64
	SysBytes        int64
	TotalAllocBytes int64
	NumGC           int32
	GCCPUFraction   float64
}

// SystemStatsHost holds host-level CPU, memory, and disk usage for the machine running the process.
type SystemStatsHost struct {
	CPUPercent        *float64
	MemoryTotalBytes  int64
	MemoryUsedBytes   int64
	MemoryUsedPercent float64
	DiskPath          string
	DiskTotalBytes    int64
	DiskUsedBytes     int64
	DiskUsedPercent   float64
}

// SystemStatsCaches holds in-process cache and pool telemetry.
type SystemStatsCaches struct {
	HelixUserOAuthCacheEntries int
	HelixAppAccessTokenCached  bool
	LoginLimiterTrackedIPs     int
	PgxAcquiredConns           int32
	PgxIdleConns               int32
	PgxTotalConns              int32
	PgxMaxConns                int32
	PgxAcquireCount            int64
	PgxCanceledAcquireCount    int64
}

// SystemStatsSnapshot is the full payload after one collector refresh.
type SystemStatsSnapshot struct {
	CapturedAt time.Time
	Tables     SystemStatsTableCounts
	Process    SystemStatsProcess
	Host       SystemStatsHost
	Caches     SystemStatsCaches
}
