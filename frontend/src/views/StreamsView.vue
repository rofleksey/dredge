<script setup lang="ts">
import { watchDebounced } from '@vueuse/core';
import { onMounted, ref } from 'vue';
import { RouterLink } from 'vue-router';
import { DefaultService } from '../api/generated';
import type { RecordedStream } from '../api/generated';
import { formatDateTime } from '../lib/dateTime';
import { LoadMoreRow, PageHeader, TextInput } from '../components/core';
import { notifyApiError } from '../lib/notifyApiError';

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
    notifyApiError(e, { id: 'streams-more', title: 'Streams', fallbackMessage: 'Request failed.' });
  } finally {
    loadingMore.value = false;
  }
}

async function applyFilterDebounced(): Promise<void> {
  try {
    await loadFirst();
  } catch (e) {
    streams.value = [];
    notifyApiError(e, { id: 'streams-load', title: 'Streams', fallbackMessage: 'Request failed.' });
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
    notifyApiError(e, { id: 'streams-load', title: 'Streams', fallbackMessage: 'Request failed.' });
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
    <PageHeader
      title="Streams"
      subtitle="Recorded broadcasts for monitored channels"
      layout="stacked"
      size="large"
    />

    <div class="streams-filter">
      <TextInput
        v-model="channelFilter"
        label="Channel"
        name="channel"
        type="search"
        autocomplete="off"
        placeholder="Filter by login…"
        density="compact"
        surface="base"
      />
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

    <LoadMoreRow
      v-if="streams.length && hasMore"
      variant="ghost"
      :loading="loadingMore"
      @click="loadMore"
    />
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

.streams-filter {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  align-items: flex-end;
}

.streams-filter :deep(.text-input-root) {
  min-width: 12rem;
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

</style>
