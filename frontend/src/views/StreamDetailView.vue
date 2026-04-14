<script setup lang="ts">
import { useDebounceFn } from '@vueuse/core';
import { computed, onMounted, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import ChatMessageLine from '../components/ChatMessageLine.vue';
import TwitchUserLink from '../components/TwitchUserLink.vue';
import { ApiError, ChatHistoryEntry, DefaultService } from '../api/generated';
import type { RecordedStream } from '../api/generated';
import type { StreamLeaderboardEntry } from '../api/generated';
import { StreamLeaderboardSort } from '../api/generated/models/StreamLeaderboardSort';
import type { UserActivityEvent } from '../api/generated/models/UserActivityEvent';
import { formatDateTime } from '../lib/dateTime';
import { effectiveChatterIsSus, effectiveSuspicionTitle } from '../lib/suspicionOverlay';
import { notify } from '../lib/notify';
import { useLiveSocketStore } from '../stores/liveSocket';

defineOptions({ name: 'StreamDetailView' });

const liveSocket = useLiveSocketStore();

const route = useRoute();

const streamId = computed(() => Number.parseInt(String(route.params.id), 10));

const meta = ref<RecordedStream | null>(null);
const loadingMeta = ref(false);

type StreamTab = 'leaderboard' | 'messages' | 'activity';
const tab = ref<StreamTab>('leaderboard');

const leaderboard = ref<StreamLeaderboardEntry[]>([]);
const loadingLb = ref(false);
const lbSort = ref<StreamLeaderboardSort>(StreamLeaderboardSort.PRESENCE_DESC);
const lbFilter = ref('');

const messages = ref<ChatHistoryEntry[]>([]);
const loadingMsg = ref(false);
const loadingMsgMore = ref(false);
const msgCursorAt = ref<string | undefined>();
const msgCursorId = ref<number | undefined>();
const msgHasMore = ref(true);

const activity = ref<UserActivityEvent[]>([]);
const loadingAct = ref(false);
const loadingActMore = ref(false);
const actCursorAt = ref<string | undefined>();
const actCursorId = ref<number | undefined>();
const actHasMore = ref(true);

function notifyErr(e: unknown, id: string, title: string): void {
  const msg =
    e instanceof ApiError && e.body && typeof e.body.message === 'string'
      ? e.body.message
      : 'Request failed.';
  notify({ id, type: 'error', title, description: msg });
}

async function loadMeta(): Promise<void> {
  if (!Number.isFinite(streamId.value)) {
    meta.value = null;
    return;
  }
  loadingMeta.value = true;
  try {
    meta.value = await DefaultService.getRecordedStream({ streamId: streamId.value });
  } catch (e) {
    meta.value = null;
    notifyErr(e, 'stream-meta', 'Stream');
  } finally {
    loadingMeta.value = false;
  }
}

async function loadLeaderboard(): Promise<void> {
  if (!Number.isFinite(streamId.value)) {
    return;
  }
  loadingLb.value = true;
  try {
    leaderboard.value = await DefaultService.getRecordedStreamLeaderboard({
      streamId: streamId.value,
      sort: lbSort.value,
      q: lbFilter.value.trim() || undefined,
    });
  } catch (e) {
    leaderboard.value = [];
    notifyErr(e, 'stream-lb', 'Leaderboard');
  } finally {
    loadingLb.value = false;
  }
}

async function loadMessages(first: boolean): Promise<void> {
  if (!Number.isFinite(streamId.value)) {
    return;
  }
  if (first) {
    loadingMsg.value = true;
    msgCursorAt.value = undefined;
    msgCursorId.value = undefined;
    msgHasMore.value = true;
    messages.value = [];
  }
  try {
    const list = await DefaultService.listRecordedStreamMessages({
      streamId: streamId.value,
      limit: 80,
      cursorCreatedAt: first ? undefined : msgCursorAt.value,
      cursorId: first ? undefined : msgCursorId.value,
    });
    if (first) {
      messages.value = list;
    } else {
      messages.value = messages.value.concat(list);
    }
    if (list.length < 80) {
      msgHasMore.value = false;
    } else {
      const last = list[list.length - 1];
      msgCursorAt.value = last.created_at;
      msgCursorId.value = last.id;
    }
  } catch (e) {
    if (first) {
      messages.value = [];
    }
    notifyErr(e, 'stream-msg', 'Messages');
  } finally {
    loadingMsg.value = false;
    loadingMsgMore.value = false;
  }
}

async function loadActivity(first: boolean): Promise<void> {
  if (!Number.isFinite(streamId.value)) {
    return;
  }
  if (first) {
    loadingAct.value = true;
    actCursorAt.value = undefined;
    actCursorId.value = undefined;
    actHasMore.value = true;
    activity.value = [];
  }
  try {
    const list = await DefaultService.listRecordedStreamActivity({
      streamId: streamId.value,
      limit: 80,
      cursorCreatedAt: first ? undefined : actCursorAt.value,
      cursorId: first ? undefined : actCursorId.value,
    });
    if (first) {
      activity.value = list;
    } else {
      activity.value = activity.value.concat(list);
    }
    if (list.length < 80) {
      actHasMore.value = false;
    } else {
      const last = list[list.length - 1];
      actCursorAt.value = last.created_at;
      actCursorId.value = last.id;
    }
  } catch (e) {
    if (first) {
      activity.value = [];
    }
    notifyErr(e, 'stream-act', 'Activity');
  } finally {
    loadingAct.value = false;
    loadingActMore.value = false;
  }
}

function formatClock(sec: number): string {
  const s = Math.max(0, Math.floor(sec));
  const h = Math.floor(s / 3600);
  const m = Math.floor((s % 3600) / 60);
  const r = s % 60;
  return [h, m, r].map((n) => String(n).padStart(2, '0')).join(':');
}

function formatWhen(iso: string): string {
  return formatDateTime(iso);
}

function activityLabel(e: UserActivityEvent): string {
  const ch = e.channel ? `#${e.channel}` : '';
  switch (e.event_type) {
    case 'chat_online':
      return `Joined ${ch || 'chat'}`;
    case 'chat_offline':
      return `Left ${ch || 'chat'}`;
    default:
      return e.event_type;
  }
}

function rowChatterIsSus(m: ChatHistoryEntry): boolean {
  return effectiveChatterIsSus(m.chatter_user_id ?? undefined, m.chatter_is_sus, liveSocket.suspicionByTwitchId);
}

function rowChatterSusTitle(m: ChatHistoryEntry): string {
  const eff = rowChatterIsSus(m);
  return effectiveSuspicionTitle(m.chatter_user_id ?? undefined, eff, liveSocket.suspicionByTwitchId) ?? '';
}

async function refreshTab(): Promise<void> {
  if (tab.value === 'leaderboard') {
    await loadLeaderboard();
  } else if (tab.value === 'messages') {
    await loadMessages(true);
  } else {
    await loadActivity(true);
  }
}

watch(streamId, () => {
  void loadMeta();
  void refreshTab();
});

watch(tab, (t) => {
  if (t === 'leaderboard' && !loadingLb.value && !leaderboard.value.length) {
    void loadLeaderboard();
  }
  if (t === 'messages' && !messages.value.length && !loadingMsg.value) {
    void loadMessages(true);
  }
  if (t === 'activity' && !activity.value.length && !loadingAct.value) {
    void loadActivity(true);
  }
});

watch(lbSort, () => {
  if (tab.value === 'leaderboard') {
    void loadLeaderboard();
  }
});

const debouncedLbFilter = useDebounceFn(() => {
  if (tab.value === 'leaderboard') {
    void loadLeaderboard();
  }
}, 320);

watch(lbFilter, () => {
  debouncedLbFilter();
});

onMounted(() => {
  void loadMeta();
  void loadLeaderboard();
});

function loadMoreMessages(): void {
  if (loadingMsgMore.value || !msgHasMore.value) {
    return;
  }
  loadingMsgMore.value = true;
  void loadMessages(false);
}

function loadMoreActivity(): void {
  if (loadingActMore.value || !actHasMore.value) {
    return;
  }
  loadingActMore.value = true;
  void loadActivity(false);
}
</script>

<template>
  <div class="stream-detail">
    <p v-if="loadingMeta" class="muted">Loading stream…</p>
    <template v-else-if="meta">
      <header class="stream-head">
        <h1>
          <span class="live-pill" :class="{ 'live-pill--off': meta.ended_at }">{{
            meta.ended_at ? 'Past' : 'Live'
          }}</span>
          #{{ meta.channel_login }}
        </h1>
        <p class="meta-line">
          <span>{{ meta.title?.trim() || '—' }}</span>
          <span class="muted">·</span>
          <span class="muted">{{ meta.game_name?.trim() || '—' }}</span>
        </p>
        <p class="muted small">
          Started {{ formatWhen(meta.started_at) }}
          <template v-if="meta.ended_at"> · Ended {{ formatWhen(meta.ended_at) }}</template>
        </p>
      </header>

      <nav class="tabs">
        <button type="button" :class="{ active: tab === 'leaderboard' }" @click="tab = 'leaderboard'">
          Leaderboard
        </button>
        <button type="button" :class="{ active: tab === 'messages' }" @click="tab = 'messages'">Messages</button>
        <button type="button" :class="{ active: tab === 'activity' }" @click="tab = 'activity'">Activity</button>
      </nav>

      <section v-show="tab === 'leaderboard'" class="panel">
        <div class="toolbar">
          <label class="grow">
            <span class="sr-only">Filter</span>
            <input v-model="lbFilter" type="search" placeholder="Filter login…" />
          </label>
          <label>
            <span class="sr-only">Sort</span>
            <select v-model="lbSort">
              <option :value="StreamLeaderboardSort.PRESENCE_DESC">Presence (high)</option>
              <option :value="StreamLeaderboardSort.PRESENCE_ASC">Presence (low)</option>
              <option :value="StreamLeaderboardSort.MESSAGES_DESC">Messages (high)</option>
              <option :value="StreamLeaderboardSort.MESSAGES_ASC">Messages (low)</option>
              <option :value="StreamLeaderboardSort.LOGIN_AZ">Name A–Z</option>
              <option :value="StreamLeaderboardSort.LOGIN_ZA">Name Z–A</option>
              <option :value="StreamLeaderboardSort.ACCOUNT_NEW">Newest accounts</option>
              <option :value="StreamLeaderboardSort.ACCOUNT_OLD">Oldest accounts</option>
            </select>
          </label>
        </div>
        <p v-if="loadingLb" class="muted">Loading…</p>
        <table v-else-if="leaderboard.length" class="lb-table">
          <thead>
            <tr>
              <th>User</th>
              <th>Presence</th>
              <th>Messages</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in leaderboard" :key="row.user_twitch_id">
              <td>
                <TwitchUserLink
                  :login="row.login"
                  :user-twitch-id="row.user_twitch_id"
                  :highlight-channel="meta.channel_login"
                  variant="chat"
                />
              </td>
              <td>{{ formatClock(row.presence_seconds) }}</td>
              <td>{{ row.message_count }}</td>
            </tr>
          </tbody>
        </table>
        <p v-else class="muted">No data for this stream.</p>
      </section>

      <section v-show="tab === 'messages'" class="panel">
        <p v-if="loadingMsg" class="muted">Loading messages…</p>
        <ul v-else-if="messages.length" class="msg-list">
          <li v-for="m in messages" :key="m.id">
            <ChatMessageLine
              :user="m.user"
              :message="m.message"
              :keyword="m.keyword_match"
              :user-marked="m.chatter_marked"
              :user-is-sus="rowChatterIsSus(m)"
              :suspicious-title="rowChatterSusTitle(m)"
              :from-sent="m.source === 'sent'"
              :badge-tags="m.badge_tags"
              :show-timestamp="true"
              :created-at="m.created_at"
              :chatter-user-id="m.chatter_user_id ?? null"
              :highlight-channel="meta.channel_login"
              :first-message="m.first_message"
            />
          </li>
        </ul>
        <p v-else class="muted">No messages recorded for this stream.</p>
        <div v-if="messages.length && msgHasMore" class="more-row">
          <button type="button" class="btn-more" :disabled="loadingMsgMore" @click="loadMoreMessages">
            {{ loadingMsgMore ? 'Loading…' : 'Load more' }}
          </button>
        </div>
      </section>

      <section v-show="tab === 'activity'" class="panel">
        <p v-if="loadingAct" class="muted">Loading activity…</p>
        <ul v-else-if="activity.length" class="activity-list">
          <li v-for="e in activity" :key="e.id">
            <time class="act-time" :datetime="e.created_at">{{ formatWhen(e.created_at) }}</time>
            <span class="act-user">{{ e.username }}</span>
            <span class="act-body">{{ activityLabel(e) }}</span>
          </li>
        </ul>
        <p v-else class="muted">No activity in this window.</p>
        <div v-if="activity.length && actHasMore" class="more-row">
          <button type="button" class="btn-more" :disabled="loadingActMore" @click="loadMoreActivity">
            {{ loadingActMore ? 'Loading…' : 'Load more' }}
          </button>
        </div>
      </section>
    </template>
    <p v-else class="muted">Stream not found.</p>
  </div>
</template>

<style scoped lang="scss">
.stream-detail {
  padding: 0.75rem 1rem;
  max-width: 56rem;
  margin: 0 auto;
  width: 100%;
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: 0.65rem;
}

.stream-head h1 {
  margin: 0 0 0.25rem;
  font-size: 1.35rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.live-pill {
  font-size: 0.75rem;
  padding: 0.1rem 0.45rem;
  border-radius: 0.25rem;
  background: rgba(145, 70, 255, 0.2);
  color: var(--accent-bright);
  border: 1px solid var(--accent);

  &--off {
    background: var(--bg-elevated);
    color: var(--text-muted);
    border-color: var(--border);
  }
}

.meta-line {
  margin: 0;
}

.small {
  font-size: 0.85rem;
}

.tabs {
  display: flex;
  gap: 0.35rem;
  border-bottom: 1px solid var(--border);
  padding-bottom: 0.25rem;

  button {
    padding: 0.35rem 0.65rem;
    border: none;
    background: transparent;
    color: var(--text-muted);
    cursor: pointer;
    border-radius: 0.25rem 0.25rem 0 0;

    &.active {
      color: var(--text);
      background: var(--bg-elevated);
      border: 1px solid var(--border);
      border-bottom-color: transparent;
    }
  }
}

.panel {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  align-items: center;

  .grow {
    flex: 1;
    min-width: 8rem;
  }

  input,
  select {
    width: 100%;
    padding: 0.35rem 0.5rem;
    border-radius: 0.25rem;
    border: 1px solid var(--border);
    background: var(--bg-base);
    color: var(--text);
  }
}

.lb-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.9rem;

  th,
  td {
    text-align: left;
    padding: 0.35rem 0.5rem;
    border-bottom: 1px solid var(--border);
  }
}

.msg-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}

.activity-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  font-size: 0.9rem;

  li {
    display: grid;
    grid-template-columns: 10rem 10rem 1fr;
    gap: 0.5rem;
    align-items: baseline;
  }
}

.act-time {
  color: var(--text-muted);
  font-size: 0.8rem;
}

.act-user {
  font-weight: 600;
}

.more-row {
  margin-top: 0.35rem;
}

.btn-more {
  padding: 0.35rem 0.75rem;
  border-radius: 0.25rem;
  border: 1px dashed var(--border);
  background: transparent;
  color: var(--text-muted);
  cursor: pointer;
}
</style>
