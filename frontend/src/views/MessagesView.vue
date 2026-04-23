<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import AppModal from '../components/AppModal.vue';
import ChatMessageLine from '../components/ChatMessageLine.vue';
import { ChatHistoryEntry, DefaultService } from '../api/generated';
import type { ChatBadgeTag } from '../lib/chatBadges';
import { effectiveChatterIsSus, effectiveSuspicionTitle } from '../lib/suspicionOverlay';
import { notifyApiError } from '../lib/notifyApiError';
import { useLiveSocketStore } from '../stores/liveSocket';
import { Button, LoadMoreRow, PageHeader, TextInput } from '../components/core';

defineOptions({ name: 'MessagesView' });

const liveSocket = useLiveSocketStore();

const loading = ref(false);
const loadingMore = ref(false);
const filtersApplying = ref(false);
const messages = ref<ChatHistoryEntry[]>([]);
const filtersOpen = ref(false);
const totalCount = ref<number | null>(null);

const applied = ref({
  username: '',
  text: '',
  channel: '',
  createdFrom: '',
  createdTo: '',
});

const draft = ref({ ...applied.value });

const filterSummary = computed(() => {
  const parts: string[] = [];
  if (applied.value.username) {
    parts.push(`user: ${applied.value.username}`);
  }
  if (applied.value.text) {
    parts.push(`text: ${applied.value.text}`);
  }
  if (applied.value.channel) {
    parts.push(`#${applied.value.channel}`);
  }
  if (applied.value.createdFrom) {
    parts.push(`from ${applied.value.createdFrom}`);
  }
  if (applied.value.createdTo) {
    parts.push(`to ${applied.value.createdTo}`);
  }
  return parts.length ? parts.join(' · ') : 'None';
});

function toIsoFromLocal(dtLocal: string): string | undefined {
  if (!dtLocal.trim()) {
    return undefined;
  }
  const d = new Date(dtLocal);
  if (!Number.isFinite(d.getTime())) {
    return undefined;
  }
  return d.toISOString();
}

