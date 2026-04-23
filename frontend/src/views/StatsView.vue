<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue';
import { DefaultService } from '../api/generated';
import type { SystemStatsResponse, SystemStatsTables } from '../api/generated';
import { PageHeader } from '../components/core';
import { formatDateTime } from '../lib/dateTime';
import { notifyApiError } from '../lib/notifyApiError';

defineOptions({ name: 'StatsView' });

/** Must match server stats collector TTL. */
const POLL_MS = 5000;

const data = ref<SystemStatsResponse | null>(null);
const loading = ref(false);
let timer: ReturnType<typeof setInterval> | null = null;

const byteFmt = new Intl.NumberFormat(undefined, {
  notation: 'compact',
  maximumFractionDigits: 1,
});

function formatBytes(n: number): string {
  if (!Number.isFinite(n)) {
    return '—';
  }

  return `${byteFmt.format(n)} B`;
}

function formatPct(n: number | null): string {
  if (n === null || !Number.isFinite(n)) {
    return '—';
  }

  return `${n.toFixed(1)}%`;
}

function formatNum(n: number): string {
  if (!Number.isFinite(n)) {
    return '—';
  }

  return byteFmt.format(n);
}

const capturedLabel = computed(() => {
  if (!data.value) {
    return '';
  }

  return `Snapshot: ${formatDateTime(data.value.captured_at)}`;
});

const TABLE_ROWS: { key: keyof SystemStatsTables; label: string }[] = [
  { key: 'twitch_users', label: 'Twitch users' },
  { key: 'twitch_accounts_active', label: 'Twitch accounts (active)' },
  { key: 'twitch_accounts_all', label: 'Twitch accounts (all rows)' },
  { key: 'rules', label: 'Rules' },
  { key: 'notification_entries', label: 'Notification entries' },
  { key: 'streams', label: 'Streams (sessions)' },
  { key: 'streams_open', label: 'Streams (live / open)' },
  { key: 'chat_messages', label: 'Chat messages' },
  { key: 'channel_chatters', label: 'Channel chatter rows' },
  { key: 'user_activity_events', label: 'User activity events' },
  { key: 'twitch_user_helix_meta', label: 'Helix meta rows' },
  { key: 'twitch_user_channel_follows', label: 'Channel follow edges' },
  { key: 'user_followed_channels', label: 'Followed channels (synced)' },
  { key: 'channel_blacklist', label: 'Channel blacklist' },
  { key: 'rule_trigger_events', label: 'Rule trigger events' },
  { key: 'irc_joined_samples', label: 'IRC joined samples' },
  { key: 'twitch_discovery_candidates', label: 'Discovery candidates' },
  { key: 'twitch_discovery_denied', label: 'Discovery denied' },
  { key: 'ai_conversations', label: 'AI conversations' },
  { key: 'ai_messages', label: 'AI messages' },
];

async function load(): Promise<void> {
  loading.value = true;
  try {
    data.value = await DefaultService.getSystemStats();
  } catch (e) {
    notifyApiError(e, {
      id: 'stats-load',
      title: 'Stats',
      fallbackMessage: 'Could not load system stats.',
    });
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  void load();
  timer = setInterval(() => void load(), POLL_MS);
});

onUnmounted(() => {
  if (timer) {
    clearInterval(timer);
    timer = null;
  }
});
</script>

<template>
  <div class="page stats-page">
    <PageHeader title="System stats" :subtitle="capturedLabel" layout="inline" />
    <p class="muted muted--body stats-hint">Table counts and host metrics refresh at most every 5 seconds (server cache).</p>

    <p v-if="loading && !data" class="muted muted--body">Loading…</p>

    <template v-else-if="data">
      <section class="stats-section">
        <h2 class="stats-h2">Database</h2>
        <dl class="stats-grid">
          <template v-for="row in TABLE_ROWS" :key="row.key">
            <dt>{{ row.label }}</dt>
            <dd>{{ formatNum(data.tables[row.key]) }}</dd>
          </template>
        </dl>
      </section>

      <section class="stats-section">
        <h2 class="stats-h2">Go process</h2>
        <dl class="stats-grid">
          <dt>Goroutines</dt>
          <dd>{{ data.process.goroutines }}</dd>
          <dt>Heap alloc</dt>
          <dd>{{ formatBytes(data.process.heap_alloc_bytes) }}</dd>
          <dt>Heap sys</dt>
          <dd>{{ formatBytes(data.process.heap_sys_bytes) }}</dd>
          <dt>Sys (runtime)</dt>
          <dd>{{ formatBytes(data.process.sys_bytes) }}</dd>
          <dt>Total alloc (cumulative)</dt>
          <dd>{{ formatBytes(data.process.total_alloc_bytes) }}</dd>
          <dt>GC runs</dt>
          <dd>{{ data.process.num_gc }}</dd>
          <dt>GC CPU fraction</dt>
          <dd>{{ data.process.gc_cpu_fraction.toFixed(4) }}</dd>
        </dl>
      </section>

      <section class="stats-section">
        <h2 class="stats-h2">Host</h2>
        <dl class="stats-grid">
          <dt>CPU</dt>
          <dd>{{ formatPct(data.host.cpu_percent) }}</dd>
          <dt>Memory used</dt>
          <dd>
            {{ formatBytes(data.host.memory_used_bytes) }} /
            {{ formatBytes(data.host.memory_total_bytes) }}
            ({{ formatPct(data.host.memory_used_percent) }})
          </dd>
          <dt>Disk ({{ data.host.disk_path }})</dt>
          <dd>
            {{ formatBytes(data.host.disk_used_bytes) }} /
            {{ formatBytes(data.host.disk_total_bytes) }}
            ({{ formatPct(data.host.disk_used_percent) }})
          </dd>
        </dl>
      </section>

      <section class="stats-section">
        <h2 class="stats-h2">Caches &amp; pool</h2>
        <dl class="stats-grid">
          <dt>Helix user OAuth cache entries</dt>
          <dd>{{ data.caches.helix_user_oauth_cache_entries }}</dd>
          <dt>Helix app access token (warm)</dt>
          <dd>{{ data.caches.helix_app_access_token_cached ? 'yes' : 'no' }}</dd>
          <dt>Login limiter tracked IPs</dt>
          <dd>{{ data.caches.login_limiter_tracked_ips }}</dd>
          <dt>pgx acquired / idle / total / max</dt>
          <dd>
            {{ data.caches.pgx_acquired_conns }} / {{ data.caches.pgx_idle_conns }} /
            {{ data.caches.pgx_total_conns }} / {{ data.caches.pgx_max_conns }}
          </dd>
          <dt>pgx acquire / canceled acquire</dt>
          <dd>{{ data.caches.pgx_acquire_count }} / {{ data.caches.pgx_canceled_acquire_count }}</dd>
        </dl>
      </section>
    </template>
  </div>
</template>

<style scoped lang="scss">
.page {
  padding: 0.75rem;
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.stats-hint {
  margin: 0 0 0.5rem;
}

.stats-section {
  margin-top: 0.75rem;
}

.stats-h2 {
  margin: 0 0 0.4rem;
  font-size: 1rem;
  font-weight: 600;
  color: var(--text);
}

.stats-grid {
  display: grid;
  grid-template-columns: minmax(10rem, 1fr) auto;
  gap: 0.35rem 1rem;
  margin: 0;
  max-width: 48rem;

  dt {
    margin: 0;
    color: var(--text-muted);
    font-size: 0.88rem;
  }

  dd {
    margin: 0;
    font-variant-numeric: tabular-nums;
    text-align: right;
    justify-self: end;
  }
}
</style>
