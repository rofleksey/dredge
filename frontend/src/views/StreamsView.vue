<script setup lang="ts">
import { watchDebounced } from '@vueuse/core';
import { onMounted, ref } from 'vue';
import { RouterLink } from 'vue-router';
import { ApiError, DefaultService } from '../api/generated';
import type { RecordedStream } from '../api/generated';
import { formatDateTime } from '../lib/dateTime';
import { notify } from '../lib/notify';

defineOptions({ name: 'StreamsView' });

const streams = ref<RecordedStream[]>([]);
const loading = ref(false);
const loadingMore = ref(false);
const channelFilter = ref('');

const cursorStartedAt = ref<string | undefined>();
const cursorId = ref<number | undefined>();
const hasMore = ref(true);

async function loadFirst(): Promise<void> {
  if (loading.value) {
    return;
  }
  loading.value = true;
  cursorStartedAt.value = undefined;
  cursorId.value = undefined;
  hasMore.value = true;
  try {
    await loadPage(false);
  } finally {
    loading.value = false;
  }
}

async function loadPage(append: boolean): Promise<void> {
  const list = await DefaultService.listRecordedStreams({
    channelLogin: channelFilter.value.trim() || undefined,
    limit: 50,
    cursorStartedAt: append ? cursorStartedAt.value : undefined,
    cursorId: append ? cursorId.value : undefined,
  });
  if (append) {
    streams.value = streams.value.concat(list);
  } else {
    streams.value = list;
  }
  if (list.length < 50) {
    hasMore.value = false;
  } else {
    const last = list[list.length - 1];
    cursorStartedAt.value = last.started_at;
    cursorId.value = last.id;
    hasMore.value = true;
  }
}

async function loadMore(): Promise<void> {
  if (loadingMore.value || !hasMore.value || !streams.value.length) {
    return;
  }
  loadingMore.value = true;
  try {
    await loadPage(true);
  } catch (e) {
    notifyFromErr(e, 'streams-more');
  } finally {
    loadingMore.value = false;
  }
}

function notifyFromErr(e: unknown, id: string): void {
  const msg =
    e instanceof ApiError && e.body && typeof e.body.message === 'string'
      ? e.body.message
      : 'Request failed.';
  notify({ id, type: 'error', title: 'Streams', description: msg });
}

async function applyFilterDebounced(): Promise<void> {
  try {
    await loadFirst();
  } catch (e) {
    streams.value = [];
    notifyFromErr(e, 'streams-load');
  }
}

watchDebounced(
  channelFilter,
  () => {
    void applyFilterDebounced();
  },
  { debounce: 350 },
);

onMounted(() => {
  void loadFirst().catch((e) => {
    streams.value = [];
    notifyFromErr(e, 'streams-load');
  });
});

function formatWhen(iso: string): string {
  return formatDateTime(iso);
}

function statusLabel(s: RecordedStream): string {
  return s.ended_at ? 'Ended' : 'Live';
}
</script>

<template>
  <div class="streams-page">
    <header class="streams-head">
      <h1>Streams</h1>
      <p class="muted">Recorded broadcasts for monitored channels</p>
    </header>

    <div class="streams-filter">
      <label class="field">
        <span class="label">Channel</span>
        <input
          v-model="channelFilter"
          type="search"
          name="channel"
          placeholder="Filter by login…"
          autocomplete="off"
        />
      </label>
    </div>

    <p v-if="loading" class="muted">Loading…</p>
    <table v-else-if="streams.length" class="streams-table">
      <thead>
        <tr>
          <th>Title</th>
          <th>Channel</th>
          <th>Category</th>
          <th>Started</th>
          <th>Status</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="s in streams" :key="s.id">
          <td>
            <RouterLink class="link title" :to="{ name: 'stream', params: { id: String(s.id) } }">
              {{ s.title?.trim() || '—' }}
            </RouterLink>
          </td>
          <td>
            <RouterLink class="link" :to="{ name: 'stream', params: { id: String(s.id) } }">
              #{{ s.channel_login }}
            </RouterLink>
          </td>
          <td class="muted">{{ s.game_name?.trim() || '—' }}</td>
          <td class="muted">{{ formatWhen(s.started_at) }}</td>
          <td>
            <span :class="['status', { 'status--live': !s.ended_at }]">{{ statusLabel(s) }}</span>
          </td>
        </tr>
      </tbody>
    </table>
    <p v-else class="muted">No streams recorded yet.</p>

    <div v-if="streams.length && hasMore" class="more-row">
      <button type="button" class="btn-more" :disabled="loadingMore" @click="loadMore">
        {{ loadingMore ? 'Loading…' : 'Load more' }}
      </button>
    </div>
  </div>
</template>

<style scoped lang="scss">
.streams-page {
  padding: 0.75rem 1rem;
  max-width: 72rem;
  margin: 0 auto;
  width: 100%;
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.streams-head h1 {
  margin: 0 0 0.25rem;
  font-size: 1.35rem;
}

.streams-filter {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  align-items: flex-end;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
  min-width: 12rem;

  .label {
    font-size: 0.8rem;
    color: var(--text-muted);
  }

  input {
    padding: 0.35rem 0.5rem;
    border-radius: 0.25rem;
    border: 1px solid var(--border);
    background: var(--bg-base);
    color: var(--text);
  }
}

.streams-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.9rem;

  th,
  td {
    text-align: left;
    padding: 0.35rem 0.5rem;
    border-bottom: 1px solid var(--border);
  }

  th {
    color: var(--text-muted);
    font-weight: 600;
  }
}

.link {
  color: var(--accent-bright);
  text-decoration: none;

  &:hover {
    text-decoration: underline;
  }

  &.title {
    color: var(--text);
  }
}

.status {
  font-size: 0.8rem;
  padding: 0.1rem 0.4rem;
  border-radius: 0.2rem;
  background: var(--bg-elevated);
  border: 1px solid var(--border);

  &--live {
    border-color: var(--accent);
    color: var(--accent-bright);
  }
}

.more-row {
  margin-top: 0.5rem;
}

.btn-more {
  padding: 0.35rem 0.75rem;
  border-radius: 0.25rem;
  border: 1px dashed var(--border);
  background: transparent;
  color: var(--text-muted);
  cursor: pointer;

  &:hover:not(:disabled) {
    border-color: var(--accent);
    color: var(--accent);
  }
}
</style>
