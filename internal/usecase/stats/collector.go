package stats

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"

	"github.com/rofleksey/dredge/internal/entity"
	httpmw "github.com/rofleksey/dredge/internal/http/middleware"
	"github.com/rofleksey/dredge/internal/repository"
)

const cacheTTL = 5 * time.Second

// HelixOAuthCache is implemented by the Twitch Helix client for operator stats.
type HelixOAuthCache interface {
	StatsSnapshot() (userOAuthEntries int, appAccessTokenCached bool)
}

// Collector builds a cached system stats snapshot (TTL 5s) for operator dashboards.
type Collector struct {
	store   repository.Store
	helix   HelixOAuthCache
	limiter *httpmw.LoginLimiter
	pool    *pgxpool.Pool

	mu          sync.Mutex
	last        entity.SystemStatsSnapshot
	lastValid   bool
	lastRefresh time.Time
}

// NewCollector wires DB, Helix, login limiter, and pgx pool for snapshots. pool may be nil (pgx stats omitted).
func NewCollector(store repository.Store, hx HelixOAuthCache, limiter *httpmw.LoginLimiter, pool *pgxpool.Pool) *Collector {
	return &Collector{
		store:   store,
		helix:   hx,
		limiter: limiter,
		pool:    pool,
	}
}

// Get returns the latest snapshot, refreshing if the cache is older than cacheTTL.
func (c *Collector) Get(ctx context.Context) (entity.SystemStatsSnapshot, error) {
	now := time.Now()

	c.mu.Lock()
	if c.lastValid && now.Sub(c.lastRefresh) < cacheTTL {
		out := c.last
		c.mu.Unlock()

		return out, nil
	}
	c.mu.Unlock()

	tables, err := c.store.SystemStatsTableCounts(ctx)
	if err != nil {
		return entity.SystemStatsSnapshot{}, err
	}

	var ms runtime.MemStats

	runtime.ReadMemStats(&ms)

	proc := entity.SystemStatsProcess{
		Goroutines:      runtime.NumGoroutine(),
		HeapAllocBytes:  int64(ms.HeapAlloc),
		HeapSysBytes:    int64(ms.HeapSys),
		SysBytes:        int64(ms.Sys),
		TotalAllocBytes: int64(ms.TotalAlloc),
		NumGC:           int32(ms.NumGC),
		GCCPUFraction:   ms.GCCPUFraction,
	}

	host := entity.SystemStatsHost{
		DiskPath: diskRootPath(),
	}

	if vm, err := mem.VirtualMemory(); err == nil && vm != nil {
		host.MemoryTotalBytes = int64(vm.Total)
		host.MemoryUsedBytes = int64(vm.Used)
		host.MemoryUsedPercent = vm.UsedPercent
	}

	cpuPercents, err := cpu.Percent(150*time.Millisecond, false)
	if err == nil && len(cpuPercents) > 0 {
		v := cpuPercents[0]
		host.CPUPercent = &v
	}

	if du, err := disk.Usage(host.DiskPath); err == nil && du != nil {
		host.DiskTotalBytes = int64(du.Total)
		host.DiskUsedBytes = int64(du.Used)
		host.DiskUsedPercent = du.UsedPercent
	}

	helixN, helixWarm := 0, false
	if c.helix != nil {
		helixN, helixWarm = c.helix.StatsSnapshot()
	}

	caches := entity.SystemStatsCaches{
		HelixUserOAuthCacheEntries: helixN,
		HelixAppAccessTokenCached:  helixWarm,
		LoginLimiterTrackedIPs:     c.limiter.TrackedIPCount(),
	}

	if c.pool != nil {
		st := c.pool.Stat()
		if st != nil {
			caches.PgxAcquiredConns = st.AcquiredConns()
			caches.PgxIdleConns = st.IdleConns()
			caches.PgxTotalConns = st.TotalConns()
			caches.PgxMaxConns = st.MaxConns()
			caches.PgxAcquireCount = st.AcquireCount()
			caches.PgxCanceledAcquireCount = st.CanceledAcquireCount()
		}
	}

	snap := entity.SystemStatsSnapshot{
		CapturedAt: now.UTC(),
		Tables:     tables,
		Process:    proc,
		Host:       host,
		Caches:     caches,
	}

	c.mu.Lock()
	c.last = snap
	c.lastValid = true
	c.lastRefresh = now
	out := c.last
	c.mu.Unlock()

	return out, nil
}

func diskRootPath() string {
	if runtime.GOOS == "windows" {
		d := os.Getenv("SystemDrive")
		if d == "" {
			d = "C:"
		}

		return d + string(filepath.Separator)
	}

	return "/"
}
