<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { ApiError, DefaultService } from '../api/generated';
import type { NotificationEntry } from '../api/generated';
import { notify } from '../lib/notify';

defineOptions({ name: 'NotificationsView' });

const pageSize = 50;
const loading = ref(false);
const loadingMore = ref(false);
const hasMore = ref(true);
const entries = ref<NotificationEntry[]>([]);

const totalLoadedLabel = computed(() => `${entries.value.length.toLocaleString()} loaded`);

function settingPreview(e: NotificationEntry): string {
  const raw = JSON.stringify(e.settings ?? {});
  if (!raw) {
    return '{}';
  }
  if (raw.length <= 220) {
    return raw;
  }
  return `${raw.slice(0, 217)}...`;
}

function createdAtLabel(ts: string): string {
  const d = new Date(ts);
  if (!Number.isFinite(d.getTime())) {
    return ts;
  }
  return d.toLocaleString();
}

async function fetchFirst(): Promise<void> {
  loading.value = true;
  try {
    const list = await DefaultService.listNotifications({ limit: pageSize });
    entries.value = list;
    hasMore.value = list.length === pageSize;
  } catch (e) {
    entries.value = [];
    hasMore.value = false;
    const msg =
      e instanceof ApiError && e.body && typeof e.body.message === 'string'
        ? e.body.message
        : 'Could not load notifications.';
    notify({
      id: 'notifications-load',
      type: 'error',
      title: 'Notifications',
      description: msg,
    });
  } finally {
    loading.value = false;
  }
}

async function fetchMore(): Promise<void> {
  if (!hasMore.value || loadingMore.value || entries.value.length === 0) {
    return;
  }

  loadingMore.value = true;
  try {
    const last = entries.value[entries.value.length - 1];
    const next = await DefaultService.listNotifications({
      limit: pageSize,
      cursorCreatedAt: last.created_at,
      cursorId: last.id,
    });
    const seen = new Set(entries.value.map((x) => x.id));
    for (const row of next) {
      if (!seen.has(row.id)) {
        entries.value.push(row);
      }
    }
    hasMore.value = next.length === pageSize;
  } catch {
    notify({
      id: 'notifications-more',
      type: 'error',
      title: 'Notifications',
      description: 'Could not load more notifications.',
    });
  } finally {
    loadingMore.value = false;
  }
}

onMounted(() => {
  void fetchFirst();
});
</script>

<template>
  <div class="page">
    <header class="head">
      <h1>Notifications</h1>
      <span class="muted small">{{ totalLoadedLabel }}</span>
    </header>

    <p v-if="loading" class="muted">Loading...</p>
    <p v-else-if="entries.length === 0" class="muted">No notifications.</p>

    <ul v-else class="list">
      <li v-for="e in entries" :key="e.id" class="item">
        <div class="item-top">
          <span class="tag">{{ e.provider }}</span>
          <span class="status" :class="{ 'status-off': !e.enabled }">{{ e.enabled ? 'enabled' : 'disabled' }}</span>
        </div>
        <p class="meta">#{{ e.id }} · {{ createdAtLabel(e.created_at) }}</p>
        <p class="settings">{{ settingPreview(e) }}</p>
      </li>
    </ul>

    <div v-if="!loading && entries.length > 0" class="more-row">
      <button type="button" class="btn-more" :disabled="loadingMore || !hasMore" @click="fetchMore">
        {{ loadingMore ? 'Loading...' : hasMore ? 'Show more' : 'No more notifications' }}
      </button>
    </div>
  </div>
</template>

<style scoped lang="scss">
.page {
  padding: 0.75rem;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 0.5rem;
  margin-bottom: 0.75rem;
}

h1 {
  margin: 0;
  font-size: 1.15rem;
}

.list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.item {
  border: 1px solid var(--border);
  border-radius: 0.35rem;
  background: var(--bg-elevated);
  padding: 0.55rem 0.6rem;
}

.item-top {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.tag {
  font-size: 0.76rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--accent-bright);
}

.status {
  font-size: 0.78rem;
  color: #2ecc71;
}

.status-off {
  color: var(--text-muted);
}

.meta {
  margin: 0.25rem 0 0;
  font-size: 0.78rem;
  color: var(--text-muted);
}

.settings {
  margin: 0.35rem 0 0;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 0.78rem;
  color: var(--text);
  white-space: pre-wrap;
  word-break: break-word;
}

.more-row {
  margin-top: 0.65rem;
}

.btn-more {
  padding: 0.42rem 0.85rem;
  border: 1px solid var(--border);
  border-radius: 0.25rem;
  background: var(--bg-base);
  color: var(--text);
  cursor: pointer;

  &:disabled {
    opacity: 0.55;
    cursor: not-allowed;
  }
}

.muted {
  color: var(--text-muted);
}

.small {
  font-size: 0.78rem;
}
</style>