function buildQuery(appendCursor: boolean) {
  const q: Parameters<typeof DefaultService.listTwitchMessages>[0] = {
    limit: 80,
    username: applied.value.username.trim() || undefined,
    text: applied.value.text.trim() || undefined,
    channel: applied.value.channel.replace(/^#/, '').trim().toLowerCase() || undefined,
    createdFrom: toIsoFromLocal(applied.value.createdFrom),
    createdTo: toIsoFromLocal(applied.value.createdTo),
  };

  if (appendCursor && messages.value.length) {
    const last = messages.value[messages.value.length - 1];
    q.cursorCreatedAt = last.created_at;
    q.cursorId = last.id;
  }

  return q;
}

function countQuery(): Parameters<typeof DefaultService.countTwitchMessages>[0] {
  return {
    username: applied.value.username.trim() || undefined,
    text: applied.value.text.trim() || undefined,
    channel: applied.value.channel.replace(/^#/, '').trim().toLowerCase() || undefined,
    createdFrom: toIsoFromLocal(applied.value.createdFrom),
    createdTo: toIsoFromLocal(applied.value.createdTo),
  };
}

async function fetchFirst(): Promise<void> {
  loading.value = true;
  try {
    const [list, cnt] = await Promise.all([
      DefaultService.listTwitchMessages(buildQuery(false)),
      DefaultService.countTwitchMessages(countQuery()),
    ]);
    messages.value = list;
    totalCount.value = cnt.total;
  } catch (e) {
    messages.value = [];
    totalCount.value = null;
    notifyApiError(e, {
      id: 'messages-load',
      title: 'Messages',
      fallbackMessage: 'Could not load messages.',
    });
  } finally {
    loading.value = false;
  }
}

async function refreshCount(): Promise<void> {
  try {
    const cnt = await DefaultService.countTwitchMessages(countQuery());
    totalCount.value = cnt.total;
  } catch {
    totalCount.value = null;
  }
}

async function fetchMore(): Promise<void> {
  if (!messages.value.length || loadingMore.value) {
    return;
  }
  loadingMore.value = true;
  try {
    const next = await DefaultService.listTwitchMessages(buildQuery(true));
    const seen = new Set(messages.value.map((m) => m.id));
    for (const m of next) {
      if (!seen.has(m.id)) {
        messages.value.push(m);
        seen.add(m.id);
      }
    }
    void refreshCount();
  } catch (e) {
    notifyApiError(e, {
      id: 'messages-more',
      title: 'Messages',
      fallbackMessage: 'Could not load more.',
    });
  } finally {
    loadingMore.value = false;
  }
}

function openFilters(): void {
  draft.value = { ...applied.value };
  filtersOpen.value = true;
}

async function applyFilters(): Promise<void> {
  if (filtersApplying.value) {
    return;
  }
  filtersApplying.value = true;
  try {
    applied.value = { ...draft.value };
    await fetchFirst();
    filtersOpen.value = false;
  } finally {
    filtersApplying.value = false;
  }
}

function clearFilters(): void {
  const empty = { username: '', text: '', channel: '', createdFrom: '', createdTo: '' };
  draft.value = { ...empty };
  applied.value = { ...empty };
  filtersOpen.value = false;
  void fetchFirst();
}

onMounted(() => {
  void fetchFirst();
});

function rowBadges(m: ChatHistoryEntry): ChatBadgeTag[] {
  return [...(m.badge_tags ?? [])] as ChatBadgeTag[];
}

function rowChatterIsSus(m: ChatHistoryEntry): boolean {
  return effectiveChatterIsSus(m.chatter_user_id ?? undefined, m.chatter_is_sus, liveSocket.suspicionByTwitchId);
}

function rowChatterSusTitle(m: ChatHistoryEntry): string {
  const eff = rowChatterIsSus(m);
  return effectiveSuspicionTitle(m.chatter_user_id ?? undefined, eff, liveSocket.suspicionByTwitchId) ?? '';
}
</script>

<template>
  <div class="page messages-page">
    <PageHeader title="Messages" :total-count="totalCount" layout="inline">
      <template #trailing>
        <div class="toolbar">
          <button type="button" class="btn-filter" @click="openFilters">Filters</button>
          <span class="filter-hint" :title="filterSummary">Filters: {{ filterSummary }}</span>
        </div>
      </template>
    </PageHeader>

    <p v-if="loading" class="muted muted--body">Loading…</p>
    <ul v-else class="lines">
      <ChatMessageLine
        v-for="m in messages"
        :key="m.id"
        :user="m.user"
        :message="m.message"
        :keyword="m.keyword_match"
        :from-sent="m.source === ChatHistoryEntry.source.SENT"
        :badge-tags="rowBadges(m)"
        :show-timestamp="true"
        :created-at="m.created_at"
        :chatter-user-id="m.chatter_user_id ?? undefined"
        :user-marked="m.chatter_marked"
        :user-is-sus="rowChatterIsSus(m)"
        :suspicious-title="rowChatterSusTitle(m)"
        :first-message="m.first_message"
        show-channel
        :channel-login="m.channel"
      />
    </ul>

    <LoadMoreRow v-if="!loading && messages.length" :loading="loadingMore" @click="fetchMore" />

    <AppModal :open="filtersOpen" title="Message filters" @close="filtersOpen = false">
      <template #default>
        <div class="fields">
          <TextInput v-model="draft.username" label="Username" autocomplete="off" density="compact" />
          <TextInput v-model="draft.text" label="Text contains" autocomplete="off" density="compact" />
          <TextInput
            v-model="draft.channel"
            label="Channel"
            placeholder="channel login"
            autocomplete="off"
            density="compact"
          />
          <TextInput v-model="draft.createdFrom" label="From" type="datetime-local" density="compact" />
          <TextInput v-model="draft.createdTo" label="To" type="datetime-local" density="compact" />
        </div>
      </template>
      <template #footer>
        <Button native-type="button" variant="ghost" size="small" @click="clearFilters">Clear</Button>
        <Button native-type="button" size="small" :loading="filtersApplying" @click="applyFilters">
          {{ filtersApplying ? 'Applying…' : 'Apply' }}
        </Button>
      </template>
    </AppModal>
  </div>
</template>

<style scoped lang="scss">
.page {
  padding: 0.75rem;
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
}

.btn-filter {
  padding: 0.4rem 0.75rem;
  border-radius: 0.25rem;
  border: 1px solid var(--border);
  background: var(--bg-elevated);
  color: var(--text);
  font-size: 0.85rem;
  cursor: pointer;

  &:hover {
    background: var(--bg-hover);
  }
}

.filter-hint {
  font-size: 0.78rem;
  color: var(--text-muted);
  max-width: min(40rem, 100%);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.lines {
  list-style: none;
  margin: 0;
  padding: 0;
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  font-size: 0.82rem;
  line-height: 1.35;
  border: 1px solid var(--border);
  border-radius: 0.35rem;
  background: var(--bg-elevated);
}

.fields {
  display: flex;
  flex-direction: column;
  gap: 0.65rem;
}
</style>
