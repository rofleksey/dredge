<script setup lang="ts">
import AppModal from '../../components/AppModal.vue';
import TwitchUserLink from '../../components/TwitchUserLink.vue';
import type { ChannelLive, ChannelChatterEntry } from '../../api/generated';
import { formatDateTime } from '../../lib/dateTime';
import type { ViewerSortMode } from './types';

defineProps<{
  open: boolean;
  channelLive: ChannelLive | null;
  formatSessionClock: (iso?: string | null) => string;
  formatViewerDisplay: (live: Pick<ChannelLive, 'viewer_count' | 'channel_chatter_count'>) => string;
  formatPresentElapsed: (iso: string) => string;
  formatAccountDate: (iso?: string | null) => string;
  viewerChatters: ChannelChatterEntry[];
  loadingViewerChatters: boolean;
  displayedViewerChatters: ChannelChatterEntry[];
  selectedChannel: string;
  normCh: (c: string) => string;
}>();

const viewerFilterQuery = defineModel<string>('viewerFilterQuery', { required: true });
const viewerSort = defineModel<ViewerSortMode>('viewerSort', { required: true });

const emit = defineEmits<{
  close: [];
}>();
</script>

<template>
  <AppModal :open="open" wide title="Stream details" @close="emit('close')">
    <dl v-if="channelLive" class="viewer-dl">
      <div>
        <dt>Title</dt>
        <dd>{{ channelLive.title ?? '—' }}</dd>
      </div>
      <div>
        <dt>Category</dt>
        <dd>{{ channelLive.game_name ?? '—' }}</dd>
      </div>
      <div>
        <dt>Viewers</dt>
        <dd>{{ formatViewerDisplay(channelLive) }}</dd>
      </div>
      <div>
        <dt>Live since</dt>
        <dd>{{ channelLive.started_at ? formatDateTime(channelLive.started_at) : '—' }}</dd>
      </div>
      <div>
        <dt>Session uptime</dt>
        <dd>{{ formatSessionClock(channelLive.started_at) }}</dd>
      </div>
    </dl>
    <div v-if="channelLive?.is_live" class="viewer-chatters">
      <h3 class="viewer-chatters-title">Chatters in channel</h3>
      <div v-if="viewerChatters.length" class="viewer-chatter-toolbar">
        <label class="viewer-filter">
          <span class="sr-only">Filter by name</span>
          <input
            v-model="viewerFilterQuery"
            type="search"
            name="viewer_filter"
            autocomplete="off"
            autocorrect="off"
            spellcheck="false"
            placeholder="Filter…"
          />
        </label>
        <label class="viewer-sort">
          <span class="sr-only">Sort</span>
          <select v-model="viewerSort" name="viewer_sort">
            <option value="present_new">New in chat first</option>
            <option value="present_old">Longest in chat first</option>
            <option value="login_az">Name A–Z</option>
            <option value="login_za">Name Z–A</option>
            <option value="account_new">Newest Twitch accounts</option>
            <option value="account_old">Oldest Twitch accounts</option>
            <option value="message_high">Most messages</option>
            <option value="message_low">Fewest messages</option>
          </select>
        </label>
      </div>
      <p v-if="loadingViewerChatters" class="muted muted--compact tiny">Loading…</p>
      <ul v-else-if="displayedViewerChatters.length" class="viewer-chatter-list">
        <li v-for="c in displayedViewerChatters" :key="c.user_twitch_id" class="viewer-chatter-row">
          <TwitchUserLink
            :login="c.login"
            :user-twitch-id="c.user_twitch_id"
            :highlight-channel="normCh(selectedChannel)"
            variant="chat"
          />
          <span class="viewer-chatter-meta">Present {{ formatPresentElapsed(c.present_since) }}</span>
          <span v-if="c.message_count != null && c.message_count !== undefined" class="viewer-chatter-meta"
            >Messages {{ c.message_count }}</span
          >
          <span v-if="c.account_created_at" class="viewer-chatter-meta"
            >Account {{ formatAccountDate(c.account_created_at) }}</span
          >
        </li>
      </ul>
      <p v-else-if="viewerChatters.length && !displayedViewerChatters.length" class="muted muted--compact tiny">
        No names match the filter.
      </p>
      <p v-else class="muted muted--compact tiny">No names loaded.</p>
    </div>
  </AppModal>
</template>

<style scoped lang="scss">
.tiny {
  font-size: 0.72rem;
}

.viewer-chatters {
  margin-top: 0.75rem;
  padding-top: 0.65rem;
  border-top: 1px solid var(--border);
}

.viewer-chatters-title {
  margin: 0 0 0.4rem;
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-muted);
}

.viewer-chatter-toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  align-items: center;
  margin-bottom: 0.5rem;

  input,
  select {
    padding: 0.3rem 0.45rem;
    border-radius: 0.2rem;
    border: 1px solid var(--border);
    background: var(--bg-base);
    color: var(--text);
    font-size: 0.78rem;
    min-width: 0;
  }

  .viewer-filter {
    flex: 1 1 10rem;
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: 0.15rem;
  }

  .viewer-sort {
    flex: 0 1 auto;
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: 0.15rem;
  }
}

.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

.viewer-chatter-list {
  list-style: none;
  margin: 0;
  padding: 0;
  height: min(36vh, 32rem);
  min-height: min(36vh, 32rem);
  max-height: min(36vh, 32rem);
  overflow: auto;
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
  font-size: 0.82rem;
}

.viewer-chatter-row {
  display: flex;
  flex-direction: column;
  gap: 0.12rem;
  padding: 0.35rem 0.25rem;
  border-radius: 0.25rem;
  border: 1px solid var(--border);
  background: var(--bg-base);
}

.viewer-chatter-meta {
  font-size: 0.72rem;
  color: var(--text-muted);
  font-variant-numeric: tabular-nums;
}

.viewer-dl {
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.55rem;

  div {
    display: grid;
    grid-template-columns: 7rem 1fr;
    gap: 0.35rem;
    align-items: baseline;
  }

  dt {
    margin: 0;
    font-size: 0.75rem;
    color: var(--text-muted);
  }

  dd {
    margin: 0;
    font-size: 0.85rem;
    word-break: break-word;
  }
}
</style>
